#!/bin/bash
echo "===WHO LISTENS 8080==="
ss -tulnp | grep 8080
echo "===SING-BOX SERVICE DETAIL==="
systemctl cat sing-box.service 2>/dev/null | grep -E "ExecStart|WorkingDirectory"
echo "===SING-BOX PROCESS==="
ps aux | grep -E "sing-box" | grep -v grep
echo "===GRPCURL LIST==="
if ! command -v grpcurl >/dev/null 2>&1; then
  curl -sL -o /tmp/grpcurl.tar.gz https://github.com/fullstorydev/grpcurl/releases/download/v1.9.3/grpcurl_1.9.3_linux_x86_64.tar.gz
  tar -xzf /tmp/grpcurl.tar.gz -C /tmp
fi
/tmp/grpcurl -plaintext -connect-timeout 5 127.0.0.1:8080 list 2>&1 | head -20
