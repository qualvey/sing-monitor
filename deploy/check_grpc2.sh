#!/bin/bash
echo "===DOWNLOAD GRPCURL==="
curl -sL -o /tmp/grpcurl.tar.gz https://github.com/fullstorydev/grpcurl/releases/download/v1.9.3/grpcurl_1.9.3_linux_x86_64.tar.gz
ls -la /tmp/grpcurl.tar.gz
mkdir -p /tmp/gc && tar -xzf /tmp/grpcurl.tar.gz -C /tmp/gc
ls -la /tmp/gc/
echo "===LIST SERVICES==="
/tmp/gc/grpcurl -plaintext -connect-timeout 5 127.0.0.1:8080 list 2>&1 | head -20
