#!/bin/bash
TOKEN=$(curl -s -X POST http://127.0.0.1:8091/api/v1/auth/login -H 'Content-Type: application/json' -d '{"password":"testpass"}' | python3 -c 'import sys,json;print(json.load(sys.stdin).get("token",""))')
echo "TOKEN_LEN=${#TOKEN}"
if [ -z "$TOKEN" ]; then echo "LOGIN FAIL"; exit 1; fi
echo "===STATS==="
curl -s http://127.0.0.1:8091/api/v1/stats -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('count:', len(d))
for r in d[:3]: print(' ', {k: r.get(k) for k in ['category','target_name','uplink_bytes','downlink_bytes','total_bytes']})
"
echo "===INBOUNDS==="
curl -s http://127.0.0.1:8091/api/v1/inbounds -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('count:', len(d))
for r in d[:3]: print(' ', {k: r.get(k) for k in ['id','tag','type','listen_port','enable']})
"
echo "===USERS==="
curl -s http://127.0.0.1:8091/api/v1/users -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
d=json.load(sys.stdin)
print('count:', len(d))
r=d[0]
print(' ', {k: r.get(k) for k in ['email','cycle_days','period_total_bytes','used_traffic']})
"
