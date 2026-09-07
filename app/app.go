package main

// App — the application core: wires config, folder detection, watchers and webhooks,
// and exposes state to the frontend (Wails bindings + the "state" event).

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/systray"
	"github.com/gen2brain/beeep"
	"github.com/ggdiam/TarkovPilot/app/internal/api"
	"github.com/ggdiam/TarkovPilot/app/internal/config"
	"github.com/ggdiam/TarkovPilot/app/internal/detect"
	"github.com/ggdiam/TarkovPilot/app/internal/updater"
	"github.com/ggdiam/TarkovPilot/app/internal/watcher"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	heartbeatInterval = 60 * time.Second
	// polling the website version (/api/be/pilot/version) — works without a key too
	versionCheckInterval = 10 * time.Minute
	maxEventLog          = 30
)

// State — a state snapshot for the frontend
type State struct {
	Version       string `json:"version"`
	LatestVersion string `json:"latestVersion"`
	UpdateReady   bool   `json:"updateReady"`

	HookId string `json:"hookId"`
	Host   string `json:"host"`
	Lang   string `json:"lang"`

	GameFolder          string `json:"gameFolder"`
	GameFolderOverride  string `json:"gameFolderOverride"`
	ScreenshotsFolder   string `json:"screenshotsFolder"`
	ScreenshotsOverride string `json:"screenshotsOverride"`

	GameFound           bool `json:"gameFound"`
	LogsWatching        bool `json:"logsWatching"`
	ScreenshotsWatching bool `json:"screenshotsWatching"`
	Pro                 bool `json:"pro"`

	// server connection state: nokey | checking | ok | badkey | offline
	ConnState string `json:"connState"`

	AutoStart      bool     `json:"autoStart"`
	AutoClean      bool     `json:"autoClean"`
	StartMinimized bool     `json:"startMinimized"`
	EventLog       []string `json:"eventLog"`
}

type App struct {
	ctx context.Context

	mu            sync.Mutex
	latestVersion string
	// version we already showed a Windows notification for (don't spam every 10 min)
	notifiedVersion string
	eventLog        []string

	client  *api.Client
	screens *watcher.ScreensWatcher
	logs    *watcher.LogsWatcher

	// re-localizes the tray menu on language change (set in onTrayReady)
	trayUpdate func(lang string)

	statusCh chan struct{}

	questSyncMu sync.Mutex
}

func NewApp() *App {
	a := &App{
		screens:  watcher.NewScreensWatcher(),
		logs:     watcher.NewLogsWatcher(),
		statusCh: make(chan struct{}, 1),
	}

	a.client = api.NewClient(
		func() string { return config.Get().HookId },
		func() string { return config.Get().Host },
	)

	a.screens.OnScreenshot = a.onScreenshot
	a.screens.OnError = func(err error) { a.logEvent("Screenshots watcher error: " + err.Error()) }
	a.logs.OnMap = a.onMapChange
	a.logs.OnQuest = a.onQuest
	a.logs.OnLog = a.logEvent

	return a
}

// startup is called by Wails on start
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// settings are loaded in main() before the window is created
	updater.CleanupOldBinary()

	// the log file must not grow forever — on start drop it if it got big
	if fi, err := os.Stat(eventsLogPath()); err == nil && fi.Size() > 1<<20 {
		_ = os.Remove(eventsLogPath())
	}
	a.logEvent("App started, version " + Version)

	a.restartWatchers()

	// heartbeat: immediately, then once a minute; plus on demand via statusCh
	go func() {
		a.sendStatus()
		ticker := time.NewTicker(heartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				a.retryWatchers()
				a.sendStatus()
			case <-a.statusCh:
				a.sendStatus()
			}
		}
	}()

	// website version every 10 minutes while the app is alive: works without a key,
	// on a new version — Windows notification + Update banner in the window
	go func() {
		a.checkLatestVersion()
		ticker := time.NewTicker(versionCheckInterval)
		defer ticker.Stop()
		for range ticker.C {
			a.checkLatestVersion()
		}
	}()
}

// shutdown is called by Wails on exit: if auto-clean is enabled,
// delete the current session's screenshots (same as on map change)
func (a *App) shutdown(ctx context.Context) {
	if config.Get().AutoClean {
		if n := a.screens.CleanSessionScreenshots(); n > 0 {
			a.logEvent(fmt.Sprintf("Screenshots deleted on exit: %d", n))
		}
	}
}

// requestStatusSend — request an out-of-schedule heartbeat (after settings change)
func (a *App) requestStatusSend() {
	select {
	case a.statusCh <- struct{}{}:
	default:
	}
}

// restartWatchers recreates the watchers from current settings
func (a *App) restartWatchers() {
	cfg := config.Get()

	gameFolder := detect.GameFolder(cfg.GameFolder)
	screensFolder := detect.ScreenshotsFolder(cfg.ScreenshotsFolder)

	a.logs.Start(detect.LogsFolder(gameFolder))

	if err := a.screens.Start(screensFolder); err != nil {
		a.logEvent("Screenshots folder not found: " + screensFolder)
	}

	a.emitState()
}

// retryWatchers — quiet watcher restart from the heartbeat tick.
// The needed folders may appear after the app has started: the game creates
// the screenshots folder on the first screenshot and the logs folder on the
// first game launch. Without the retry the feature would stay silent until
// the app is restarted
func (a *App) retryWatchers() {
	cfg := config.Get()

	if !a.screens.Watching() {
		folder := detect.ScreenshotsFolder(cfg.ScreenshotsFolder)
		if err := a.screens.Start(folder); err == nil {
			a.logEvent("Screenshots folder found: " + folder)
			a.emitState()
		}
	}

	if !a.logs.Watching() {
		// the path may resolve later (the game got installed/found) — re-detect
		a.logs.Start(detect.LogsFolder(detect.GameFolder(cfg.GameFolder)))
		if a.logs.Watching() {
			a.logEvent("Logs folder found")
			a.emitState()
		}
	}
}

// --- watcher events ---

func (a *App) onScreenshot(filename string) {
	a.logEvent("Screenshot: " + filename)
	go func() {
		if err := a.client.SendScreenshot(filename); err != nil {
			a.logEvent("Send error: " + err.Error())
		}
		a.emitState()
	}()
}

func (a *App) onMapChange(rawMap string) {
	a.logEvent("Map: " + rawMap)
	go func() {
		if err := a.client.SendMap(rawMap); err != nil {
			a.logEvent("Send error: " + err.Error())
		}
		// like the ps1 script: on map change clean up the past session's screenshots
		if config.Get().AutoClean {
			if n := a.screens.CleanSessionScreenshots(); n > 0 {
				a.logEvent(fmt.Sprintf("Screenshots deleted: %d", n))
			}
		}
		a.emitState()
	}()
}

func (a *App) onQuest(questId, status, profileKey, gameMode string) {
	a.logEvent("Quest update: " + questId + " status: " + status)
	go func() {
		if err := a.client.SendQuest(questId, status, profileKey, gameMode); err != nil {
			a.logEvent("Send error: " + err.Error())
		}
		a.emitState()
	}()
}

// --- status ---

func (a *App) currentAppStatus() api.AppStatus {
	cfg := config.Get()
	gameFolder := detect.GameFolder(cfg.GameFolder)
	return api.AppStatus{
		Version:             Version,
		GameFolder:          gameFolder,
		ScreenshotsFolder:   detect.ScreenshotsFolder(cfg.ScreenshotsFolder),
		GameFound:           detect.GameFound(gameFolder),
		LogsWatching:        a.logs.Watching(),
		ScreenshotsWatching: a.screens.Watching(),
	}
}

func (a *App) sendStatus() {
	if config.Get().HookId == "" {
		return
	}
	wasOK := a.client.ServerOK()
	wasChecked := a.client.ServerChecked()
	wasBad := a.client.BadKey()
	latest, err := a.client.SendStatus(a.currentAppStatus())
	a.setLatestVersion(latest)

	// show heartbeat errors in the event log only on transitions,
	// to avoid spamming every minute
	if errors.Is(err, api.ErrBadKey) {
		if !wasBad {
			a.logEvent("Connection key rejected by server")
		}
	} else if err != nil && (!wasChecked || wasOK) {
		a.logEvent("Server error: " + err.Error())
	}
	a.emitState()
}

// checkLatestVersion — poll the website version (the keyless channel)
func (a *App) checkLatestVersion() {
	v, err := a.client.FetchLatestVersion()
	if err != nil {
		return
	}
	a.setLatestVersion(v)
}

// versionNewer — true if a is newer than b (numeric comparison of "x.y.z" segments).
// A plain "a != b" won't do: the website may carry a version older than the app
// (e.g. before a release), and the update banner must not be shown then
func versionNewer(a, b string) bool {
	pa, pb := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(pa) || i < len(pb); i++ {
		na, nb := 0, 0
		if i < len(pa) {
			na, _ = strconv.Atoi(strings.TrimSpace(pa[i]))
		}
		if i < len(pb) {
			nb, _ = strconv.Atoi(strings.TrimSpace(pb[i]))
		}
		if na != nb {
			return na > nb
		}
	}
	return false
}

// setLatestVersion remembers the latest version; a new version is notified
// once (Windows notification), after that the Update banner in the window shows it
func (a *App) setLatestVersion(v string) {
	if v == "" {
		return
	}
	a.mu.Lock()
	changed := a.latestVersion != v
	a.latestVersion = v
	needNotify := versionNewer(v, Version) && a.notifiedVersion != v
	if needNotify {
		a.notifiedVersion = v
	}
	a.mu.Unlock()

	if needNotify {
		msg := "New version " + v + " is available — click Update in the TarkovPilot window"
		if config.Get().Lang == "ru" {
			msg = "Доступна новая версия " + v + " — нажмите «Обновить» в окне TarkovPilot"
		}
		if err := beeep.Notify("TarkovPilot", msg, ""); err != nil {
			a.logEvent("Notify error: " + err.Error())
		}
		a.logEvent("New version available: " + v)
	}
	if changed {
		a.emitState()
	}
}

// connState — aggregated connection status for the UI
func (a *App) connState() string {
	if config.Get().HookId == "" {
		return "nokey"
	}
	if !a.client.ServerChecked() {
		return "checking"
	}
	if a.client.BadKey() {
		return "badkey"
	}
	if a.client.ServerOK() {
		return "ok"
	}
	return "offline"
}

// --- event log for the UI ---

// eventsLogPath — file copy of the event log (%APPDATA%\TarkovPilot\events.log):
// the UI log lives in memory and is lost on restart, the file remains for diagnostics
func eventsLogPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "TarkovPilot", "events.log")
}

func appendEventsFile(msg string) {
	path := eventsLogPath()
	if path == "" {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(time.Now().Format("2006-01-02 15:04:05") + "  " + msg + "\n")
}

func (a *App) logEvent(msg string) {
	a.mu.Lock()
	stamp := time.Now().Format("15:04:05")
	a.eventLog = append([]string{stamp + "  " + msg}, a.eventLog...)
	if len(a.eventLog) > maxEventLog {
		a.eventLog = a.eventLog[:maxEventLog]
	}
	a.mu.Unlock()
	appendEventsFile(msg)
	a.emitState()
}

// --- state for the frontend ---

func (a *App) buildState() State {
	cfg := config.Get()
	st := a.currentAppStatus()

	a.mu.Lock()
	latest := a.latestVersion
	eventLog := append([]string{}, a.eventLog...)
	a.mu.Unlock()

	return State{
		Version:       Version,
		LatestVersion: latest,
		UpdateReady:   versionNewer(latest, Version),

		HookId: cfg.HookId,
		Host:   cfg.Host,
		Lang:   cfg.Lang,

		GameFolder:          st.GameFolder,
		GameFolderOverride:  cfg.GameFolder,
		ScreenshotsFolder:   st.ScreenshotsFolder,
		ScreenshotsOverride: cfg.ScreenshotsFolder,

		GameFound:           st.GameFound,
		LogsWatching:        st.LogsWatching,
		ScreenshotsWatching: st.ScreenshotsWatching,
		Pro:                 a.client.Pro(),
		ConnState:           a.connState(),

		AutoStart:      isAutoStartEnabled(),
		AutoClean:      cfg.AutoClean,
		StartMinimized: cfg.StartMinimized,
		EventLog:       eventLog,
	}
}

func (a *App) emitState() {
	if a.ctx == nil {
		return
	}
	wailsRuntime.EventsEmit(a.ctx, "state", a.buildState())
}

// ================= frontend bindings =================

func (a *App) GetState() State { return a.buildState() }

func (a *App) SetHookId(hookId string) State {
	hookId = strings.TrimSpace(hookId)
	config.Update(func(s *config.Settings) { s.HookId = hookId })
	a.client.ResetCheck()
	if hookId == "" {
		a.logEvent("Connection key removed")
		return a.buildState()
	}
	a.logEvent("Connection key saved")
	// check the key synchronously: the UI waits for "accepted / rejected / offline"
	a.sendStatus()
	return a.buildState()
}

// TestConnection sends the current app status immediately without changing the saved key.
func (a *App) TestConnection() State {
	a.sendStatus()
	return a.buildState()
}

// SetRegion — the "Region" switch: UI language and website change together
// (en → tarkov-market.com, ru → tarkov-market.ru)
func (a *App) SetRegion(region string) State {
	if region != "ru" {
		region = "en"
	}
	host := "tarkov-market.com"
	if region == "ru" {
		host = "tarkov-market.ru"
	}
	config.Update(func(s *config.Settings) {
		s.Lang = region
		s.Host = host
	})
	if a.trayUpdate != nil {
		a.trayUpdate(region)
	}
	// host changed — re-check the connection and the version
	a.client.ResetCheck()
	a.requestStatusSend()
	go a.checkLatestVersion()
	return a.buildState()
}

func (a *App) SetAutoClean(enabled bool) State {
	config.Update(func(s *config.Settings) { s.AutoClean = enabled })
	return a.buildState()
}

// SetStartMinimized — start in the tray on every launch (takes effect on the next start)
func (a *App) SetStartMinimized(enabled bool) State {
	config.Update(func(s *config.Settings) { s.StartMinimized = enabled })
	return a.buildState()
}

func (a *App) SetAutoStart(enabled bool) State {
	if err := setAutoStart(enabled); err != nil {
		a.logEvent("Autostart error: " + err.Error())
	}
	return a.buildState()
}

// BrowseGameFolder — game folder picker dialog (empty string = cancel)
func (a *App) BrowseGameFolder() State {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Escape from Tarkov folder",
	})
	if err != nil {
		a.logEvent("Folder dialog error: " + err.Error())
	} else if dir != "" {
		a.logEvent("Game folder set: " + dir)
		config.Update(func(s *config.Settings) { s.GameFolder = dir })
		a.restartWatchers()
		a.requestStatusSend()
	}
	return a.buildState()
}

func (a *App) ClearGameFolder() State {
	config.Update(func(s *config.Settings) { s.GameFolder = "" })
	a.logEvent("Game folder: auto-detect")
	a.restartWatchers()
	a.requestStatusSend()
	return a.buildState()
}

func (a *App) BrowseScreenshotsFolder() State {
	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "EFT Screenshots folder",
	})
	if err != nil {
		a.logEvent("Folder dialog error: " + err.Error())
	} else if dir != "" {
		a.logEvent("Screenshots folder set: " + dir)
		config.Update(func(s *config.Settings) { s.ScreenshotsFolder = dir })
		a.restartWatchers()
		a.requestStatusSend()
	}
	return a.buildState()
}

func (a *App) ClearScreenshotsFolder() State {
	config.Update(func(s *config.Settings) { s.ScreenshotsFolder = "" })
	a.logEvent("Screenshots folder: auto-detect")
	a.restartWatchers()
	a.requestStatusSend()
	return a.buildState()
}

// ScanQuestHistory scans local EFT logs from 2026 onward and sends only hashed
// profile keys and completed quest ids to the website.
func (a *App) ScanQuestHistory() api.QuestSyncResponse {
	a.questSyncMu.Lock()
	defer a.questSyncMu.Unlock()

	if config.Get().HookId == "" {
		return questSyncFailure("connection_required")
	}
	if !a.client.Pro() {
		return questSyncFailure("pro_required")
	}

	cfg := config.Get()
	logsFolder := detect.LogsFolder(detect.GameFolder(cfg.GameFolder))
	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	scan, err := watcher.ScanQuestHistory(logsFolder, since)
	if err != nil {
		a.logEvent("Quest history scan failed: " + err.Error())
		return questSyncFailure("logs_not_found")
	}

	resp, err := a.client.SendQuestScan(scan.Profiles, scan.SkippedEvents)
	if err != nil {
		a.logEvent("Quest history upload failed: " + err.Error())
		a.emitState()
		return questSyncFailure("network_error")
	}
	if !resp.Ok {
		a.emitState()
		return resp
	}

	quests := 0
	for _, profile := range scan.Profiles {
		quests += len(profile.QuestIds)
	}
	a.logEvent(fmt.Sprintf("Quest history scan: %d profiles, %d completed quests", len(scan.Profiles), quests))
	return resp
}

// ImportQuestProfile imports one scanned local profile into a website character.
func (a *App) ImportQuestProfile(profileKey, charUID string) api.QuestSyncResponse {
	a.questSyncMu.Lock()
	defer a.questSyncMu.Unlock()

	resp, err := a.client.ImportQuestProfile(profileKey, charUID)
	if err != nil {
		a.logEvent("Quest history import failed: " + err.Error())
		a.emitState()
		return questSyncFailure("network_error")
	}
	if resp.Ok && resp.ImportResult != nil {
		a.logEvent(fmt.Sprintf("Quest history imported: %d new quests", resp.ImportResult.ImportedCount))
	}
	a.emitState()
	return resp
}

// CreateQuestCharacterAndImport creates a website character and imports a scanned profile.
func (a *App) CreateQuestCharacterAndImport(profileKey, fraction string) api.QuestSyncResponse {
	a.questSyncMu.Lock()
	defer a.questSyncMu.Unlock()

	resp, err := a.client.CreateQuestCharacterAndImport(profileKey, fraction)
	if err != nil {
		a.logEvent("Quest character creation failed: " + err.Error())
		a.emitState()
		return questSyncFailure("network_error")
	}
	if resp.Ok && resp.ImportResult != nil {
		a.logEvent(fmt.Sprintf("Quest character created and imported: %d new quests", resp.ImportResult.ImportedCount))
	}
	a.emitState()
	return resp
}

func questSyncFailure(code string) api.QuestSyncResponse {
	return api.QuestSyncResponse{Ok: false, Error: code}
}

// OpenPilotPage opens the Pilot page on the website (where the connection key is).
// ?way=app — the website opens the "App" tab right away, avoiding confusion with the script instructions
func (a *App) OpenPilotPage() {
	wailsRuntime.BrowserOpenURL(a.ctx, "https://"+config.Get().Host+"/pilot?way=app")
}

// DoUpdate starts self-update; on success the process will restart
func (a *App) DoUpdate() State {
	a.logEvent("Updating...")
	if err := updater.Update(config.Get().Host); err != nil {
		a.logEvent("Update failed: " + err.Error())
	}
	return a.buildState()
}

// Quit — full exit: window, tray, process
func (a *App) Quit() {
	systray.Quit()
	if a.ctx != nil {
		wailsRuntime.Quit(a.ctx)
	}
}
