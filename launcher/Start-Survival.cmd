@echo off
setlocal
cd /d "%~dp0"
if errorlevel 1 (
    echo De launchermap kon niet worden geopend.
    pause
    exit /b 1
)
"%~dp0SurvivalReimagined.exe"
if errorlevel 1 (
    echo De launcher kon niet normaal worden gestart.
    pause
)
