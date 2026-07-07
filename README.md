# sing-monitor

sing-monitor 是一个为 [sing-box](https://github.com/SagerNet/sing-box) 设计的外围监控系统后端服务。它通过 sing-box 提供的 gRPC API 定期拉取用户流量信息以及系统运行状态，并将其持久化到本地 SQLite 数据库中，同时对外提供 RESTful API 供前端展示控制面板。

## 功能特性

- 📊 **流量监控**：自动识别 `sing-box` 统计数据中的用户（通过 `tag` 解析），记录上下行流量。
- 🖥️ **系统状态监控**：定时采集系统的内存占用、Goroutines 数量、Uptime 等关键指标。
- 💾 **轻量级持久化**：内置 GORM + SQLite，零配置即可实现数据的持久化存储，方便进行时序数据分析。
- ⚙️ **外部配置化**：通过 `config.json` 灵活配置监听端口、定时任务频率以及 gRPC 地址。
- 🌐 **RESTful API**：提供了标准的 JSON API，可轻松对接到 Vue/React 等前端框架。

## 前置要求

- [Go](https://golang.org/dl/) 1.20 或以上版本
- 运行中的 `sing-box` 实例（必须在配置文件中开启 experimental v2rayapi 相关的 gRPC 服务）

## 快速开始

### 1. 克隆代码

```bash
git clone git@github.com:qualvey/sing-monitor.git
cd sing-monitor/server
```

### 2. 准备配置文件

根据项目提供的示例配置创建你自己的配置文件：

```bash
cp config.example.json config.json
```

编辑 `config.json` 根据你的实际环境修改参数：
```json
{
  "api_server_port": 8080,
  "sing_box_grpc_addr": "127.0.0.1:10000",
  "collect_interval_seconds": 300,
  "db_path": "sing-monitor.db"
}
```
- `api_server_port`：本服务对前端暴露的 HTTP API 端口
- `sing_box_grpc_addr`：你的 sing-box gRPC 监听地址
- `collect_interval_seconds`：定时拉取数据的间隔时间（单位：秒）
- `db_path`：SQLite 数据库文件的存储路径

### 3. 运行服务

下载依赖并启动服务：
```bash
go mod tidy
go run .
```

启动成功后，服务端会自动根据配置间隔时间通过 gRPC 采集 sing-box 的数据。

## API 文档

监控系统默认监听指定的 `api_server_port`，所有接口返回均为 JSON 格式，且默认允许跨域 (CORS) 请求。

### 获取所有用户流量汇总
**`GET /api/users`**
返回通过解析 gRPC 流量 Tag 获取到的用户列表信息。

### 获取流量趋势记录
**`GET /api/traffic/trend`**
查询具体时间点的流量增量日志，默认倒序返回近期 100 条。
- `user_id` (可选): 指定查询特定用户的流量。

### 获取系统运行状态
**`GET /api/sys/stats`**
查询近期的系统资源占用状况（用于系统大盘图表）。

## 目录结构
```text
sing-monitor/
├── server/                 # Go 后端源码
│   ├── api/                # Gin REST API 路由与控制器
│   ├── collector/          # 核心定时任务，连接 gRPC 拉取数据
│   ├── config/             # 外部配置解析逻辑
│   ├── db/                 # GORM 数据库初始化
│   ├── models/             # 数据表结构定义 (User, TrafficLog, SysStatLog)
│   ├── config.example.json # 配置文件范例
│   └── main.go             # 程序入口
└── .gitignore              # git 忽略配置
```

## 协议
本项目作为独立开发者的个人工程，供学习与自用监控参考。
