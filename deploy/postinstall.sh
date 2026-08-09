#!/bin/sh
# nfpm postinstall: register and start service (install/upgrade)
set -e

# data dir for sqlite db
if [ ! -d /var/lib/sing-monitor ]; then
  mkdir -p /var/lib/sing-monitor
fi

systemctl daemon-reload
systemctl enable sing-monitor.service
systemctl restart sing-monitor.service || true
