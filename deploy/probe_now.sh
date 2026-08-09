#!/bin/bash
echo "===1. CURRENT QueryStats (reset=true)==="
cd /tmp/gc
./grpcurl -plaintext -proto /tmp/v2ray_full.proto -import-path /tmp -d '{"pattern":"user>>>","reset":true}' 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/QueryStats 2>&1 | grep -B1 '"value"' | head -30
echo "===2. QueryStats AGAIN (12s later, non-reset view)==="
sleep 12
./grpcurl -plaintext -proto /tmp/v2ray_full.proto -import-path /tmp -d '{"pattern":"user>>>","reset":true}' 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/QueryStats 2>&1 | grep -B1 '"value"' | head -30
echo "===3. SING-BOX LOG: recent connections==="
journalctl -u sing-box.service --no-pager -n 40 2>/dev/null | grep -iE "user|connect|traffic|inbound" | tail -15
