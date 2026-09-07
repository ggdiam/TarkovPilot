// Package config — app settings in %APPDATA%\TarkovPilot\settings.json.
// We don't write next to the exe: it may live in a folder without write access.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Settings struct {
	// connection key from the website (user.pilot.script.hookId)
	HookId string `json:"hookId"`

	// website host: tarkov-market.com (default) or tarkov-market.ru
	Host string `json:"host"`

	// game folder override; empty — auto-detect
	GameFolder string `json:"gameFolder"`

	// screenshots folder override; empty — Documents\Escape from Tarkov\Screenshots
	ScreenshotsFolder string `json:"screenshotsFolder"`

	// delete current session's screenshots on map change (like the ps1 script)
	AutoClean bool `json:"autoClean"`

	// start directly in the tray on every launch (not only via Windows autostart)
	StartMinimized bool `json:"startMinimized"`

	// UI language: "en" | "ru"
	Lang string `json:"lang"`
}

var (
	mu       sync.Mutex
	settings Settings
)

func dir() string {
	base, err := os.UserConfigDir() // %APPDATA%
	if err != nil {
		base = "."
	}
	return filepath.Join(base, "TarkovPilot")
}

func path() string { return filepath.Join(dir(), "settings.json") }

// Load reads settings; a missing file is not an error (defaults)
func Load() Settings {
	mu.Lock()
	defer mu.Unlock()

	settings = Settings{Host: "tarkov-market.com", AutoClean: true, Lang: "en"}

	b, err := os.ReadFile(path())
	if err == nil {
		_ = json.Unmarshal(b, &settings)
	}
	if settings.Host == "" {
		settings.Host = "tarkov-market.com"
	}
	if settings.Lang == "" {
		settings.Lang = "en"
	}
	return settings
}

func Get() Settings {
	mu.Lock()
	defer mu.Unlock()
	return settings
}

// Update mutates settings via the callback and persists them to disk
func Update(fn func(*Settings)) Settings {
	mu.Lock()
	defer mu.Unlock()

	fn(&settings)

	_ = os.MkdirAll(dir(), 0o755)
	b, err := json.MarshalIndent(settings, "", "    ")
	if err == nil {
		_ = os.WriteFile(path(), b, 0o644)
	}
	return settings
}
