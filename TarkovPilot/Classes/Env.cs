using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;
using System.Diagnostics;
using System.IO;
using Microsoft.Win32;
using System.Windows.Forms;

namespace TarkovPilot
{
    public static class Env
    {
        static Env()
        {
            FileVersionInfo versionInfo = FileVersionInfo.GetVersionInfo(Path.Combine(AppDomain.CurrentDomain.BaseDirectory, "TarkovPilot.exe"));
            //Logger.Log($"File version: {versionInfo.FileVersion}");

            Version = versionInfo.FileVersion;
        }

        // first logs read on app start
        public static bool InitialLogsRead { get; set; } = true;

        public static string Version = "0.0";

#if DEBUG
        public static string WebsiteUrl = "http://localhost:3000/pilot";
#else
        public static string WebsiteUrl = "https://tarkov-market.com/pilot";
#endif


        private static string _gameFolder = null;
        public static string GameFolder
        {
            get
            {
                if (_gameFolder == null)
                {
                    RegistryKey key = Registry.LocalMachine.OpenSubKey("SOFTWARE\\WOW6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\EscapeFromTarkov");
                    var installPath = key?.GetValue("InstallLocation")?.ToString();
                    key?.Dispose();

                    if (!String.IsNullOrEmpty(installPath))
                    {
                        _gameFolder = installPath;
                    }
                }

                return _gameFolder;
            }

            set { _gameFolder = value; }
        }

        public static string LogsFolder
        {
            get
            {
                return Path.Combine(GameFolder, "Logs"); ;
            }
        }

        private static string _screenshotsFolder;
        public static string ScreenshotsFolder
        {
            get
            {
                if (_screenshotsFolder == null)
                {
                    _screenshotsFolder = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.MyDocuments), "Escape From Tarkov", "Screenshots");
                }
                return _screenshotsFolder;
            }
            set { _screenshotsFolder = value; }
        }

        private static bool _mapChangeEnabled = true;
        public static bool MapChangeEnabled
        {
            get { return _mapChangeEnabled; }
            set { _mapChangeEnabled = value; }
        }

        //===================== AppContext Settings ============================

        public static void SetSettings(AppSettings settings, bool force = false)
        {
            if (force || !String.IsNullOrEmpty(settings.gameFolder))
            {
                Env.GameFolder = settings.gameFolder ?? null;
            }
            if (force || !String.IsNullOrEmpty(settings.screenshotsFolder))
            {
                Env.ScreenshotsFolder = settings.screenshotsFolder ?? null;
            }

            Env.MapChangeEnabled = settings.mapChangeEnabled;
        }

        public static AppSettings GetSettings()
        {
            AppSettings settings = new AppSettings()
            {
                gameFolder = Env.GameFolder,
                screenshotsFolder = Env.ScreenshotsFolder,
                mapChangeEnabled = Env.MapChangeEnabled,
            };
            return settings;
        }

        public static void ResetSettings()
        {
            AppSettings settings = new AppSettings()
            {
                gameFolder = null,
                screenshotsFolder = null,
                mapChangeEnabled = true,
            };
            SetSettings(settings, true);
        }

        //===================== AppContext Settings ============================

        public static void RestartApp()
        {
            Application.Restart();
            Environment.Exit(0);
        }
    }
}
