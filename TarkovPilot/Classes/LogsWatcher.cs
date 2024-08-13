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
        static FileSystemWatcher logsWatcher;
        static string curLogFolder;
        static Dictionary<string, long> filePositions = new Dictionary<string, long>();


        static Dictionary<string, string> LocationToMap = new Dictionary<string, string>
        {
            { "sandbox_high", MapName.Ground_Zero },
            { "factory_day", MapName.Factory },
            { "factory_night", MapName.Factory },
            { "factory4_day", MapName.Factory },
            { "factory4_night", MapName.Factory },
            { "bigmap", MapName.Customs },
            { "woods", MapName.Woods },
            { "shoreline", MapName.Shoreline },
            { "interchange", MapName.Interchange },
            { "shopping_mall", MapName.Interchange },
            { "rezervbase", MapName.Reserve },
            { "rezerv_base", MapName.Reserve },
            { "laboratory", MapName.The_Lab },
            { "lighthouse", MapName.Lighthouse },
            { "tarkovstreets", MapName.Streets_of_Tarkov },
            { "city", MapName.Streets_of_Tarkov },
        };

        public static void Start()
        {
            // newest log folder
            curLogFolder = GetLatestLogFolder();
            if (curLogFolder != null)
            {
                MonitorLogFolder(curLogFolder);
            }

            // lookig for new folders creation
            logsFoldersWatcher = new FileSystemWatcher(Env.LogsFolder)
            {
                NotifyFilter = NotifyFilters.DirectoryName,
                EnableRaisingEvents = true,
            };

            logsFoldersWatcher.Created += OnNewFolderCreated;
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
            if (logsWatcher != null)
            {
                logsWatcher.Created -= OnLogFileChanged;
                logsWatcher.Changed -= OnLogFileChanged;
                logsWatcher.Dispose();
                logsWatcher = null;
            }
        }

        static void MonitorLogFolder(string logsFolder)
        {
            // clear prev
            ClearLogsWatcher();

            // create new
            logsWatcher = new FileSystemWatcher(logsFolder)
            {
                NotifyFilter = NotifyFilters.LastWrite,
                Filter = "*application.log",
                EnableRaisingEvents = true,
            };
            logsWatcher.Created += OnLogFileChanged;
            logsWatcher.Changed += OnLogFileChanged;

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

        static void OnLogFileChanged(object sender, FileSystemEventArgs e)
        {
            try
            {
                ProcessLogFile(e.FullPath);
            }
            catch (Exception ex)
            {
                Logger.Log($"Watcher:OnLogFileChanged err; {ex.Message}");
            }
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
                            if (line.Contains(LOCATION_SUBSTRING))
                            {
                                var map = ParseMap(line);
                                Server.SendMap(map);
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
        }

        public static string ParseMap(string line)
        {
            // line
            // 2024-08-11 22:07:55.907 +03:00|0.14.9.7.31124|Debug|application|TRACE-NetworkGameCreate profileStatus: 'Profileid: 662b8d646901ea852700f0a1, Status: Busy, RaidMode: Online, Ip: 134.119.194.154, Port: 17046, Location: factory4_night, Sid: 134.119.194.154-17046_11.08.24_19.07.00, GameMode: deathmatch, shortId: 5M44SZ' 
            // 2024-08-12 20:17:30.017 +03:00|0.14.9.7.31124|Debug|application|TRACE-NetworkGameCreate profileStatus: 'Profileid: 662b8d646901ea852700f0a2, Status: Busy, RaidMode: Online, Ip: 134.119.204.18, Port: 17002, Location: factory4_day, Sid: 134.119.204.18-17002_12.08.24_17.16.41, GameMode: deathmatch, shortId: 5DHK36' 
            var match = Regex.Match(line, LocationRe, RegexOptions.IgnoreCase);
            if (match.Success)
            {
                var loc = match.Groups["loc"].Value.ToLower();
                if (LocationToMap.TryGetValue(loc, out string map))
                {
                    return map;
                }
            }

            return null;
        }
    }
}
