#!/bin/bash
pkill -f sing-monitor-te 2>/dev/null
sleep 1
cp /home/user/singtest/sing-monitor-dbg /home/user/singtest/sing-monitor-test
chmod +x /home/user/singtest/sing-monitor-test
cd /home/user/singtest
setsid nohup ./sing-monitor-test -config config.test.yaml > /tmp/dbg.log 2>&1 < /dev/null &
sleep 30
grep -E 'RT\]|Main|Collector' /tmp/dbg.log | tail -30
