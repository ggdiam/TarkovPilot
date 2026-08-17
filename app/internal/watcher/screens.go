// Screenshots folder watcher: a new *.png → callback with the file name.
// The file name contains coordinates — the website parses them, the app doesn't read the file.
package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type ScreensWatcher struct {
	mu      sync.Mutex
	watcher *fsnotify.Watcher
	folder  string
	// when watching started — for auto-cleaning session screenshots
	startedAt time.Time

	OnScreenshot func(filename string)
	OnError      func(err error)
}

func NewScreensWatcher() *ScreensWatcher {
	return &ScreensWatcher{}
}

// Watching — the folder is being watched
func (w *ScreensWatcher) Watching() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.watcher != nil
}

// Start begins watching (stopping the previous watch)
func (w *ScreensWatcher) Start(folder string) error {
	w.Stop()

	fi, err := os.Stat(folder)
	if err != nil || !fi.IsDir() {
		return os.ErrNotExist
	}

	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := fsw.Add(folder); err != nil {
		fsw.Close()
		return err
	}

	w.mu.Lock()
	w.watcher = fsw
	w.folder = folder
	w.startedAt = time.Now()
	w.mu.Unlock()

	go func() {
		for {
			select {
			case ev, ok := <-fsw.Events:
				if !ok {
					return
				}
				if ev.Op.Has(fsnotify.Create) && strings.EqualFold(filepath.Ext(ev.Name), ".png") {
					if w.OnScreenshot != nil {
						w.OnScreenshot(filepath.Base(ev.Name))
					}
				}
			case err, ok := <-fsw.Errors:
				if !ok {
					return
				}
				if w.OnError != nil {
					w.OnError(err)
				}
			}
		}
	}()

	return nil
}

func (w *ScreensWatcher) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.watcher != nil {
		w.watcher.Close()
		w.watcher = nil
	}
}

// CleanSessionScreenshots deletes screenshots taken after watching started
// (called on map change — like in the ps1 script). Returns the number deleted
func (w *ScreensWatcher) CleanSessionScreenshots() int {
	w.mu.Lock()
	folder := w.folder
	startedAt := w.startedAt
	w.mu.Unlock()

	if folder == "" {
		return 0
	}

	deleted := 0
	entries, err := os.ReadDir(folder)
	if err != nil {
		return 0
	}
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".png") {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		if fi.ModTime().After(startedAt) {
			if os.Remove(filepath.Join(folder, e.Name())) == nil {
				deleted++
			}
		}
	}
	return deleted
}
