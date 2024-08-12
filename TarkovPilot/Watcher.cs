using System;
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
        public static void Start()
        {
            //Logger.Log($"Watcher: logs folder: '{Env.LogsFolder}'");
            //Logger.Log($"Watcher: screenshots folder: '{Env.ScreenshotsFolder}'");

            ScreenshotsWatcher.Start();
            LogsWatcher.Start();

            Logger.Log($"Watcher: started");            
        }

        public static void Stop()
        {
            ScreenshotsWatcher.Stop();
            LogsWatcher.Stop();
            Logger.Log($"Watcher: stopped");
        }

        public static void Restart()
        {
            ScreenshotsWatcher.Restart();
            LogsWatcher.Restart();
        }

        
    }
}
