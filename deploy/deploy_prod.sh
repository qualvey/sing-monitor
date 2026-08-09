#!/bin/bash
# sing-monitor v0.2.0 生产部署脚本（在 rjp 服务器上执行）
# 用法: bash deploy_prod.sh
set -e

echo "=== 1. 备份旧二进制 ==="
sudo cp /usr/local/bin/sing-monitor /usr/local/bin/sing-monitor.singbox-monitor.bak 2>/dev/null || true
echo "已备份到 /usr/local/bin/sing-monitor.singbox-monitor.bak"

echo "=== 2. 安装新二进制 ==="
# 从本机 scp 到 ~/singtest/sing-monitor-test 后执行：
sudo cp ~/singtest/sing-monitor-test /usr/local/bin/sing-monitor
sudo chmod 755 /usr/local/bin/sing-monitor

echo "=== 3. 备份现有配置 ==="
sudo cp /etc/sing-monitor/config.yaml /etc/sing-monitor/config.yaml.bak_v020 2>/dev/null || true

echo "=== 4. 重启服务 ==="
sudo systemctl restart v2rayapi.service
sleep 3
sudo systemctl status v2rayapi.service --no-pager | head -12

echo "=== 5. 验证 ==="
TOKEN=$(curl -s -X POST http://127.0.0.1:8090/api/v1/auth/login -H 'Content-Type: application/json' -d '{"password":"passwd0001"}' | python3 -c "import sys,json;print(json.load(sys.stdin).get('token',''))")
echo "登录 TOKEN_LEN=${#TOKEN}"
if [ -n "$TOKEN" ]; then
  curl -s http://127.0.0.1:8090/api/v1/users -H "Authorization: Bearer $TOKEN" | python3 -c "
import sys,json
users=json.load(sys.stdin)
print('用户数:', len(users))
for u in users[:3]:
    print(' ', u['email'], '| 周期', u.get('cycle_days'),'天 | 周期内流量', round(u.get('period_total_bytes',0)/1024**3,2),'GB')
"
  echo "前端: $(curl -s http://127.0.0.1:8090/ | head -c 60)"
else
  echo "登录失败！检查 /etc/sing-monitor/config.yaml 的 auth.password"
  sudo journalctl -u v2rayapi.service --no-pager | tail -15
  exit 1
fi

echo ""
echo "=== 部署完成 ==="
echo "访问 https://rjp.wowoha.top/control/ （浏览器可能需强刷 Ctrl+Shift+R）"
echo "回滚: sudo cp /usr/local/bin/sing-monitor.singbox-monitor.bak /usr/local/bin/sing-monitor && sudo systemctl restart v2rayapi.service"
