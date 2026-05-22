@echo off
rem DMXAPI 配置工具一键安装脚本 (Windows CMD)
rem 用法: curl -fsSL https://cnb.cool/dmxapi/opencode_dmxapi/-/git/raw/main/install.cmd -o "%TEMP%\install.cmd" ^&^& call "%TEMP%\install.cmd"

setlocal

set "PS1_URL=https://cnb.cool/dmxapi/opencode_dmxapi/-/git/raw/main/install.ps1"

where powershell >nul 2>nul
if errorlevel 1 (
    echo X 未找到 PowerShell，无法继续。请使用 Windows 10/11 自带的 PowerShell，或参考 README 手动下载二进制。
    exit /b 1
)

echo ==^> 正在通过 PowerShell 执行一键安装...
set "PS1_TMP=%TEMP%\dmxapi_install_%RANDOM%_%RANDOM%.ps1"
powershell -NoProfile -ExecutionPolicy Bypass -Command "Invoke-WebRequest -UseBasicParsing -Uri '%PS1_URL%' -OutFile '%PS1_TMP%'"
if errorlevel 1 (
    echo X 下载 install.ps1 失败
    if exist "%PS1_TMP%" del /q "%PS1_TMP%" >nul 2>nul
    exit /b 1
)
powershell -NoProfile -ExecutionPolicy Bypass -File "%PS1_TMP%"
set "EXITCODE=%ERRORLEVEL%"
del /q "%PS1_TMP%" >nul 2>nul

endlocal & exit /b %EXITCODE%
