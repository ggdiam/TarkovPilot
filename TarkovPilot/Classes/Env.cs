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
        //public static bool InitialLogsRead { get; set; } = true;

        public static string Version = "0.0";

        // дефолтный host для RELEASE; при необходимости перезаписывается значением из settings.json или SETTINGS_INIT
        public const string DEFAULT_HOST = "tarkov-market.com";
        // whitelist разрешённых хостов — защита от подмены апдейта зловредным JS, открытым в браузере юзера
        static readonly string[] AllowedHosts = { "tarkov-market.com", "tarkov-market.ru" };

#if DEBUG
        public static string Scheme = "http";
        public static string WebsiteHost = "localhost:3000";
#else
        public static string Scheme = "https";
        public static string WebsiteHost = DEFAULT_HOST;
#endif

        public static string WebsiteUrl => $"{Scheme}://{WebsiteHost}/pilot";
        public static string UpdateZipUrl => $"{Scheme}://{WebsiteHost}/pilot/update.zip";
        public static string VersionUrl => $"{Scheme}://{WebsiteHost}/api/be/pilot/version";


        private static string _gameFolder = null;
        public static string GameFolder
        {
            get
            {
                if (_gameFolder == null)
                {
                    string installPath = null;

                    RegistryKey key = Registry.LocalMachine.OpenSubKey("SOFTWARE\\WOW6432Node\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\EscapeFromTarkov");
                    installPath = key?.GetValue("InstallLocation")?.ToString();
                    key?.Dispose();

                    if (string.IsNullOrEmpty(installPath))
                    {
                        installPath = "C:\\Battlestate Games\\EFT"; // default path
                    }

                    _gameFolder = installPath;
                }

                return _gameFolder;
            }

            set { _gameFolder = value; }
        }

        public static string LogsFolder
        {
            get
            {
                try
                {
                    return Path.Combine(GameFolder, "Logs");
                }
                catch
                {
                    return null;
                }
            }
        }

        private static string _screenshotsFolder;
        public static string ScreenshotsFolder
        {
            get
            {
                if (_screenshotsFolder == null)
                {
                    _screenshotsFolder = Path.Combine(Environment.GetFolderPath(Environment.SpecialFolder.MyDocuments), "Escape from Tarkov", "Screenshots");
                }
                return _screenshotsFolder;
            }
            set { _screenshotsFolder = value; }
        }

        //===================== AppContext Settings ============================

        public static void SetSettings(AppSettings settings)
        {
            Env.GameFolder = settings.gameFolder;
            Env.ScreenshotsFolder = settings.screenshotsFolder;
#if !DEBUG
            // host применяем только в RELEASE; в DEBUG не сбиваем localhost:3000
            if (!string.IsNullOrEmpty(settings.host) && Array.IndexOf(AllowedHosts, settings.host) >= 0)
            {
                Env.WebsiteHost = settings.host;
            }
#endif
        }

        public static AppSettings GetSettings()
        {
            AppSettings settings = new AppSettings()
            {
                gameFolder = Env.GameFolder,
                screenshotsFolder = Env.ScreenshotsFolder,
#if DEBUG
                // в DEBUG не персистим localhost:3000 в settings.json
                host = null,
#else
                host = Env.WebsiteHost,
#endif
            };
            return settings;
        }

        public static void ResetSettings()
        {
            AppSettings settings = new AppSettings()
            {
                gameFolder = null,
                screenshotsFolder = null,
                // host не сбрасываем — это контекст сайта, а не пользовательская настройка папок;
                // при нужде он перепишется следующим SETTINGS_INIT
                host = null,
            };
            SetSettings(settings);
        }

        // вызывается из обработчика SETTINGS_INIT; возвращает true, если host реально изменился (тогда стоит Settings.Save())
        public static bool SetHostFromSite(string host)
        {
#if DEBUG
            // в DEBUG разработка идёт против localhost:3000 — игнорируем
            return false;
#else
            if (string.IsNullOrEmpty(host)) return false;
            if (Array.IndexOf(AllowedHosts, host) < 0)
            {
                Logger.Log($"SETTINGS_INIT: host '{host}' rejected (not in whitelist)");
                return false;
            }
            if (Env.WebsiteHost == host) return false;
            Logger.Log($"Host updated: {Env.WebsiteHost} -> {host}");
            Env.WebsiteHost = host;
            return true;
#endif
        }

        //===================== AppContext Settings ============================

        public static void RestartApp()
        {
            Application.Restart();
            Environment.Exit(0);
        }
    }
}
