# TarkovPilot v2 (Go + Wails)

Tarkov Pilot desktop app: settings window + tray icon. Watches EFT screenshots
and logs and sends events to the tarkov-market website via webhooks. The browser
does not connect to the app.

## Build

Requires Go 1.24+ and Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).

```powershell
cd app
wails build        # → build/bin/TarkovPilot.exe
```

The frontend is static files in `frontend/dist` (vanilla HTML/CSS/JS), no node/npm needed.

## Release

The app version is set in `version.go`. Building and rollout to the website servers
is done by an internal deploy script (not in this repo — it is public): it runs
`wails build`, packs the exe into `TarkovPilot.zip` and publishes it together with
`version.txt`. Running apps see the new version and show the Update button themselves.

## Structure

- `main.go` — Wails bootstrap, single instance (mutex), systray (fyne.io/systray)
- `autostart.go` — Windows Run entry; autostart uses `--hidden` to launch directly into the tray
  (the "Start minimized" option does the same for every launch)
- `app.go` — the core: state, frontend bindings, heartbeat (60 sec), website version
  polling (10 min, `GET /api/be/pilot/version`, host by region; newer — Update banner +
  a single Windows notification), connection status `connState`, watcher events
- `internal/config` — settings in `%APPDATA%\TarkovPilot\settings.json`
- `internal/detect` — game folder detection (BSG registry / Steam vdf / default) and screenshots
- `internal/watcher` — screenshots (fsnotify) and logs (1 sec polling, tail from the end)
- `internal/api` — webhooks `/api/be/pilot/script/webhook` (hookId)
- `internal/updater` — self-update: rename the running exe → replace → restart
- `frontend/dist` — UI (dark flat style, en/ru)
