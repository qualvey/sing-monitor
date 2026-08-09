#!/bin/bash
export PGPASSWORD=***
echo "===1. PROD DB COUNTS==="
psql -U singbox -h localhost -d singbox -c "select 'inbound_nodes' t, count(*) from inbound_nodes union all select 'traffic_totals', count(*) from traffic_totals union all select 'users', count(*) from users union all select 'traffic_logs', count(*) from traffic_logs;"
echo "===2. TOTALS SAMPLE==="
psql -U singbox -h localhost -d singbox -c "select id, category, target_name, uplink_bytes, downlink_bytes, total_bytes, updated_at from traffic_totals order by id limit 8;"
echo "===3. INBOUNDS SAMPLE==="
psql -U singbox -h localhost -d singbox -c "select id, tag, type, listen_port, enable from inbound_nodes order by id limit 12;"
echo "===4. API STATS==="
TOKEN=*** -s -X POST http://127.0.0.1:8090/api/v1/auth/login -H 'Content-Type: application/json' -d '{"password":"***"}' | python3 -c "import sys,json;print(json.load(sys.stdin).get('token',''))")
curl -s http://127.0.0.1:8090/api/v1/stats -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('stats count:', len(d))
for r in d[:5]: print(' ', r.get('category'), r.get('target_name'), r.get('uplink_bytes'), r.get('downlink_bytes'), r.get('total_bytes'))
" 2>&1 | head -10
echo "===5. API INBOUNDS==="
curl -s http://127.0.0.1:8090/api/v1/inbounds -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('inbounds count:', len(d))
for r in d[:5]: print(' ', r.get('id'), r.get('tag'), r.get('type'), r.get('listen_port'), r.get('enable'))
" 2>&1 | head -10
echo "===6. SERVICE LOG ERRORS==="
sudo -n journalctl -u sing-monitor.service --no-pager -n 30 2>/dev/null | grep -iE "error|fail|panic" | tail -10 || journalctl -u sing-monitor.service --no-pager -n 30 2>/dev/null | tail -5
