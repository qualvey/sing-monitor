# sing-monitor deb 部署（Debian/Ubuntu，推荐）

## 安装

从 GitHub Release 下载 deb 包，一行安装：

```bash
sudo apt install ./sing-monitor_1.0.0_amd64.deb
```

安装完成服务已自动注册并启动。

- 二进制：`/usr/bin/sing-monitor-server`
- 配置：`/etc/sing-monitor/config.json`（首次安装后编辑，升级不覆盖）
- 数据：`/var/lib/sing-monitor/sing-monitor.db`

## 管理

```bash
systemctl status sing-monitor        # 状态
sudo systemctl restart sing-monitor  # 重启
sudo systemctl stop sing-monitor     # 停止
journalctl -u sing-monitor -f        # 日志
```

## 升级

```bash
sudo apt install ./sing-monitor_1.1.0_amd64.deb   # 覆盖安装即升级
```

升级时：服务自动重启，配置保留（`/etc/sing-monitor/config.json` 不会被覆盖），数据保留。

## 卸载

```bash
sudo apt remove sing-monitor    # 保留 /var/lib/sing-monitor 数据
sudo apt purge sing-monitor     # 连数据一起删除
```
