#!/bin/bash
echo "===SING-DELIVER==="
systemctl cat sing-deliver.service 2>/dev/null | grep -E "ExecStart|WorkingDirectory"
echo "===SING-DELIVER PROC==="
ps aux | grep -E "sing-deliver|uvicorn" | grep -v grep | head -5
echo "===ALL LISTEN PORTS (user-visible)==="
ss -tulnp 2>/dev/null | grep -E "LISTEN" | head -30
echo "===FIND WHO OWNS 8080 via /proc==="
for pid in $(ls /proc | grep -E '^[0-9]+$'); do
  if ls -l /proc/$pid/fd 2>/dev/null | grep -q "socket:\[" ; then
    :
  fi
done
# 用 lsof 如果可用
if command -v lsof >/dev/null 2>&1; then
  lsof -i :8080 2>/dev/null || echo "(lsof needs sudo)"
else
  echo "(no lsof)"
fi
echo "===SING-BOX CONFIG v2ray listen==="
python3 -c "
import json,re
raw=open('/etc/sing-box/config.json').read()
raw=re.sub(r'^\s*//.*$','',raw,flags=re.M)
c=json.loads(raw)
print(c['experimental']['v2ray_api'])
"
