using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.RegularExpressions;
using System.Threading.Tasks;

namespace TarkovPilot
{
    public static class LogsWatcher
    {
        const string LOCATION_SUBSTRING = "application|TRACE-NetworkGameCreate profileStatus";
        static readonly string LocationRe = @"location:\s*(?<loc>\S+),";

        static FileSystemWatcher logsFoldersWatcher;
        static LogFileWatcher logFileWatcher;
        static string curLogFolder;
        static Dictionary<string, long> filePositions = new Dictionary<string, long>();

        public static void Start()
        {
            if (!Directory.Exists(Env.LogsFolder))
            {
                Logger.Log($"Watcher: logs folder not found: '{Env.LogsFolder}'");
                return;
            }

            // newest log folder
            curLogFolder = GetLatestLogFolder();
            if (curLogFolder != null)
            {
                MonitorLogFolder(curLogFolder);
            }

            // lookig for new folders creation
            logsFoldersWatcher = new FileSystemWatcher(Env.LogsFolder);
            logsFoldersWatcher.Created += OnNewFolderCreated;
            logsFoldersWatcher.EnableRaisingEvents = true;
        }

        public static void Stop()
        {
            ClearLogsFoldersWatcher();
            ClearLogsWatcher();
        }

        public static void Restart()
        {
            Stop();
            Start();
        }

        static void ClearLogsFoldersWatcher()
        {
            if (logsFoldersWatcher != null)
            {
                logsFoldersWatcher.Created -= OnNewFolderCreated;
                logsFoldersWatcher.Dispose();
                logsFoldersWatcher = null;
            }

            filePositions.Clear();
        }

        static void ClearLogsWatcher()
        {
            if (logFileWatcher != null)
            {
                logFileWatcher.Created -= OnLogFileChanged;
                logFileWatcher.Changed -= OnLogFileChanged;
                logFileWatcher.Stop();
                logFileWatcher = null;
            }
        }

        static void MonitorLogFolder(string logsFolder)
        {
            // clear prev
            ClearLogsWatcher();

            // log file watcher
            logFileWatcher = new LogFileWatcher(logsFolder, "*application.log");
            logFileWatcher.Created += OnLogFileChanged;
            logFileWatcher.Changed += OnLogFileChanged;
            logFileWatcher.Start();

            Logger.Log($"Watcher: monitoring logs folder: '{logsFolder}'");
        }

        static void OnNewFolderCreated(object sender, FileSystemEventArgs e)
        {
            // check new folder - newest
            var newDirectory = e.FullPath;
            if (Directory.GetCreationTime(newDirectory) > Directory.GetCreationTime(curLogFolder))
            {
                curLogFolder = newDirectory;
                // monitor new folder
                MonitorLogFolder(curLogFolder);
            }
        }

        static string GetLatestLogFolder()
        {
            var directories = Directory.GetDirectories(Env.LogsFolder);
            if (directories.Length == 0)
                return null;

            // sort by create date
            var latestDirectory = directories.OrderByDescending(d => Directory.GetCreationTime(d)).FirstOrDefault();
            return latestDirectory;
        }

        static void OnLogFileChanged(object sender, FileChangedEventArgs e)
        {
            ProcessLogFile(e.FullPath);
        }

        static void ProcessLogFile(string filePath)
        {
            try
            {
                // map
                string map = null;

                // last read position
                long lastPosition = 0;
                if (filePositions.ContainsKey(filePath))
                {
                    lastPosition = filePositions[filePath];
                }

                //Logger.Log($"Watcher:ProcessLogFile processing '{filePath}' position: {lastPosition}");

                using (var stream = new FileStream(filePath, FileMode.Open, FileAccess.Read, FileShare.ReadWrite))
                {
                    stream.Seek(lastPosition, SeekOrigin.Begin);

                    using (var reader = new StreamReader(stream))
                    {
                        string line;
                        while ((line = reader.ReadLine()) != null)
                        {
                            if (line.Contains(LOCATION_SUBSTRING))
                            {
                                var _map = ParseMap(line);
                                if (!String.IsNullOrEmpty(_map))
                                {
                                    map = _map;
                                }
                            }
                        }

                        // save read position
                        filePositions[filePath] = stream.Position;
                    }
                }

                // skip map change events on first log file read at app start
                if (!Env.InitialLogsRead && !String.IsNullOrEmpty(map))
                {
                    Server.SendMap(map);
                }
            }
            catch (Exception ex)
            {
                Logger.Log($"Watcher: error processing log file '{filePath}': {ex.Message}");
            }

            // initial read completed
            Env.InitialLogsRead = false;
        }

        public static string ParseMap(string line)
        {
            // line
            var match = Regex.Match(line, LocationRe, RegexOptions.IgnoreCase);
            if (match.Success)
            {
                var loc = match.Groups["loc"].Value.ToLower();
                if (Dict.LocationToMap.TryGetValue(loc, out string map))
                {
                    return map;
                }
            }

            return null;
        }
    }
}
