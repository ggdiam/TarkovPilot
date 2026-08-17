package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPollCallsOnLogWithoutWatcherLock(t *testing.T) {
	logsFolder := t.TempDir()
	if err := os.Mkdir(filepath.Join(logsFolder, "log_session"), 0o755); err != nil {
		t.Fatalf("failed to create the test session folder: %v", err)
	}

	w := NewLogsWatcher()
	w.logsFolder = logsFolder
	w.stopCh = make(chan struct{})
	w.OnLog = func(string) {
		// The real callback builds app state and calls Watching.
		// If poll still holds w.mu, this call blocks forever.
		_ = w.Watching()
	}

	done := make(chan struct{})
	go func() {
		w.poll()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("poll deadlocked while calling OnLog")
	}
}
