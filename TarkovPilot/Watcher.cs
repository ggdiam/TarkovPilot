﻿using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;
using Microsoft.Win32;

namespace TarkovPilot
{
    public static class Watcher
    {
        static readonly string ScreenshotRe = @"\d{4}-\d{2}-\d{2}\[\d{2}-\d{2}\]_(?<position>.+) \(\d\)\.png";
        static readonly string PositionRe = @"(?<y>-?[\d.]+), (?<z>-?[\d.]+), (?<x>-?[\d.]+)_.*";

        static FileSystemWatcher screenshotsWatcher;

        public static void Start()
        {
            if (!Directory.Exists(Env.ScreenshotsFolder))
            {
                Directory.CreateDirectory(Env.ScreenshotsFolder);
                Logger.Log($"Watcher: folder created: '{Env.ScreenshotsFolder}'");
            }

            screenshotsWatcher = new FileSystemWatcher(Env.ScreenshotsFolder)
            {
                EnableRaisingEvents = true,
            };
            screenshotsWatcher.Created += OnScreenshot;

            Logger.Log($"Watcher: started");
            Logger.Log($"Watcher: logs folder: '{Env.LogsFolder}'");
            Logger.Log($"Watcher: screenshots folder: '{Env.ScreenshotsFolder}'");
        }

        public static void Stop()
        {
            if (screenshotsWatcher != null)
            {
                screenshotsWatcher.Created -= OnScreenshot;

                screenshotsWatcher.EnableRaisingEvents = false;
                screenshotsWatcher.Dispose();

                screenshotsWatcher = null;

                Logger.Log($"Watcher: stopped");
            }
        }

        public static void Restart()
        {
            Stop();
            Start();
        }

        static void OnScreenshot(object sender, FileSystemEventArgs e)
        {
            try
            {
                string filename = e.Name ?? "";
                //Logger.Log($"Watcher:OnScreenshot {filename}");
                var match = Regex.Match(filename, ScreenshotRe);
                if (match.Success)
                {
                    var _position = match.Groups["position"].Value;
                    //Logger.Log($"Watcher:OnScreenshot position [{_position}]");
                    var posMatch = Regex.Match(_position, PositionRe);
                    if (posMatch.Success)
                    {
                        var position = new Position(posMatch.Groups["x"].Value, posMatch.Groups["y"].Value, posMatch.Groups["z"].Value);
                        //Logger.Log($"Watcher:OnScreenshot position {position}");
                        Server.SendPosition(position);
                    }
                }
            }
            catch (Exception ex)
            {
                Logger.Log($"Watcher:OnScreenshot err; {ex.Message}");
            }
        }
    }
}