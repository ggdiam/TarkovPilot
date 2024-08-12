using System;
using System.Linq;
using System.Threading;
using System.Threading.Tasks;
using System.Windows.Forms;
using System.Text.Json;
using Fleck;
using System.Collections.Concurrent;
using System.Diagnostics;
using System.Reflection;

namespace TarkovPilot
{
    static class Server
    {
        const string WS_URL = "ws://0.0.0.0:5123";

        static volatile bool isClosing = false;
        static WebSocketServer _server = null;
        static readonly ConcurrentDictionary<IWebSocketConnection, bool> _sockets = new ConcurrentDictionary<IWebSocketConnection, bool>();

        static Server()
        {
            Application.ApplicationExit += (object sender, EventArgs e) => Stop();
        }

        public static bool CanSend
        {
            get
            {
                return _sockets.Count > 0;
            }
        }

        public static void Stop()
        {
            isClosing = true;

            if (_server != null)
            {
                _server.Dispose();
                _server = null;
            }
        }

        public static async void Start()
        {
            isClosing = false;

            // Server start
            StartServer();

#if DEBUG
            var posTask = Task.Run(() => SendRandomPosition());
            await posTask;
#endif
        }

        static void StartServer()
        {
            FleckLog.Level = LogLevel.Debug;
            _server = new WebSocketServer(WS_URL);

            _server.Start(socket =>
            {
                socket.OnOpen = () =>
                {
                    _sockets.TryAdd(socket, true);
                    Logger.Log($"Client connected! count: {_sockets.Count}");

                    SendConfiguration();
                };
                socket.OnClose = () =>
                {
                    _sockets.TryRemove(socket, out _);
                    Logger.Log($"Client disconnected! count: {_sockets.Count}");
                };
                socket.OnMessage = (msg) =>
                {
                    //Logger.Log($"OnMessage: [{msg}]");
                    ProcessMessage(msg);
                    //socket.Send("Echo: " + msg);
                };
            });

            Logger.Log($"WebSocket Server started at {WS_URL}");
        }

        static void SendRandomPosition()
        {
            while (!isClosing)
            {
                Random rnd = new Random();

                var fields = typeof(MapName).GetFields(BindingFlags.Public | BindingFlags.Static)
                                   .Where(f => f.FieldType == typeof(string))
                                   .Select(f => (string)f.GetValue(null))
                                   .ToArray();
                var map = fields[rnd.Next(fields.Length)];
                
                // waiting, to be sure messages order
                SendMap(map);
                Thread.Sleep(2000);

                // lab 0,0 position fix
                if (map == MapName.The_Lab)
                {
                    SendPosition(new Position(rnd.Next(10) * 10 - 340, rnd.Next(10) * 10 - 200, rnd.Next(10) * 10));
                }
                else
                {
                    SendPosition(new Position(rnd.Next(10) * 10, rnd.Next(10) * 10, rnd.Next(10) * 10));
                }

                Thread.Sleep(5000);
            }
        }

        static void SendData(Object data)
        {
            try
            {
                if (_sockets.Count == 0) return;

                var json = JsonSerializer.Serialize(data);

                // Send message to all connected clients
                foreach (var socket in _sockets.Keys.ToList().AsReadOnly())
                {
                    socket.Send(json);
                }
            }
            catch (Exception ex)
            {
                Debug.WriteLine("Server: SendData err", ex.Message);
            }
        }

        public static bool SendLog(string msg)
        {
            try
            {
                if (_sockets.Count == 0) return false;

                // Send message to all connected clients
                foreach (var socket in _sockets.Keys.ToList().AsReadOnly())
                {
                    socket.Send(msg);
                }

                return true;
            }
            catch (Exception ex)
            {
                Debug.WriteLine("Server: SendLog err", ex.Message);
                return false;
            }
        }

        public static void SendMap(string map)
        {
            MapChangeData data = new MapChangeData()
            {
                messageType = WsMessageType.MAP_CHANGE,
                map = map,
            };

            Logger.Log($"MapChange: {data}");
            SendData(data);
        }

        public static void SendPosition(Position pos)
        {
            UpdatePositionData posData = new UpdatePositionData()
            {
                messageType = WsMessageType.POSITION_UPDATE,
                x = pos.X,
                y = pos.Y,
                z = pos.Z,
            };

            Logger.Log($"UpdatePosition: {posData}");
            SendData(posData);
        }

        public static void SendConfiguration()
        {
            ConfigurationData data = new ConfigurationData()
            {
                messageType = WsMessageType.CONFIGURATION,
                version = Env.Version,
                gameFolder = Env.GameFolder,
                screenshotsFolder = Env.ScreenshotsFolder,
            };

            SendData(data);
        }

        static T ParseJson<T>(string json)
        {
            try
            {
                // Deserilize to object
                return JsonSerializer.Deserialize<T>(json);
            }
            catch (Exception) { }; // ignore
            return default(T);
        }

        static void ProcessMessage(string json)
        {
            WsMessage msg = ParseJson<WsMessage>(json);

            if (msg != null && msg.messageType == WsMessageType.SETTINGS_UPDATE)
            {
                var settings = ParseJson<UpdateSettingsData>(json);
                Logger.Log($"Settings set: \n{settings}");

                Env.SetSettings(settings, true);
                Settings.Save();
                SendConfiguration();

                Watcher.Restart();
                //Env.RestartApp();
            }
            else if (msg != null && msg.messageType == WsMessageType.SETTINGS_RESET)
            {
                Settings.Delete();
                Env.ResetSettings();
                SendConfiguration();

                Watcher.Restart();
                //Env.RestartApp();
            }
            else if (msg != null && msg.messageType == WsMessageType.UPDATE)
            {
                Updater.CheckUpdate();
            }
            else
            {
                SendLog("Echo: " + json);
            }
        }
    }
}
