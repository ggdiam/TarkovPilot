namespace TarkovPilot
{
    public class AppSettings
    {
        public string gameFolder { get; set; }
        public string screenshotsFolder { get; set; }
        // host сайта (tarkov-market.com / tarkov-market.ru); приходит от вкладки в SETTINGS_INIT
        public string host { get; set; }
        public override string ToString()
        {
            return $"gameFolder: '{gameFolder}' \nscreenshotsFolder: '{screenshotsFolder}' \nhost: '{host}'";
        }
    }

    public class MapChangeData : WsMessage
    {
        public string map { get; set; }

        public override string ToString()
        {
            return $"{map}";
        }
    }

    public class UpdatePositionData : WsMessage
    {
        public float x { get; set; }
        public float y { get; set; }
        public float z { get; set; }

        public override string ToString()
        {
            return $"x:{x} y:{y} z:{z}";
        }
    }

    public class SendFilenameData : WsMessage
    {
        public string filename { get; set; }

        public override string ToString()
        {
            return $"{filename}";
        }
    }

    public class QuestUpdateData : WsMessage
    {
        public string questId { get; set; }
        public string status { get; set; }

        public override string ToString()
        {
            return $"{questId} {status}";
        }
    }

    public class WsMessage
    {
        public string messageType { get; set; }
        public override string ToString()
        {
            return $"messageType: {messageType}";
        }
    }

    public class ConfigurationData : WsMessage
    {
        public string gameFolder { get; set; }
        public string screenshotsFolder { get; set; }
        public string gameFolderErr { get; set; }
        public string screenshotsFolderErr { get; set; }
        public string version { get; set; }
        public override string ToString()
        {
            return $"gameFolder: '{gameFolder}' \nscreenshotsFolder: '{screenshotsFolder}' \nversion: '{version}' \ngameFolderErr: '{gameFolderErr}' \nscreenshotsFolderErr: '{screenshotsFolderErr}'";
        }
    }

    public class UpdateSettingsData : AppSettings
    {
        public string messageType { get; set; }
    }

    public class SettingsInitData : WsMessage
    {
        public string host { get; set; }
    }
}
