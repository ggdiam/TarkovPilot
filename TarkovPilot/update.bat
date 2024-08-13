@echo off

echo waiting 1s for^ TarkovPilot close
timeout /t 1 /nobreak >nul

xcopy .\update\*.* . /s /e /h /y

start "" TarkovPilot.exe updated
echo TarkovPilot started