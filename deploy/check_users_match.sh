#!/bin/bash
echo "===CONFIG stats.users==="
python3 -c "
import json, re
raw = open('/etc/sing-box/config.json').read()
raw = re.sub(r'^\s*//.*$', '', raw, flags=re.M)
c = json.loads(raw)
users = c['experimental']['v2ray_api']['stats']['users']
print('stats.users count:', len(users))
print(users)
inbounds_users = set()
for ib in c.get('inbounds', []):
    for u in ib.get('users', []):
        inbounds_users.add(u.get('name'))
print('inbounds users count:', len(inbounds_users))
"
echo "===DB USERS==="
PG_LINE=$(grep -A6 '^postgres:' /etc/sing-monitor/config.yaml | grep 'password:' | head -1)
PGPASSWORD=*** "$PG_LINE" | awk '{print $2}' | tr -d '"')
export PGPASSWORD
psql -U singbox -h localhost -d singbox -c "select email, enable from users order by id;"
