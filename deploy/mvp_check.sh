#!/bin/bash
echo "===1. PROD VERSION==="
/usr/bin/sing-monitor-server -version 2>&1 | head -1
echo "===2. SERVICE STATUS==="
systemctl is-active sing-monitor.service
echo "===3. COLLECTOR HEALTH (last 5)==="
journalctl -u sing-monitor.service --no-pager -n 8 2>/dev/null | grep -E "Collector|Main|error|Error" | tail -6
echo "===4. DB HEALTH==="
PG_LINE=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1)
PGPASSWORD=*** "$PG_LINE" | awk '{print $2}' | tr -d '"')
export PGPASSWORD
psql -U singbox -h localhost -d singbox -c "select (select count(*) from users) users, (select count(*) from inbound_nodes) inbounds, (select count(*) from traffic_logs) logs, (select count(*) from traffic_logs where timestamp > now() - interval '5 minutes') logs_5min, (select count(*) from users where cycle_start is not null) with_cycle;"
echo "===5. API CHECK==="
TOKEN=$(curl -s -X POST http://127.0.0.1:8090/api/v1/auth/login -H 'Content-Type: application/json' -d '{"password":"***"}' | python3 -c 'import sys,json;print(json.load(sys.stdin).get("token",""))')
echo "token len: ${#TOKEN}"
curl -s http://127.0.0.1:8090/api/v1/version -H "Authorization: Bearer $TOKEN"
echo ""
curl -s -o /dev/null -w "users api: %{http_code}\n" http://127.0.0.1:8090/api/v1/users -H "Authorization: Bearer $TOKEN"
curl -s -o /dev/null -w "frontend: %{http_code}\n" http://127.0.0.1:8090/
echo "===6. SING-BOX CHECK==="
ss -tulnp 2>/dev/null | grep -E ":8080|:8090" | head -4
