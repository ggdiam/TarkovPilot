package main

// Autostart with Windows: a key in HKCU\Software\Microsoft\Windows\CurrentVersion\Run.
// Does not require administrator rights.

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const runKeyPath = `Software\Microsoft\Windows\CurrentVersion\Run`
const runValueName = "TarkovPilot"
const autoStartArg = "--hidden"

func autoStartCommand(exe string) string {
	return `"` + exe + `" ` + autoStartArg
}

func autoStartValueMatches(value, exe string) bool {
	// Accept the old value so an existing opt-in can be migrated on startup.
	return value == autoStartCommand(exe) || value == `"`+exe+`"`
}

func isAutoStartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	v, _, err := k.GetStringValue(runValueName)
	if err != nil {
		return false
	}

	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return autoStartValueMatches(v, exe)
}

func setAutoStart(enabled bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, runKeyPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if !enabled {
		err := k.DeleteValue(runValueName)
		if err == registry.ErrNotExist {
			return nil
		}
		return err
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return k.SetStringValue(runValueName, autoStartCommand(exe))
}
