#!/bin/bash
echo "===KILL TEST INSTANCES==="
pkill -f 'sing-monitor-test' 2>/dev/null
pkill -f 'sing-monitor-dbg' 2>/dev/null
sleep 1
echo "remaining sing processes:"
ps aux | grep -E 'sing-monitor' | grep -v grep
echo "===WAIT 25s FOR COLLECTOR==="
sleep 25
echo "===PROD DB LAST 40s==="
PG_LINE=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1)
PGPASSWORD=*** "$PG_LINE" | awk '{print $2}' | tr -d '"')
export PGPASSWORD
psql -U singbox -h localhost -d singbox -c "select target_name, count(*) polls, sum(uplink_delta) up, sum(downlink_delta) down from traffic_logs where timestamp > now() - interval '40 seconds' group by target_name order by up+down desc;"
