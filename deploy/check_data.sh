#!/bin/bash
export PGPASSWORD=passwd
echo "===COUNT NEW LOGS==="
psql -U singbox -h localhost -d singbox_test -c "select count(*) from traffic_logs where timestamp > now() - interval '60 seconds';"
echo "===SAMPLE==="
psql -U singbox -h localhost -d singbox_test -c "select category, target_name, uplink_delta, downlink_delta from traffic_logs order by id desc limit 6;"
echo "===TOTALS==="
psql -U singbox -h localhost -d singbox_test -c "select category, target_name, total_bytes from traffic_totals order by total_bytes desc limit 3;"
