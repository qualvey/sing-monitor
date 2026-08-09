#!/bin/bash
# 测试实例验证 json tag 修复
pkill -f sing-monitor-te 2>/dev/null || true
pkill -f sing-monitor-fix 2>/dev/null || true
sleep 1
cp /home/user/singtest/sing-monitor-fix /home/user/singtest/sing-monitor-test
chmod +x /home/user/singtest/sing-monitor-test
cd /home/user/singtest && nohup ./sing-monitor-test -config config.test.yaml > /tmp/fix2.log 2>&1 &
sleep 4

AUTH_PWD=$(grep -A3 '^auth:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1 | awk '{print $2}' | tr -d '"')
TOKEN=$(curl -s -X POST http://127.0.0.1:8091/api/v1/auth/login -H 'Content-Type: application/json' -d "{\"password\":\"${AUTH_PWD}\"}" | python3 -c "import sys,json;print(json.load(sys.stdin).get('token',''))")
echo "TOKEN_LEN=${#TOKEN}"

echo "===STATS (should have snake_case fields)==="
curl -s http://127.0.0.1:8091/api/v1/stats -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('count:', len(d))
for r in d[:4]: print(' ', {k: r.get(k) for k in ['category','target_name','uplink_bytes','downlink_bytes','total_bytes']})
"
echo "===INBOUNDS==="
curl -s http://127.0.0.1:8091/api/v1/inbounds -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('count:', len(d))
for r in d[:3]: print(' ', {k: r.get(k) for k in ['id','tag','type','listen_port','enable','server_name']})
"
echo "===USERS (cycle fields)==="
curl -s http://127.0.0.1:8091/api/v1/users -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('count:', len(d))
r=d[0]
print(' ', {k: r.get(k) for k in ['email','cycle_days','period_total_bytes','used_traffic']})
"
