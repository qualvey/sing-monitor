#!/bin/bash
cd /tmp/gc
# 完整 proto（去掉 extensions import）
cat > /tmp/v2ray_full.proto <<'PROTO'
syntax = "proto3";
package v2ray.core.app.stats.command;
message GetStatsRequest { string name = 1; bool reset = 2; }
message Stat { string name = 1; int64 value = 2; }
message GetStatsResponse { Stat stat = 1; }
message QueryStatsRequest { string pattern = 1; bool reset = 2; }
message QueryStatsResponse { repeated Stat stat = 1; }
message SysStatsRequest {}
message SysStatsResponse {
  uint32 NumGoroutine = 1;
  uint64 NumGC = 2;
  uint64 Alloc = 3;
  uint64 TotalAlloc = 4;
  uint64 Sys = 5;
  uint64 Mallocs = 6;
  uint64 Frees = 7;
  uint64 LiveObjects = 8;
  uint64 PauseTotalNs = 9;
  uint32 Uptime = 10;
}
service StatsService {
  rpc GetStats(GetStatsRequest) returns (GetStatsResponse);
  rpc QueryStats(QueryStatsRequest) returns (QueryStatsResponse);
  rpc GetSysStats(SysStatsRequest) returns (SysStatsResponse);
}
PROTO
cp /tmp/v2ray_full.proto .
echo "===FULL QUERY (with values)==="
./grpcurl -plaintext -proto v2ray_full.proto -import-path . -d '{"pattern":"user>>>","reset":true}' 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/QueryStats 2>&1 | head -40
echo "===CURRENT WS SNAPSHOT==="
python3 /tmp/ws_check.py
