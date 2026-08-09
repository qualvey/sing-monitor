#!/bin/bash
echo "===VERSION==="
head -3 /tmp/fix3.log
echo "===RECORDED LOG==="
grep -E 'recorded' /tmp/fix3.log | tail -3
echo "===LAST 150s TRAFFIC PER USER (test db)==="
export PGPASSWORD=*** -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1 | awk '{print $2}' | tr -d '"')
psql -U singbox -h localhost -d singbox_test -c "select target_name, count(*) polls, sum(uplink_delta) up, sum(downlink_delta) down from traffic_logs where timestamp > now() - interval '150 seconds' group by target_name order by count(*) desc;"
