// EFT logs watcher — a port of LogsWatcher.cs / the ps1 script.
//
// Every tick (1 sec):
//   - the newest session folder in <game>\Logs is picked (new game session → new folder);
//   - new COMPLETE lines of application, backend and push-notifications logs are read
//     (an unfinished last line waits until the next tick);
//   - when a file is first seen, the position is set to its end (old events are not replayed).
//
// FileSystemWatcher is unreliable on the logs (the game writes via buffered I/O),
// so size polling only — both the C# and the ps1 versions did the same.
package watcher

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ggdiam/TarkovPilot/app/internal/detect"
)

const (
	// location change markers (application_*.log): PVP and PVE log differently
	locationSubstring  = "application|TRACE-NetworkGameCreate profileStatus"
	locationSubstring2 = "application|scene preset"

	// quest notification marker (push-notifications_*.log)
	taskSubstring = "push-notifications|Got notification | ChatMessageReceived"
)

var (
	locationRe  = regexp.MustCompile(`(?i)location:\s*(\S+),`)
	locationRe2 = regexp.MustCompile(`(?i)path:maps/(\w+)\.bundle`)

	// a log line starting with a date — the end of the notification JSON block
	lineStartWithDateRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{1,2}:\d{1,2}:\d{1,2}\.\d{3}`)
)

type LogsWatcher struct {
	mu     sync.Mutex
	stopCh chan struct{}

	logsFolder       string
	currentFolder    string
	positions        map[string]int64
	activeProfileKey string
	activeGameMode   string

	OnMap   func(rawMap string)
	OnQuest func(questId, status, profileKey, gameMode string)
	OnLog   func(msg string)
}

func NewLogsWatcher() *LogsWatcher {
	return &LogsWatcher{positions: map[string]int64{}}
}

// Watching — polling is running and the logs folder exists
func (w *LogsWatcher) Watching() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopCh == nil || w.logsFolder == "" {
		return false
	}
	fi, err := os.Stat(w.logsFolder)
	return err == nil && fi.IsDir()
}

// Start begins polling the logs (stopping the previous poll)
func (w *LogsWatcher) Start(logsFolder string) {
	w.Stop()

	w.mu.Lock()
	w.logsFolder = logsFolder
	w.currentFolder = ""
	w.positions = map[string]int64{}
	w.activeProfileKey = ""
	w.activeGameMode = ""
	stopCh := make(chan struct{})
	w.stopCh = stopCh
	w.mu.Unlock()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				w.poll()
			}
		}
	}()
}

func (w *LogsWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopCh != nil {
		close(w.stopCh)
		w.stopCh = nil
	}
}

func (w *LogsWatcher) poll() {
	w.mu.Lock()
	logsFolder := w.logsFolder
	w.mu.Unlock()

	latest := detect.NewestSessionFolder(logsFolder)
	if latest == "" {
		return
	}

	w.mu.Lock()
	activeSessionChanged := false
	if w.currentFolder != latest {
		// the game was restarted — a new session folder
		w.currentFolder = latest
		w.positions = map[string]int64{}
		w.activeProfileKey = ""
		w.activeGameMode = ""
		activeSessionChanged = true
	}
	w.mu.Unlock()

	// The callback builds app state and calls back into LogsWatcher.
	// It must not run under w.mu: sync.Mutex is non-recursive, that would deadlock.
	if activeSessionChanged && w.OnLog != nil {
		w.OnLog("Active log session: " + filepath.Base(latest))
	}

	files := []string{}
	for _, pattern := range []string{"*application_*.log", "*backend_*.log", "*push-notifications_*.log"} {
		matches, _ := filepath.Glob(filepath.Join(latest, pattern))
		files = append(files, matches...)
	}

	for _, path := range files {
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}

		w.mu.Lock()
		pos, known := w.positions[path]
		w.mu.Unlock()

		// first time we see the file — start from the end (tail)
		if !known {
			base := strings.ToLower(filepath.Base(path))
			if strings.Contains(base, "application_") || strings.Contains(base, "backend_") {
				w.readCurrentProfileState(path)
			}
			w.mu.Lock()
			w.positions[path] = fi.Size()
			w.mu.Unlock()
			continue
		}

		lines, newPos := readNewLines(path, pos)
		w.mu.Lock()
		w.positions[path] = newPos
		w.mu.Unlock()

		if len(lines) > 0 {
			w.processLines(lines)
		}
	}
}

// readCurrentProfileState restores routing state when the app starts in the middle
// of an EFT session. It intentionally ignores maps and quest notifications.
func (w *LogsWatcher) readCurrentProfileState(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := newLogScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if match := profileIdRe.FindStringSubmatch(line); match != nil {
			w.mu.Lock()
			w.activeProfileKey = hashProfileId(match[1])
			w.mu.Unlock()
		}
		if match := gatewayModeRe.FindStringSubmatch(line); match != nil {
			w.mu.Lock()
			w.activeGameMode = normalizeGameMode(match[1])
			w.mu.Unlock()
		}
	}
}

// readNewLines reads new COMPLETE lines starting at pos; an unfinished tail
// stays unread until the next tick
func readNewLines(path string, pos int64) ([]string, int64) {
	f, err := os.Open(path)
	if err != nil {
		return nil, pos
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, pos
	}
	// file truncated/rotated — start over
	if pos > fi.Size() {
		pos = 0
	}

	if _, err := f.Seek(pos, io.SeekStart); err != nil {
		return nil, pos
	}

	b, err := io.ReadAll(f)
	if err != nil || len(b) == 0 {
		return nil, pos
	}

	content := string(b)
	lastNl := strings.LastIndexByte(content, '\n')
	if lastNl < 0 {
		// no complete line yet
		return nil, pos
	}

	complete := content[:lastNl+1]
	consumed := int64(len(complete))

	lines := strings.Split(strings.ReplaceAll(complete, "\r\n", "\n"), "\n")
	return lines, pos + consumed
}

// pushNotification — the part of the notification JSON block we care about
type pushNotification struct {
	Message struct {
		Type       any    `json:"type"`
		TemplateId string `json:"templateId"`
	} `json:"message"`
}

func (w *LogsWatcher) processLines(lines []string) {
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		if line == "" {
			continue
		}

		if match := profileIdRe.FindStringSubmatch(line); match != nil {
			w.mu.Lock()
			w.activeProfileKey = hashProfileId(match[1])
			w.mu.Unlock()
		}
		if match := gatewayModeRe.FindStringSubmatch(line); match != nil {
			w.mu.Lock()
			w.activeGameMode = normalizeGameMode(match[1])
			w.mu.Unlock()
		}

		if strings.Contains(line, locationSubstring) {
			if m := locationRe.FindStringSubmatch(line); m != nil {
				w.emitMap(strings.ToLower(m[1]))
			}
		} else if strings.Contains(line, locationSubstring2) {
			if m := locationRe2.FindStringSubmatch(line); m != nil {
				w.emitMap(strings.ToLower(m[1]))
			}
		} else if strings.Contains(line, taskSubstring) {
			// the notification JSON spans the following lines until a dated line
			var sb strings.Builder
			i++
			for i < len(lines) {
				jl := lines[i]
				if lineStartWithDateRe.MatchString(jl) {
					i-- // the dated line is handled by the outer loop
					break
				}
				sb.WriteString(jl)
				sb.WriteString("\n")
				i++
			}

			jsonStr := strings.TrimSpace(sb.String())
			if jsonStr == "" {
				continue
			}

			var rec pushNotification
			if err := json.Unmarshal([]byte(jsonStr), &rec); err != nil {
				continue // incomplete/broken block — skip
			}
			if rec.Message.Type == nil || rec.Message.TemplateId == "" {
				continue
			}

			// templateId looks like "6574e0dedc0d635f633a5805 successMessageText"
			questId := strings.SplitN(rec.Message.TemplateId, " ", 2)[0]
			if questId == "" {
				continue
			}

			status := statusToString(rec.Message.Type)
			if w.OnQuest != nil {
				w.mu.Lock()
				profileKey := w.activeProfileKey
				gameMode := w.activeGameMode
				w.mu.Unlock()
				w.OnQuest(questId, status, profileKey, gameMode)
			}
		}
	}
}

func (w *LogsWatcher) emitMap(rawMap string) {
	if w.OnMap != nil {
		w.OnMap(rawMap)
	}
}

// statusToString — message.type in the log can be a number or a string
func statusToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64)
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}
