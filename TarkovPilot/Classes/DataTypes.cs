namespace TarkovPilot
{
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
        public string version { get; set; }
        public override string ToString()
        {
            return $"gameFolder: '{gameFolder}' \nscreenshotsFolder: '{screenshotsFolder}' \nversion: '{version}'";
        }
    }

    public class UpdateSettingsData : AppSettings
    {
        public string messageType { get; set; }        
    }
}
