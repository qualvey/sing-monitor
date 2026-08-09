#!/bin/bash
# 用 v2ray-core stats command proto 探测 8080
cd /tmp/gc
curl -sL -o /tmp/v2ray_stats.proto https://raw.githubusercontent.com/v2fly/v2ray-core/v5.30.0/app/stats/command/command.proto
echo "===PROTO HEAD==="
head -20 /tmp/v2ray_stats.proto
echo "===TRY v2ray.core.app.stats.command.StatsService/GetStats==="
./grpcurl -plaintext -proto /tmp/v2ray_stats.proto -import-path /tmp 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/GetStats -d '{"name":"user>>>笑笑>>>traffic>>>uplink","reset":false}' 2>&1 | head -10
echo "===TRY QueryStats==="
./grpcurl -plaintext -proto /tmp/v2ray_stats.proto -import-path /tmp 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/QueryStats -d '{"pattern":"user>>>","reset":false}' 2>&1 | head -15
echo "===TRY experimental==="
./grpcurl -plaintext -proto /tmp/v2ray_stats.proto -import-path /tmp 127.0.0.1:8080 experimental.v2rayapi.StatsService/QueryStats -d '{"pattern":"user>>>","reset":false}' 2>&1 | head -5
