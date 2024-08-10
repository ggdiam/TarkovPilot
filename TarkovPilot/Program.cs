using System;
using System.Collections.Generic;
using System.Linq;
using System.Threading.Tasks;
using System.Windows.Forms;
using System.Globalization;
using System.Diagnostics;
using System.Threading;

namespace TarkovPilot
{
    static class Program
    {
        /// <summary>
        /// The main entry point for the application.
        /// </summary>
        [STAThread]
        static void Main(string[] args)
        {
            bool createdNew;
            using (Mutex mutex = new Mutex(true, "TarkovPilotMutex", out createdNew))
            {
                if (createdNew)
                {
                    // First instance
                    StartApp(args);
                }
                else
                {
                    // Exit
                }
            }
        }

        static void StartApp(string[] args)
        {
            // after update logs
            Updater.CheckAfterUpdateLogic(args);

            // app start
            Logger.Log($"Tarkov Pilot started");

            Application.EnableVisualStyles();
            Application.SetCompatibleTextRenderingDefault(false);

            CreateTrayIcon();
            Settings.Load();
            Logger.Log($"Version: {Env.Version}");

            Server.Start();
            Watcher.Start();

            Application.Run();
        }

        static void CreateTrayIcon()
        {
            NotifyIcon trayIcon = new NotifyIcon
            {
                Icon = Resources.favicon,
                Visible = true,
                Text = "Tarkov Pilot"
            };

            ContextMenuStrip contextMenu = new ContextMenuStrip();
            contextMenu.Items.Add(Resources.Open, null, (s, e) =>
            {
                Process.Start(Env.WebsiteUrl);
            });
            contextMenu.Items.Add("Exit", null, (s, e) => Application.Exit());

            trayIcon.ContextMenuStrip = contextMenu;
        }
    }
}
