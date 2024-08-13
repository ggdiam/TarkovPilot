using System;

namespace TarkovPilot
{
    public class FileChangedEventArgs : EventArgs
    {
        public string FullPath { get; }
        public FileChangedEventArgs(string fullPath)
        {
            FullPath = fullPath;
        }
    }
}
