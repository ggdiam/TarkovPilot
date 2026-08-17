# TarkovPilot

TarkovPilot is an Escape from Tarkov companion application. It watches your game screenshots
and logs and sends events to https://tarkov-market.com to show your position on the
interactive map in real time.

## Features

- Maps
    - Show your position on map (from screenshots)
    - <img src="https://github.com/ggdiam/TarkovPilot/blob/master/images/pilot-woods.png"/>
    - Show your look direction (require **pro** status on TM website)
    - <img src="https://github.com/ggdiam/TarkovPilot/blob/master/images/pilot-woods-look.png"/>
    - Automatic map change (from game logs)
    - Automatic quest completion (from game logs)
- Settings window + tray icon
    - Connection key, folders, autostart with Windows, event log — everything visible in the app
- Built-in updates
    - The app checks for new versions and updates itself with one click

## How it works

The app sends events (screenshot file name, map change, quest update) directly to the
TM website backend over HTTPS, linked to your account by a **Connection key** from
https://tarkov-market.com/pilot. The website map picks them up instantly.

Source code lives in [`app/`](app/) (Go + Wails).

## Installation

- Download on TM website TarkovPilot [page](https://tarkov-market.com/pilot)
- Unzip and run `TarkovPilot.exe`, paste your Connection key into the app window

## FAQ

### How does TarkovPilot work?

- TarkovPilot watches the log files that the game creates as it's running.
  From some log messages it's possible to determine the map you are loading into
  and quest events.

- TarkovPilot watches the screenshot files that you make.
  Every screenshot file name contains the position where it was created.
  TarkovPilot sends this file name to the TM website which shows your position on map.

### Is TarkovPilot a cheat?

No.
TarkovPilot just reads game logs and your game screenshot file names.
That's all.
There is no interaction with the game process or game memory.
Also BSG or BattleEye devs can always check the source code here to be sure the app
is purely safe and doesn't break TOS.
