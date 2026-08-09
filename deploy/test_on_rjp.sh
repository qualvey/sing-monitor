#!/bin/bash
# 在 rjp 上准备测试库并启动测试实例（无需 sudo，全部在 /tmp）
set -e
export PGPASSWORD=passwd

echo "===CREATE TEST DB==="
psql -U singbox -h localhost -d postgres -c "DROP DATABASE IF EXISTS singbox_test;" 2>/dev/null || true
psql -U singbox -h localhost -d postgres -c "CREATE DATABASE singbox_test OWNER singbox;"
pg_dump -U singbox -h localhost -d singbox | psql -U singbox -h localhost -d singbox_test > /dev/null

echo "===START TEST INSTANCE==="
pkill -f singbox-monitor-test 2>/dev/null || true
sleep 1
cd /tmp && nohup /tmp/sing-monitor-test -config /tmp/config.test.yaml > /tmp/sing-monitor-test.log 2>&1 &
sleep 3

echo "===LOG==="
tail -15 /tmp/sing-monitor-test.log

echo "===SMOKE TEST==="
TOKEN=$(curl -s -X POST http://127.0.0.1:8091/api/v1/auth/login -H 'Content-Type: application/json' -d '{"password":"testpass"}' | python3 -c "import sys,json;print(json.load(sys.stdin).get('token',''))")
echo "TOKEN_LEN=${#TOKEN}"
if [ -z "$TOKEN" ]; then echo "LOGIN FAILED"; exit 1; fi

echo "--- users ---"
curl -s http://127.0.0.1:8091/api/v1/users -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
users=json.load(sys.stdin)
print('total users:', len(users))
for u in users[:3]:
    print(json.dumps({k:u.get(k) for k in ['id','email','enable','cycle_days','cycle_start','cycle_end','period_total_bytes','used_traffic']}, ensure_ascii=False))
"

echo "--- stats/users (period) ---"
curl -s http://127.0.0.1:8091/api/v1/stats/users -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('period rows:', len(d))
for r in d[:3]:
    print(json.dumps(r, ensure_ascii=False))
"

echo "--- inbounds ---"
curl -s http://127.0.0.1:8091/api/v1/inbounds -H "Authorization: Bearer $TOKEN" | python3 -c "import sys,json;print('inbounds:', len(json.load(sys.stdin)))"

echo "--- stats totals ---"
curl -s http://127.0.0.1:8091/api/v1/stats -H "Authorization: Bearer $TOKEN" | python3 -c "import sys,json;d=json.load(sys.stdin);print('totals:', len(d))"

echo "--- frontend ---"
curl -s http://127.0.0.1:8091/ | head -c 200
echo ""
echo "DONE"
