// Package detect — locating the EFT game folder (BSG launcher / Steam / default) and the screenshots folder.
// The logic mirrors the ps1 script and the C# version: candidates are collected from the registry
// and Steam libraries, and the one with a Logs folder holding the newest session folder wins.
package detect

import (
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const defaultGameFolder = `C:\Battlestate Games\EFT`

// GameFolder returns the game folder: the settings override or auto-detect
func GameFolder(override string) string {
	if override != "" {
		return override
	}

	candidates := []string{}

	// 1) BSG launcher: registry
	if k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\EscapeFromTarkov`,
		registry.QUERY_VALUE); err == nil {
		if v, _, err := k.GetStringValue("InstallLocation"); err == nil && v != "" {
			candidates = append(candidates, v)
		}
		k.Close()
	}

	// 2) Steam libraries (libraryfolders.vdf); the Steam version keeps the game in build
	for _, lib := range steamLibraries() {
		candidates = append(candidates, filepath.Join(lib, `steamapps\common\Escape from Tarkov\build`))
	}

	// 3) default standalone path
	candidates = append(candidates, defaultGameFolder)

	// pick the candidate with an existing Logs and the newest session folder
	best := ""
	bestTime := time.Time{}
	seen := map[string]bool{}
	for _, c := range candidates {
		if c == "" || seen[c] {
			continue
		}
		seen[c] = true

		logs := filepath.Join(c, "Logs")
		fi, err := os.Stat(logs)
		if err != nil || !fi.IsDir() {
			continue
		}

		t := newestSubfolderTime(logs)
		if t.IsZero() {
			t = creationTime(fi)
		}
		if best == "" || t.After(bestTime) {
			best = c
			bestTime = t
		}
	}

	if best != "" {
		return best
	}
	return defaultGameFolder
}

// GameFound — the folder exists and looks like an EFT folder: it contains Logs.
// A bare "folder exists" lies to the UI: picking any unrelated folder, the
// user saw a green status (real case on Aug 16 with Documents\Graphics)
func GameFound(gameFolder string) bool {
	fi, err := os.Stat(gameFolder)
	if err != nil || !fi.IsDir() {
		return false
	}
	lf, err := os.Stat(filepath.Join(gameFolder, "Logs"))
	return err == nil && lf.IsDir()
}

// LogsFolder — the logs folder inside the game folder
func LogsFolder(gameFolder string) string {
	return filepath.Join(gameFolder, "Logs")
}

// ScreenshotsFolder returns the screenshots folder: the override or Documents\Escape from Tarkov\Screenshots.
// Documents is resolved via Known Folders — the folder may be relocated out of the profile
func ScreenshotsFolder(override string) string {
	if override != "" {
		return override
	}

	docs, err := windows.KnownFolderPath(windows.FOLDERID_Documents, 0)
	if err != nil || docs == "" {
		home, _ := os.UserHomeDir()
		docs = filepath.Join(home, "Documents")
	}
	return filepath.Join(docs, "Escape from Tarkov", "Screenshots")
}

var vdfPathRe = regexp.MustCompile(`"path"\s+"(.+?)"`)

// steamLibraries — paths of all Steam libraries from libraryfolders.vdf
func steamLibraries() []string {
	steam := ""
	if k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Valve\Steam`, registry.QUERY_VALUE); err == nil {
		steam, _, _ = k.GetStringValue("SteamPath")
		k.Close()
	}
	if steam == "" {
		if k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\WOW6432Node\Valve\Steam`, registry.QUERY_VALUE); err == nil {
			steam, _, _ = k.GetStringValue("InstallPath")
			k.Close()
		}
	}
	if steam == "" {
		return nil
	}

	libs := []string{steam}
	vdf := filepath.Join(steam, `steamapps\libraryfolders.vdf`)
	if b, err := os.ReadFile(vdf); err == nil {
		for _, m := range vdfPathRe.FindAllStringSubmatch(string(b), -1) {
			p := filepath.FromSlash(m[1])
			// backslashes are escaped in the vdf
			p = regexp.MustCompile(`\\\\`).ReplaceAllString(p, `\`)
			libs = append(libs, p)
		}
	}
	return libs
}

// newestSubfolderTime — creation time of the newest subfolder (log session)
func newestSubfolderTime(dir string) time.Time {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return time.Time{}
	}
	best := time.Time{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		t := creationTime(fi)
		if t.After(best) {
			best = t
		}
	}
	return best
}

// NewestSessionFolder — the newest session folder in Logs (by creation time)
func NewestSessionFolder(logsDir string) string {
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return ""
	}
	best := ""
	bestTime := time.Time{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		t := creationTime(fi)
		if best == "" || t.After(bestTime) {
			best = filepath.Join(logsDir, e.Name())
			bestTime = t
		}
	}
	return best
}

// creationTime — Windows CreationTime from FileInfo (ModTime of log folders may
// update on appends, while we care about the session creation moment — like C#/ps1)
func creationTime(fi os.FileInfo) time.Time {
	if d, ok := fi.Sys().(*syscall.Win32FileAttributeData); ok {
		return time.Unix(0, d.CreationTime.Nanoseconds())
	}
	return fi.ModTime()
}
