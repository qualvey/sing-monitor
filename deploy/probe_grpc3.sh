#!/bin/bash
cd /tmp/gc
echo "===TRY v2ray.core.app.stats.command.StatsService/QueryStats==="
./grpcurl -plaintext -proto /tmp/v2ray_stats_min.proto -d '{"pattern":"user>>>","reset":false}' 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/QueryStats 2>&1 | head -8
echo "===TRY experimental.v2rayapi.StatsService/QueryStats==="
./grpcurl -plaintext -proto /tmp/v2ray_stats_min.proto -d '{"pattern":"user>>>","reset":false}' 127.0.0.1:8080 experimental.v2rayapi.StatsService/QueryStats 2>&1 | head -8
