#!/bin/bash
export PGPASSWORD=$(cat ~/.pgpw)
psql -U singbox -h localhost -d singbox -c "select (select count(*) from users) users, (select count(*) from inbound_nodes) inbounds, (select count(*) from traffic_logs) logs, (select count(*) from traffic_logs where timestamp > now() - interval '5 minutes') logs_5min, (select count(*) from users where cycle_start is not null) with_cycle;"
psql -U singbox -h localhost -d singbox -c "select target_name, count(*) polls from traffic_logs where timestamp > now() - interval '3 minutes' group by target_name order by polls desc limit 5;"
