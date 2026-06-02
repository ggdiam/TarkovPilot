using Newtonsoft.Json;
using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.RegularExpressions;

namespace TarkovPilot
{
    public static class LogsWatcher
    {
        // PVP / PVE Map change - application_000.log
        static readonly string LOCATION_SUBSTRING = "application|TRACE-NetworkGameCreate profileStatus";
        static readonly string LocationRe = @"location:\s*(?<loc>\S+),";

        // Task complete - push-notifications_000.log
        static readonly string TASK_SUBSTRING = "push-notifications|Got notification | ChatMessageReceived";
        static readonly string LINE_START_WITH_DATE = "^\\d{4}-\\d{2}-\\d{2} \\d{1,2}:\\d{1,2}:\\d{1,2}.\\d{3}";

        // PVE Map change - application_000.log
        static readonly string LOCATION_SUBSTRING2 = "application|scene preset";
        static readonly string LocationRe2 = @"path:maps\/(?<loc>\w+)\.bundle";

        static FileSystemWatcher logsFoldersWatcher;
        static LogFileWatcher appLogFileWatcher;
        static LogFileWatcher notifLogFileWatcher;
        static string curLogFolder;
        static Dictionary<string, long> filePositions = new Dictionary<string, long>();
        
        static int _initialLogsReadCount = 0;
        static bool IsAllInitialLogsRead
        {
            get
            {
                return _initialLogsReadCount == 2;
            }
        }

        static void SetInitialLogsReadDone()
        {
            if (!IsAllInitialLogsRead)
            {
                _initialLogsReadCount++;
            }
        }

        static void ResetInitialLogsReadDone()
        {
            _initialLogsReadCount = 0;
            filePositions.Clear();
        }

        public static void Start()
        {
            ResetInitialLogsReadDone();

            if (string.IsNullOrEmpty(Env.GameFolder))
            {
                Logger.Log($"Watcher: game folder not set");
                return;
            }
            if (!Directory.Exists(Env.GameFolder))
            {
                Logger.Log($"Watcher: game folder not found: '{Env.GameFolder}'");
                return;
            }
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

        static void MonitorLogFolder(string logsFolder)
        {
            // clear prev
            ClearLogsWatcher();

            // log file watcher
            appLogFileWatcher = new LogFileWatcher(logsFolder, "*application_*.log");
            appLogFileWatcher.Created += OnLogFileChanged;
            appLogFileWatcher.Changed += OnLogFileChanged;
            appLogFileWatcher.Start();

            // log file watcher
            notifLogFileWatcher = new LogFileWatcher(logsFolder, "*push-notifications_*.log");
            notifLogFileWatcher.Created += OnLogFileChanged;
            notifLogFileWatcher.Changed += OnLogFileChanged;
            notifLogFileWatcher.Start();

            Logger.Log($"Watcher: monitoring logs folder: '{logsFolder}'");
        }

        static void ClearLogsWatcher()
        {
            if (appLogFileWatcher != null)
            {
                appLogFileWatcher.Created -= OnLogFileChanged;
                appLogFileWatcher.Changed -= OnLogFileChanged;
                appLogFileWatcher.Stop();
                appLogFileWatcher = null;
            }

            if (notifLogFileWatcher != null)
            {
                notifLogFileWatcher.Created -= OnLogFileChanged;
                notifLogFileWatcher.Changed -= OnLogFileChanged;
                notifLogFileWatcher.Stop();
                notifLogFileWatcher = null;
            }
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
                            // initial read - skipping processing
                            if (!IsAllInitialLogsRead)
                            {
                                continue;
                            }

                            if (line.Contains(LOCATION_SUBSTRING))
                            {
                                // parsing raw location name
                                var map = ParseLoc(line, LocationRe);
                                if (!String.IsNullOrEmpty(map))
                                {
                                    // sending raw location name
                                    Server.SendMap(map);
                                }
                            }
                            else if (line.Contains(LOCATION_SUBSTRING2))
                            {
                                // parsing raw location name
                                var map = ParseLoc(line, LocationRe2);
                                if (!String.IsNullOrEmpty(map))
                                {
                                    // sending raw location name
                                    Server.SendMap(map);
                                }
                            }
                            else if (line.Contains(TASK_SUBSTRING))
                            {
                                // reading json
                                StringBuilder jsonBuilder = new StringBuilder();

                                // reading next line (json first line)
                                line = reader.ReadLine();

                                // while not EOF
                                while (line != null)
                                {
                                    // line - starts with date - new log record - exiting json parse
                                    var match = Regex.Match(line, LINE_START_WITH_DATE, RegexOptions.IgnoreCase);
                                    if (match.Success)
                                    {
                                        break;
                                    }

                                    jsonBuilder.AppendLine(line);
                                    // reading next line
                                    line = reader.ReadLine();
                                }

                                // parse JSON
                                try
                                {
                                    string jsonString = jsonBuilder.ToString();
                                    if (!string.IsNullOrEmpty(jsonString))
                                    {
                                        dynamic questRec = JsonConvert.DeserializeObject(jsonString);
                                        if (questRec != null && questRec.message != null
                                            && questRec.message.type != null
                                            && questRec.message.templateId != null)
                                        {
                                            string status = questRec.message.type.ToString();

                                            // "6574e0dedc0d635f633a5805 successMessageText"
                                            string templateId = questRec.message.templateId;
                                            string[] parts = templateId.Split(' ');
                                            if (parts.Length > 0)
                                            {
                                                var questId = parts[0];
                                                if (!string.IsNullOrEmpty(questId))
                                                {
                                                    Logger.Log($"Watcher: questId: {questId}, status: {status}");
                                                    Server.SendQuestUpdate(questId, status);
                                                }
                                            }
                                        }
                                    }

                                }
                                catch (System.Text.Json.JsonException jsonEx)
                                {
                                    Logger.Log($"Watcher:ProcessLogFile error processing JSON: {jsonEx.Message}");
                                }
                            }
                        }

                        // save read position
                        filePositions[filePath] = stream.Position;
                    }
                }
            }
            catch (Exception ex)
            {
                Logger.Log($"Watcher: error processing log file '{filePath}': {ex.Message}");
            }

            // initial read completed
            SetInitialLogsReadDone();
        }

        public static string ParseLoc(string line, string locationRe)
        {
            // line
            var match = Regex.Match(line, locationRe, RegexOptions.IgnoreCase);
            if (match.Success)
            {
                var map = match.Groups["loc"].Value.ToLower();
                return map;
            }

            return null;
        }
    }
}
