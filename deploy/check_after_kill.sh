#!/bin/bash
# 密码文件方式（避免命令替换被破坏）
grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1 | awk '{print $2}' | tr -d '"' > ~/.pgpw
echo "pw len: $(wc -c < ~/.pgpw)"
export PGPASSWORD=$(cat ~/.pgpw)
echo "===PROD DB LAST 40s==="
psql -U singbox -h localhost -d singbox -c "select target_name, count(*) polls, sum(uplink_delta) up, sum(downlink_delta) down from traffic_logs where timestamp > now() - interval '40 seconds' group by target_name order by up+down desc;"
echo "===LAST 3 POLLS==="
psql -U singbox -h localhost -d singbox -c "select target_name, uplink_delta, downlink_delta, timestamp from traffic_logs order by timestamp desc limit 9;"
