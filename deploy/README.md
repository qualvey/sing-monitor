# sing-monitor systemd 服务（二进制部署，推荐）

Go 静态二进制 + systemd 是单服务部署的标准姿势，比 Docker 少一层抽象、启动更快、资源占用更小。

## 部署步骤

### 1. 下载二进制

从 GitHub Release 下载对应平台产物：

```bash
# 服务器是 linux/amd64 或 linux/arm64
curl -L -o sing-monitor-server \
  https://github.com/qualvey/sing-monitor/releases/download/v0.1.0/sing-monitor-server_0.1.0_linux_amd64
chmod +x sing-monitor-server
```

### 2. 安装到系统目录

```bash
sudo mkdir -p /opt/sing-monitor
sudo mv sing-monitor-server /opt/sing-monitor/
sudo cp server/config.example.json /opt/sing-monitor/config.json
# 编辑 config.json：改 sing_box_grpc_addr、collect_interval_seconds 等
```

### 3. 创建 systemd 服务

```bash
sudo cp deploy/sing-monitor.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now sing-monitor
```

### 4. 查看状态与日志

```bash
systemctl status sing-monitor
journalctl -u sing-monitor -f
```

## 升级

替换二进制 + 重启即可：

```bash
sudo systemctl stop sing-monitor
sudo curl -L -o /opt/sing-monitor/sing-monitor-server <新版本 URL>
sudo systemctl start sing-monitor
```

## 卸载

```bash
sudo systemctl disable --now sing-monitor
sudo rm /etc/systemd/system/sing-monitor.service /opt/sing-monitor -rf
```
