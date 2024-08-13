# TarkovPilot



TarkovPilot is an Escape from Tarkov companion application that can automatically upload your screenshots file names to https://tarkov-market.com and show your position on map.

## Features

- Maps
    - Show you position on map (from screenshots)
    - Map change (if possible to determine from logs)
- Automatic updates
    - Press button on website to update.  
      No need to download new version every time.

## UI

App have no UI, only tray icon.  
All configurations and logs you can find on TM website TarkovPilot's [page](https://tarkov-market.com/pilot)

<img src="https://github.com/ggdiam/TarkovPilot/blob/master/images/TarkovPilot%20page.png"/>

## Installation

- Latest version you can find on TM website TarkovPilot's [page](https://tarkov-market.com/pilot).
- Here on GitHub in [latest release](https://github.com/ggdiam/TarkovPilot/releases/tag/latest)

Downloaded, extract the zip and run the `TarkovPilot.exe` executable. Open [TM website](https://tarkov-market.com/pilot), and see it's connected.

## FAQ

### How does TarkovPilot work?

- TarkovPilot watches the log files that the game creates as it's running.  
  From some log messages possible to determine map, you are loading in.

- TarkovPilot watches the screenshot files that you make.  
  In every screenshot file name coded info about position, where it was created.  
  TarkovPilot just get this position and upload to TM website and show your position on map.

### Is TarkovPilot a cheat?

No.  
TarkovPilot just reading game logs, and your game screenshots.  
Thats all.  
There is no direct interaction with the game or game memory.  
Also BSG or BattleEye devs always can check source code here to be sure app is purely safe and doesn't break TOS.
