package main

// TarkovPilot v2 — Go/Wails. Settings window + tray icon; closing the window
// minimizes to tray, quitting is done via the tray menu.
// Data goes to the TM website via webhooks, the browser no longer connects to the app.

import (
	"context"
	"embed"
	"os"
	"time"

	"fyne.io/systray"
	"github.com/ggdiam/TarkovPilot/app/internal/config"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
	sysWindows "golang.org/x/sys/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/windows/icon.ico
var trayIcon []byte

func main() {
	args := os.Args[1:]
	// settings are needed before the window is created: the "start minimized"
	// option decides whether the window is shown at all
	cfg := config.Load()
	startHidden := hasLaunchArg(args, autoStartArg) || cfg.StartMinimized

	// single instance (like the C# version: a named mutex).
	// After self-update the updater launches the new exe with the "updated" arg
	// BEFORE the old process has exited and released the mutex — in that
	// case don't give up right away, wait for it to be released
	if !acquireSingleInstance(hasLaunchArg(args, "updated")) {
		return
	}

	app := NewApp()
	// Rewrite legacy autostart entries once the new version runs, so users who
	// already enabled the option also start directly in the tray next time.
	if isAutoStartEnabled() {
		if err := setAutoStart(true); err != nil {
			app.logEvent("Autostart error: " + err.Error())
		}
	}

	err := wails.Run(&options.App{
		Title:  "TarkovPilot",
		Width:  760,
		Height: 640,
		// no native Windows frame: the frontend draws its own title bar
		// (header is a drag zone via --wails-draggable, minimize/close buttons are custom)
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 1},
		// closing the window doesn't kill the app — it lives in the tray
		HideWindowOnClose: true,
		// Windows autostart passes --hidden; a regular launch remains visible
		// unless the "start minimized" option is enabled.
		StartHidden: startHidden,
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			// systray in its own goroutine: systray.Run blocks; on Windows
			// running it off the main thread is fine
			go systray.Run(onTrayReady(app), nil)
		},
		// on exit — auto-clean session screenshots (if enabled)
		OnShutdown: app.shutdown,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
		},
	})
	if err != nil {
		println("Error:", err.Error())
	}
}

func hasLaunchArg(args []string, target string) bool {
	for _, arg := range args {
		if arg == target {
			return true
		}
	}
	return false
}

// acquireSingleInstance acquires the single-instance named mutex.
// wait — wait up to 10 seconds for the exiting old process to release the
// mutex (self-update scenario). The handle of a failed attempt must be
// closed: the mutex object lives while at least one handle is open — by
// holding it we would prevent the mutex from dying with the old process
func acquireSingleInstance(wait bool) bool {
	name := sysWindows.StringToUTF16Ptr("TarkovPilotMutex")
	attempts := 1
	if wait {
		attempts = 100
	}
	for i := 0; i < attempts; i++ {
		h, err := sysWindows.CreateMutex(nil, false, name)
		if err != sysWindows.ERROR_ALREADY_EXISTS {
			// mutex is ours; don't close the handle — it lives until the process exits
			return true
		}
		_ = sysWindows.CloseHandle(h)
		if i < attempts-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}
	return false
}

// tray labels by UI language
var trayLabels = map[string][2]string{
	"en": {"Show", "Quit"},
	"ru": {"Показать", "Выход"},
}

// onTrayReady sets up the tray menu
func onTrayReady(app *App) func() {
	return func() {
		systray.SetIcon(trayIcon)
		systray.SetTitle("TarkovPilot")
		systray.SetTooltip("TarkovPilot")

		labels := trayLabels[config.Get().Lang]
		if labels[0] == "" {
			labels = trayLabels["en"]
		}

		mShow := systray.AddMenuItem(labels[0], "")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem(labels[1], "")

		// changing the UI language re-localizes the tray too
		app.trayUpdate = func(lang string) {
			l := trayLabels[lang]
			if l[0] == "" {
				l = trayLabels["en"]
			}
			mShow.SetTitle(l[0])
			mQuit.SetTitle(l[1])
		}

		go func() {
			for {
				select {
				case <-mShow.ClickedCh:
					if app.ctx != nil {
						wailsRuntime.WindowShow(app.ctx)
					}
				case <-mQuit.ClickedCh:
					systray.Quit()
					if app.ctx != nil {
						wailsRuntime.Quit(app.ctx)
					}
				}
			}
		}()
	}
}
