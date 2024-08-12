using System;
using System.IO;
using System.IO.Compression;
using System.Net.Http;
using System.Threading.Tasks;
using System.Diagnostics;
using System.Windows.Forms;

namespace TarkovPilot
{
    public class Updater
    {
#if DEBUG
        const string UPDATE_URL = "http://localhost:3000/pilot/update.zip";
        const string VERSION_URL = "http://localhost:3000/api/be/pilot/version";
#else
        const string UPDATE_URL = "https://tarkov-market.com/pilot/update.zip";
        const string VERSION_URL = "https://tarkov-market.com/api/be/pilot/version";
#endif

        const string UPDATE_FILE_NAME = "update.zip";
        const string UPDATE_FOLDER = "update";

        static readonly string zipPath = Path.Combine(Directory.GetCurrentDirectory(), UPDATE_FILE_NAME);
        static readonly string extractPath = Path.Combine(Directory.GetCurrentDirectory(), UPDATE_FOLDER);


        public static void CheckUpdate()
        {
            Task.Run(() => CheckUpdateTask());
        }

        static async Task CheckUpdateTask()
        {
            Logger.Log("Updater: Checking for update...");
            string newVersion = await IsUpdateAvailable();
            if (!String.IsNullOrEmpty(newVersion))
            {
                Logger.Log($"Updater: New version {newVersion} found. Downloading update...");
                await Updater.DownloadUpdate();
                Logger.Log("Updater: Update downloaded");

                Logger.Log("Updater: Installing update...");
                InstallUpdate();
            }
            else
            {
                Logger.Log("Updater: You have latest version.");
            }
        }

        static async Task<string> IsUpdateAvailable()
        {
            try
            {
                using (HttpClient client = new HttpClient())
                {
                    string latestVersionString = await client.GetStringAsync(VERSION_URL);
                    latestVersionString = latestVersionString.Trim();
                    Version latestVersion = new Version(latestVersionString);
                    Version currentVersion = new Version(Env.Version);

                    if (latestVersion > currentVersion)
                    {
                        return latestVersionString;
                    }
                    else
                    {
                        return null;
                    }
                }
            }
            catch (Exception ex)
            {
                Logger.Log($"Updater: update check failed; {ex.Message}");
                return null;
            }
        }

        static async Task DownloadUpdate()
        {
            try
            {
                using (HttpClient client = new HttpClient())
                {
                    byte[] data = await client.GetByteArrayAsync(UPDATE_URL);

                    File.WriteAllBytes(UPDATE_FILE_NAME, data);
                }
            }
            catch (Exception ex)
            {
                Logger.Log($"Updater: Update download failed; {ex.Message}");
            }
        }

        static void InstallUpdate()
        {
            try
            {
                if (Directory.Exists(extractPath))
                {
                    Directory.Delete(extractPath, true);
                }

                // unpack zip
                ZipFile.ExtractToDirectory(zipPath, extractPath);


                // apply update
                ProcessStartInfo processInfo = new ProcessStartInfo("update.bat")
                {
                    UseShellExecute = true,
                    CreateNoWindow = true,
                    WorkingDirectory = AppDomain.CurrentDomain.BaseDirectory,
                };

                Process.Start(processInfo);

                // exiting
                Application.Exit();
            }
            catch (Exception ex)
            {
                // Обработка ошибок
                Logger.Log($"Updater: update install failed; {ex.Message}");
            }
        }

        public static void CheckAfterUpdateLogic(string[] args)
        {
            if (args.Length > 0)
            {
                string updArg = args[0];
                if (updArg == "updated")
                {
                    Logger.Log($"Updater: update installed succesfully");
                    CleanUp();
                }
            }
        }

        static void CleanUp()
        {
            try
            {
                if (Directory.Exists(extractPath))
                {
                    Directory.Delete(extractPath, true);
                }
                if (File.Exists(zipPath))
                {
                    File.Delete(zipPath);
                }
            }
            catch (Exception) { };
        }
    }
}
