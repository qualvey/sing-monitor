#!/bin/bash
pkill -f sing-monitor-te 2>/dev/null
sleep 1
cd /home/user/singtest
setsid nohup ./sing-monitor-test -config config.test.yaml > /tmp/fix3.log 2>&1 < /dev/null &
sleep 15
echo "===LOG==="
tail -4 /tmp/fix3.log
echo "===WS CHECK==="
python3 /tmp/ws_check.py
