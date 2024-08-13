using System;
using System.Diagnostics;
using System.IO;

namespace TarkovPilot
{
    public class Logger
    {
        const string LOG_FILE_PATH = "app.log";
        static LimitedConcurrentQueue<string> logBuffer = new LimitedConcurrentQueue<string>();

        public static void Log(string msg)
        {
            //WriteLogToFile(msg);

            // Sending buffer
            if (logBuffer.Count > 0)
            {
                SendLogBuffer();
            }

            // Sending log msg
            var isSent = (Server.CanSend && Server.SendLog(msg));
            if (!isSent)
            {
                logBuffer.Enqueue(msg);
            }
            Debug.WriteLine($"====== logBuffer ====== \n{logBuffer} \n========= end =========");
        }

        static void WriteLogToFile(string msg)
        {
            using (StreamWriter writer = new StreamWriter(LOG_FILE_PATH, true))
            {
                writer.WriteLine($"{DateTime.Now}: {msg}");
            }
        }

        public static void SendLogBuffer()
        {
            LimitedConcurrentQueue<string> notSent = new LimitedConcurrentQueue<string>();

            var logList = logBuffer.ToList();
            for (int i = 0; i < logList.Count; i++)
            {
                var msg = logList[i];
                if (Server.CanSend && Server.SendLog(msg))
                {
                    // done
                }
                else
                {
                    notSent.Enqueue(msg);
                }
            }

            logBuffer = notSent;
        }
    }
}
