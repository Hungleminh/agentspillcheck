@echo off
REM AgentSpillCheck installer for Windows 11 (PLAN 16.2, ADR-0007).
REM Docker when Docker Desktop is running, otherwise the native binary.
REM No administrator rights required; SmartScreen is never disabled.
setlocal
set "MODE=auto"
if "%~1"=="/docker" set "MODE=docker"
if "%~1"=="/native" set "MODE=native"
set "INSTALL_DIR=%LOCALAPPDATA%\AgentSpillCheck"
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"

echo AgentSpillCheck installer / Trinh cai dat AgentSpillCheck
if "%MODE%"=="auto" (
  docker info >nul 2>&1
  if %ERRORLEVEL%==0 (set "MODE=docker") else (set "MODE=native")
)

if "%MODE%"=="docker" (
  echo Docker detected: using docker compose
  curl -fsSL https://raw.githubusercontent.com/Hungleminh/agentspillcheck/main/deploy/docker-compose.yml -o "%INSTALL_DIR%\docker-compose.yml"
  if %ERRORLEVEL% NEQ 0 exit /b 1
  pushd "%INSTALL_DIR%"
  docker compose up -d
  popd
  echo Dashboard: http://127.0.0.1:7700/
  exit /b 0
)

echo Installing the native binary / Dang cai binary native
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0install.ps1"
if %ERRORLEVEL% NEQ 0 exit /b 1
"%INSTALL_DIR%\bin\spillcheck.exe" doctor
echo Run / Chay: spillcheck demo
endlocal
