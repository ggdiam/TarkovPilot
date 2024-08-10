using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace TarkovPilot
{
    public class AppSettings
    {
        public string gameFolder { get; set; }
        public string screenshotsFolder { get; set; }
        public override string ToString()
        {
            return $"gameFolder: '{gameFolder}' \nscreenshotsFolder: '{screenshotsFolder}'";
        }
    }
}
