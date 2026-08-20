package watcher

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const questCompleteStatus = "12"

var (
	logFolderYearRe = regexp.MustCompile(`^log_(\d{4})\.`)
	profileIdRe     = regexp.MustCompile(`(?i)(?:CompleteSelectedProfile\s+ProfileId:|profileStatus.*?Profileid:\s*)\s*([0-9a-f]{24})`)
	gatewayModeRe   = regexp.MustCompile(`(?i)https://gw-(pvp-season|pve|pvp)\.`)
	questIdRe       = regexp.MustCompile(`^[0-9a-f]{24}$`)
)

// QuestScanProfile is an anonymized EFT profile with completed quests found in logs.
type QuestScanProfile struct {
	ProfileKey   string    `json:"profileKey"`
	GameMode     string    `json:"gameMode"`
	QuestIds     []string  `json:"questIds"`
	FirstEventAt time.Time `json:"firstEventAt"`
	LastEventAt  time.Time `json:"lastEventAt"`
}

// QuestHistoryScan contains only anonymized data safe to send to the website.
type QuestHistoryScan struct {
	Profiles      []QuestScanProfile `json:"profiles"`
	SkippedEvents int                `json:"skippedEvents"`
}

type historyEventKind int

const (
	historyMode historyEventKind = iota
	historyProfile
	historyQuest
)

type historyEvent struct {
	at    time.Time
	kind  historyEventKind
	value string
}

type questCompletion struct {
	at         time.Time
	profileKey string
	gameMode   string
	questId    string
}

type profileModes map[string]map[string]struct{}

// ScanQuestHistory scans EFT sessions without exposing the raw EFT profile id.
func ScanQuestHistory(logsFolder string, since time.Time) (QuestHistoryScan, error) {
	entries, err := os.ReadDir(logsFolder)
	if err != nil {
		return QuestHistoryScan{}, err
	}

	var completions []questCompletion
	knownModes := profileModes{}
	for _, entry := range entries {
		if !entry.IsDir() || !sessionCanContainDate(entry.Name(), since) {
			continue
		}
		events, err := scanHistorySession(filepath.Join(logsFolder, entry.Name()), since)
		if err != nil {
			continue
		}
		completions = append(completions, correlateSessionEvents(events, knownModes)...)
	}

	return buildQuestHistoryScan(completions, knownModes), nil
}

func sessionCanContainDate(name string, since time.Time) bool {
	match := logFolderYearRe.FindStringSubmatch(name)
	if match == nil {
		return false
	}
	return match[1] >= since.Format("2006")
}

func scanHistorySession(folder string, since time.Time) ([]historyEvent, error) {
	var events []historyEvent
	foundFiles := false

	for _, spec := range []struct {
		pattern string
		parse   func(string, time.Time) ([]historyEvent, error)
	}{
		{"*application_*.log", scanStateLog},
		{"*backend_*.log", scanStateLog},
		{"*push-notifications_*.log", scanQuestLog},
	} {
		paths, _ := filepath.Glob(filepath.Join(folder, spec.pattern))
		for _, path := range paths {
			foundFiles = true
			fileEvents, err := spec.parse(path, since)
			if err == nil {
				events = append(events, fileEvents...)
			}
		}
	}

	if !foundFiles {
		return nil, errors.New("no supported log files")
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].at.Equal(events[j].at) {
			return events[i].kind < events[j].kind
		}
		return events[i].at.Before(events[j].at)
	})
	return events, nil
}

func scanStateLog(path string, since time.Time) ([]historyEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []historyEvent
	scanner := newLogScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		at, ok := parseLogTime(line)
		if !ok || at.Before(since) {
			continue
		}
		if match := gatewayModeRe.FindStringSubmatch(line); match != nil {
			events = append(events, historyEvent{at: at, kind: historyMode, value: normalizeGameMode(match[1])})
		}
		if match := profileIdRe.FindStringSubmatch(line); match != nil {
			events = append(events, historyEvent{at: at, kind: historyProfile, value: hashProfileId(match[1])})
		}
	}
	return events, scanner.Err()
}

func scanQuestLog(path string, since time.Time) ([]historyEvent, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var events []historyEvent
	scanner := newLogScanner(f)
	var notificationAt time.Time
	var jsonLines []string

	flush := func() {
		if notificationAt.IsZero() || notificationAt.Before(since) || len(jsonLines) == 0 {
			notificationAt = time.Time{}
			jsonLines = nil
			return
		}
		var rec pushNotification
		if json.Unmarshal([]byte(strings.Join(jsonLines, "\n")), &rec) == nil &&
			statusToString(rec.Message.Type) == questCompleteStatus {
			questId := strings.ToLower(strings.SplitN(rec.Message.TemplateId, " ", 2)[0])
			if questIdRe.MatchString(questId) {
				events = append(events, historyEvent{at: notificationAt, kind: historyQuest, value: questId})
			}
		}
		notificationAt = time.Time{}
		jsonLines = nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		if at, ok := parseLogTime(line); ok {
			flush()
			if strings.Contains(line, taskSubstring) {
				notificationAt = at
			}
			continue
		}
		if !notificationAt.IsZero() {
			jsonLines = append(jsonLines, line)
		}
	}
	flush()
	return events, scanner.Err()
}

func newLogScanner(f *os.File) *bufio.Scanner {
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	return scanner
}

func parseLogTime(line string) (time.Time, bool) {
	stamp := lineStartWithDateRe.FindString(line)
	if stamp == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{"2006-01-02 15:04:05.000", "2006-01-02 3:04:05.000"} {
		if parsed, err := time.ParseInLocation(layout, stamp, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func correlateSessionEvents(events []historyEvent, knownModes profileModes) []questCompletion {
	profilesInSession := map[string]struct{}{}
	for _, event := range events {
		if event.kind == historyProfile {
			profilesInSession[event.value] = struct{}{}
		}
	}

	var onlyProfile string
	if len(profilesInSession) == 1 {
		for profileKey := range profilesInSession {
			onlyProfile = profileKey
		}
	}

	var currentProfile, currentMode string
	var completions []questCompletion
	for _, event := range events {
		switch event.kind {
		case historyMode:
			currentMode = event.value
			rememberProfileMode(knownModes, currentProfile, currentMode)
		case historyProfile:
			currentProfile = event.value
			rememberProfileMode(knownModes, currentProfile, currentMode)
		case historyQuest:
			profileKey := currentProfile
			if profileKey == "" {
				profileKey = onlyProfile
			}
			rememberProfileMode(knownModes, profileKey, currentMode)
			completions = append(completions, questCompletion{
				at: event.at, profileKey: profileKey, gameMode: currentMode, questId: event.value,
			})
		}
	}
	return completions
}

func rememberProfileMode(modes profileModes, profileKey, gameMode string) {
	if profileKey == "" || gameMode == "" {
		return
	}
	if modes[profileKey] == nil {
		modes[profileKey] = map[string]struct{}{}
	}
	modes[profileKey][gameMode] = struct{}{}
}

func buildQuestHistoryScan(completions []questCompletion, knownModes profileModes) QuestHistoryScan {
	type profileAccumulator struct {
		QuestScanProfile
		quests map[string]struct{}
	}
	profiles := map[string]*profileAccumulator{}
	result := QuestHistoryScan{}

	for _, completion := range completions {
		if completion.profileKey == "" {
			result.SkippedEvents++
			continue
		}
		gameMode := completion.gameMode
		if modes := knownModes[completion.profileKey]; len(modes) == 1 {
			for inferred := range modes {
				gameMode = inferred
			}
		}
		if gameMode == "" {
			result.SkippedEvents++
			continue
		}

		key := gameMode + ":" + completion.profileKey
		acc := profiles[key]
		if acc == nil {
			acc = &profileAccumulator{
				QuestScanProfile: QuestScanProfile{
					ProfileKey:   completion.profileKey,
					GameMode:     gameMode,
					FirstEventAt: completion.at,
					LastEventAt:  completion.at,
				},
				quests: map[string]struct{}{},
			}
			profiles[key] = acc
		}
		acc.quests[completion.questId] = struct{}{}
		if completion.at.Before(acc.FirstEventAt) {
			acc.FirstEventAt = completion.at
		}
		if completion.at.After(acc.LastEventAt) {
			acc.LastEventAt = completion.at
		}
	}

	for _, acc := range profiles {
		for questId := range acc.quests {
			acc.QuestIds = append(acc.QuestIds, questId)
		}
		sort.Strings(acc.QuestIds)
		result.Profiles = append(result.Profiles, acc.QuestScanProfile)
	}
	sort.Slice(result.Profiles, func(i, j int) bool {
		if result.Profiles[i].GameMode != result.Profiles[j].GameMode {
			return result.Profiles[i].GameMode < result.Profiles[j].GameMode
		}
		return result.Profiles[i].FirstEventAt.Before(result.Profiles[j].FirstEventAt)
	})
	return result
}

func normalizeGameMode(mode string) string {
	switch strings.ToLower(mode) {
	case "pve":
		return "PVE"
	case "pvp-season":
		return "SEASON"
	default:
		return "PVP"
	}
}

func hashProfileId(profileId string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(profileId)))
	return hex.EncodeToString(sum[:])
}
