#!/bin/bash
# 验证修复版采集器（测试实例 8091，测试库）
pkill -f sing-monitor-te 2>/dev/null || true
pkill -f sing-monitor-fix 2>/dev/null || true
sleep 1
cp /home/user/singtest/sing-monitor-fix /home/user/singtest/sing-monitor-test
chmod +x /home/user/singtest/sing-monitor-test
export PGPASSWORD=***
psql -U singbox -h localhost -d postgres -c "DROP DATABASE IF EXISTS singbox_test;" 2>/dev/null
psql -U singbox -h localhost -d postgres -c "CREATE DATABASE singbox_test OWNER singbox;"
pg_dump -U singbox -h localhost -d singbox | psql -U singbox -h localhost -d singbox_test > /dev/null 2>&1

cd /home/user/singtest && nohup ./sing-monitor-test -config config.test.yaml > /tmp/fix-test.log 2>&1 &
sleep 15
echo "===LOG==="
grep -E "Collector|error|Error" /tmp/fix-test.log | head -15
echo "===TRAFFIC LOGS WRITTEN==="
psql -U singbox -h localhost -d singbox_test -c "select count(*) from traffic_logs where timestamp > now() - interval '30 seconds';"
psql -U singbox -h localhost -d singbox_test -c "select category, target_name, uplink_delta, downlink_delta from traffic_logs order by id desc limit 5;"
