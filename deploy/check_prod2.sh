#!/bin/bash
# 从生产配置读取凭据，避免在脚本中硬编码
PGPASSWORD=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1 | awk '{print $2}' | tr -d '"')
export PGPASSWORD
AUTH_PWD=$(grep -A3 '^auth:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1 | awk '{print $2}' | tr -d '"')
PGHOST=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'host:' | head -1 | awk '{print $2}' | tr -d '"')
PGUSER=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'user:' | head -1 | awk '{print $2}' | tr -d '"')
PGDB=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'dbname:' | head -1 | awk '{print $2}' | tr -d '"')
echo "conn: $PGUSER@$PGHOST/$PGDB (pwd len ${#PGPASSWORD})"

echo "===1. DB COUNTS==="
psql -U "$PGUSER" -h "$PGHOST" -d "$PGDB" -c "select 'inbound_nodes' t, count(*) from inbound_nodes union all select 'traffic_totals', count(*) from traffic_totals union all select 'users', count(*) from users union all select 'traffic_logs', count(*) from traffic_logs;"

echo "===2. TOTALS SAMPLE==="
psql -U "$PGUSER" -h "$PGHOST" -d "$PGDB" -c "select id, category, target_name, uplink_bytes, downlink_bytes, total_bytes from traffic_totals order by id limit 8;"

echo "===3. INBOUNDS SAMPLE==="
psql -U "$PGUSER" -h "$PGHOST" -d "$PGDB" -c "select id, tag, type, listen_port, enable from inbound_nodes order by id limit 12;"

echo "===4. API STATS==="
TOKEN=$(curl -s -X POST http://127.0.0.1:8090/api/v1/auth/login -H 'Content-Type: application/json' -d "{\"password\":\"${AUTH_PWD}\"}" | python3 -c "import sys,json;print(json.load(sys.stdin).get('token',''))")
echo "TOKEN_LEN=${#TOKEN}"
curl -s http://127.0.0.1:8090/api/v1/stats -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('stats count:', len(d))
for r in d[:6]: print(' ', r.get('category'), r.get('target_name'), r.get('uplink_bytes'), r.get('downlink_bytes'), r.get('total_bytes'))
" 2>&1 | head -10

echo "===5. API INBOUNDS==="
curl -s http://127.0.0.1:8090/api/v1/inbounds -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('inbounds count:', len(d))
for r in d[:6]: print(' ', r.get('id'), r.get('tag'), r.get('type'), r.get('listen_port'), r.get('enable'))
" 2>&1 | head -10

echo "===6. LOG ERRORS==="
journalctl -u sing-monitor.service --no-pager -n 40 2>/dev/null | grep -iE "error|fail|panic|warn" | tail -8
