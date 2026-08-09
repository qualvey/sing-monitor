#!/bin/bash
echo "===PROD COLLECTOR LOG (last 15)==="
journalctl -u sing-monitor.service --no-pager -n 15 2>/dev/null | grep -E "Collector|RT|Main" | tail -15
echo "===PROD DB: LAST 5 MIN PER USER==="
PG_LINE=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1)
PGPASSWORD=*** "$PG_LINE" | awk '{print $2}' | tr -d '"')
export PGPASSWORD
psql -U singbox -h localhost -d singbox -c "select target_name, count(*) polls, sum(uplink_delta+downlink_delta) total from traffic_logs where timestamp > now() - interval '5 minutes' group by target_name order by total desc;"
echo "===LAST 3 POLLS RAW==="
psql -U singbox -h localhost -d singbox -c "select target_name, uplink_delta, downlink_delta, timestamp from traffic_logs order by timestamp desc limit 12;"
