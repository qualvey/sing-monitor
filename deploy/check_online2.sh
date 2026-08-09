#!/bin/bash
# 分步读取密码，避免命令替换被破坏
PG_LINE=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1)
PGPASSWORD=$(echo "$PG_LINE" | awk '{print $2}' | tr -d '"')
export PGPASSWORD
echo "pwd len: ${#PGPASSWORD}"
echo "===LAST 150s POLLS PER USER (test db)==="
psql -U singbox -h localhost -d singbox_test -c "select target_name, count(*) as polls, sum(uplink_delta) as up, sum(downlink_delta) as down from traffic_logs where timestamp > now() - interval '150 seconds' group by target_name order by count(*) desc;"
echo "===LAST 20s (latest poll)==="
psql -U singbox -h localhost -d singbox_test -c "select target_name, uplink_delta, downlink_delta, timestamp from traffic_logs where timestamp > now() - interval '20 seconds' order by timestamp desc limit 10;"
