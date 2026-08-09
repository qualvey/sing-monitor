#!/bin/bash
echo "===SING-BOX 837357 LISTEN SOCKETS==="
# /proc/PID/net/tcp 第一列是本地地址（hex）
for pid in $(pgrep -f "sing-box.*run"); do
  echo "--- pid $pid ---"
  cat /proc/$pid/net/tcp 2>/dev/null | awk 'NR>1 {split($2,a,":"); port=strtonum("0x" a[2]); if ($4=="0A") print "  LISTEN 127.0.0.1:" port}'
  cat /proc/$pid/net/tcp6 2>/dev/null | awk 'NR>1 {split($2,a,":"); port=strtonum("0x" a[2]); if ($4=="0A") print "  LISTEN [::]:" port}'
done
echo "===CONNECT TO 8080 via grpcurl with explicit proto==="
cd /tmp/gc
# 尝试 v2ray.core 风格服务名（无需 proto，直接指定 full method 会因无 proto 失败，但先确认 reflection 已关）
./grpcurl -plaintext 127.0.0.1:8080 list 2>&1
echo "===TRY known service methods==="
# grpcurl 不带 proto 无法调；改用 python 快速 gRPC 探测
python3 - <<'EOF'
import socket
# 发送一个 gRPC 请求头到 8080，看响应（HTTP/2 prior knowledge）
s = socket.create_connection(("127.0.0.1", 8080), timeout=5)
# gRPC 调用帧：method POST /experimental.v2rayapi.StatsService/QueryStats
# 简化：直接发 HTTP/2 preface
s.sendall(b"PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n")
import time
time.sleep(0.5)
try:
    data = s.recv(1024)
    print("HTTP2 response:", data[:80])
except Exception as e:
    print("recv err:", e)
s.close()
EOF
echo "===SING-BOX CONFIG FILE TIMES==="
ls -la /etc/sing-box/*.json /etc/sing-box/*.bak 2>/dev/null | head -5
