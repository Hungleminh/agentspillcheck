@echo off
REM Start AgentSpillCheck and open the dashboard.
start "" http://127.0.0.1:7700/
"%LOCALAPPDATA%\AgentSpillCheck\bin\spillcheck.exe" up
