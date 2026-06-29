@echo off
rem ============================================================
rem  DiskSizeScan - Windows release build
rem  Builds a clean amd64 binary and copies it to .\publish
rem ============================================================
cd /d "%~dp0"

echo ============================================
echo  Building DiskSizeScan for Windows (amd64)
echo ============================================

where wails >nul 2>nul || (echo [ERROR] wails CLI not found. Run: go install github.com/wailsapp/wails/v2/cmd/wails@latest & exit /b 1)

wails build -platform windows/amd64 -clean || (echo [ERROR] Build failed. & exit /b 1)

if not exist "publish" mkdir "publish"
copy /Y "build\bin\DiskSizeScan.exe" "publish\DiskSizeScan.exe" >nul || (echo [ERROR] Could not copy output binary. & exit /b 1)

echo.
echo [OK] Output: publish\DiskSizeScan.exe
