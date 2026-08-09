#!/bin/sh
# nfpm preremove: stop and disable service before remove/upgrade
set -e

systemctl stop sing-monitor.service || true
systemctl disable sing-monitor.service || true
systemctl daemon-reload || true
