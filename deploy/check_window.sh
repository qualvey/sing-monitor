#!/bin/bash
PG_LINE=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1)
PGPASSWORD=$(echo "$PG_LINE" | awk '{print $2}' | tr -d '"')
export PGPASSWORD
echo "===LAST 10 MIN POLLS PER USER (test db)==="
psql -U singbox -h localhost -d singbox_test -c "select target_name, count(*) polls, max(timestamp) last_seen, sum(uplink_delta+downlink_delta) total from traffic_logs where timestamp > now() - interval '10 minutes' group by target_name order by last_seen desc;"
echo "===MIN TRAFFIC PER POLL==="
psql -U singbox -h localhost -d singbox_test -c "select target_name, min(uplink_delta+downlink_delta) as min_bytes, max(uplink_delta+downlink_delta) as max_bytes from traffic_logs where timestamp > now() - interval '10 minutes' group by target_name;"
