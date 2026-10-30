#!/bin/sh
# Start AgentSpillCheck and open the dashboard.
(spillcheck up &) 
sleep 1
command -v xdg-open >/dev/null 2>&1 && xdg-open http://127.0.0.1:7700/ || open http://127.0.0.1:7700/ || true
wait
