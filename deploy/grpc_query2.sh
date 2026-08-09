#!/bin/bash
cd /tmp/gc
echo "===QUERY 1==="
./grpcurl -plaintext -proto /tmp/v2ray_stats_min.proto -import-path /tmp -d '{"pattern":"user>>>","reset":true}' 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/QueryStats 2>&1 | grep -E '"name"|"value"' | head -30
sleep 12
echo "===QUERY 2 (12s later)==="
./grpcurl -plaintext -proto /tmp/v2ray_stats_min.proto -import-path /tmp -d '{"pattern":"user>>>","reset":true}' 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/QueryStats 2>&1 | grep -E '"name"|"value"' | head -30
