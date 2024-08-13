using System;
using System.Collections.Generic;
using System.Diagnostics;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading;
using System.Threading.Tasks;

namespace TarkovPilot
{
    public class LogFileWatcher
    {
        readonly string folder;
        readonly string searchPattern;
        readonly int checkInterval;

        volatile bool isStopping = false;
        long lastFileSize = 0;
        FileSystemWatcher fileCreateWatcher;

        public event EventHandler<FileChangedEventArgs> Created;
        public event EventHandler<FileChangedEventArgs> Changed;

        public LogFileWatcher(string folder, string searchPattern, int checkInterval = 5000)
        {
            this.folder = folder;
            this.searchPattern = searchPattern;
            this.checkInterval = checkInterval;
        }

        string TryGetFilePath()
        {
            string[] files = Directory.GetFiles(folder, searchPattern);
            if (files.Length > 0)
            {
                return Path.Combine(folder, files[0]);
            }

            return null;
        }

        public void Start()
        {
            Reset();

            var filePath = TryGetFilePath();

            // if file exists - start monitoring changes
            if (!String.IsNullOrEmpty(filePath))
            {
                //Logger.Log($"LogFileWatcher: StartFileChangeMonitoring");
                StartFileChangeMonitoring(filePath);
            }
            else
            {
                //Logger.Log($"LogFileWatcher: Start fileCreateWatcher");
                // waiting for file creation
                fileCreateWatcher = new FileSystemWatcher(folder, searchPattern);
                fileCreateWatcher.Created += OnLogFileCreated;
                fileCreateWatcher.Renamed += OnLogFileCreated;
                fileCreateWatcher.EnableRaisingEvents = true;
            }
        }

        void StartFileChangeMonitoring(string filePath)
        {
            Task.Run(() => CheckFile(filePath));
        }

        void OnLogFileCreated(object sender, FileSystemEventArgs e)
        {
            //Logger.Log($"LogFileWatcher: log file created {e.Name}");
            // monitoring changes
            StartFileChangeMonitoring(e.FullPath);

            // file create monitoring stop
            StopFileCreationMonitoring();

            // trigger created
            Created?.Invoke(this, new FileChangedEventArgs(e.FullPath));
        }

        void CheckFile(string filePath)
        {
            while (!isStopping)
            {
                try
                {
                    // check file size
                    FileInfo fileInfo = new FileInfo(filePath);
                    long currentFileSize = fileInfo.Length;

                    //Logger.Log($"LogFileWatcher: check {Path.GetFileName(filePath)}, size: {currentFileSize}");
                    if (currentFileSize > lastFileSize)
                    {
                        lastFileSize = currentFileSize;
                        // trigger change
                        Changed?.Invoke(this, new FileChangedEventArgs(filePath));
                        //Logger.Log($"LogFileWatcher: changed {Path.GetFileName(filePath)}");
                    }
                }
                catch (Exception ex)
                {
                    Logger.Log($"LogFileWatcher: CheckFile err: {ex.Message}");

                    // exit - stop check loop
                    return;
                }

                // wait
                Thread.Sleep(checkInterval);
            }
        }

        public void Stop()
        {
            // stop changes monitoring
            isStopping = true;

            StopFileCreationMonitoring();
        }

        void StopFileCreationMonitoring()
        {
            if (fileCreateWatcher != null)
            {
                fileCreateWatcher.Created -= OnLogFileCreated;
                fileCreateWatcher.Dispose();
                fileCreateWatcher = null;
            }
        }

        void Reset()
        {
            isStopping = false;
            lastFileSize = 0;
        }
    }
}
