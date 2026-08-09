#!/bin/bash
cd /tmp/gc
cp /tmp/v2ray_stats_min.proto /tmp/v2ray_stats_min.proto 2>/dev/null || true
echo "===TRY v2ray.core.app.stats.command.StatsService/QueryStats==="
./grpcurl -plaintext -proto /tmp/v2ray_stats_min.proto 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/QueryStats -d '{"pattern":"user>>>","reset":false}' 2>&1 | head -8
echo "===TRY experimental.v2rayapi.StatsService/QueryStats==="
./grpcurl -plaintext -proto /tmp/v2ray_stats_min.proto 127.0.0.1:8080 experimental.v2rayapi.StatsService/QueryStats -d '{"pattern":"user>>>","reset":false}' 2>&1 | head -8
