#!/bin/bash
echo "===SING-BOX VERSION==="
sing-box version 2>/dev/null | head -3
echo "===BIN PATH==="
which sing-box
echo "===STATS SERVICE STRINGS==="
BIN=$(which sing-box)
strings "$BIN" 2>/dev/null | grep -oE '[a-zA-Z0-9_.]+StatsService[^"]*' | sort -u | head -10
echo "===SING-BOX SERVICE==="
systemctl list-units --type=service 2>/dev/null | grep -i sing
cat /etc/systemd/system/sing-box.service 2>/dev/null | head -15
echo "===V2RAY API IN CONFIG==="
grep -A8 v2ray_api /etc/sing-box/config.json | head -12
