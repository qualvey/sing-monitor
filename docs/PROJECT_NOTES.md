# sing-monitor 项目开发笔记（v0.2.x 重建）

> 本文档面向后续开发者与 AI 助手，记录从 2026-08-09 开始的完整开发过程、关键决策、踩坑与修复。
> 目标：让任何人（或任何 AI）拿到仓库即可准确理解项目全貌，避免重复踩坑。

---

## 1. 项目背景与"真相"

### 1.1 两个同名项目的混淆（关键！）

- **`qualvey/sing-monitor`**（本仓库）：最初是纯后端监控（Go + gin + SQLite + `/api/users`，只有流量采集 + 3 个只读接口），**没有前端**。
- **`singbox-monitor`**（服务器上真实运行的系统）：用户用 AI 开发的独立项目，**源码已丢失**（rjp 服务器、Windows 本机、GitHub 公开仓库均找不到）。特征：
  - PostgreSQL + go:embed Vue3 前端（v0.2.0）+ `/api/v1/*` + JWT 登录 + 用户/入站节点管理 + WebSocket 实时监控
  - 二进制名 `singbox-monitor-linux-amd64`，19MB，Go 1.25.5，依赖 `v2rayapi (devel)`（用户 sing-box fork 的本地模块）
  - 已备份到 `C:\Users\Ryu\Documents\workspace\singbox-monitor.bin`（2026-08-09）

**识别证据链**（排查时用到的方法，可复用）：
1. `strings <二进制> | grep -iE "postgres|sqlite|api/v1"` → 看代码特征
2. `go version -m <二进制>` → buildinfo：module path、依赖、ldflags
3. 服务器 `~/.zsh_history` → 部署来源记录（`cp singbox-monitor-linux-amd64 /usr/local/bin/sing-monitor`）
4. `ss -tulnp` / `ps aux` → 确认实际运行进程

### 1.2 最终决策（用户拍板）

按 `singbox-monitor` 的功能**重新开发**，并加入周期流量功能：

1. **完整管理**：面板改用户/入站节点 → 自动修改 `/etc/sing-box/config.json`（只改用户和入站部分）→ `sing-box check` 校验 → 写回 → 热重载
2. **兼容现有数据库**：直接复用 PostgreSQL `singbox` 库的 5 张表（不重建），仅新增周期列
3. **周期流量**：每个用户独立计费周期（起始时间 + 天数），默认 30 天、默认从用户创建时间起算，可手动配置

---

## 2. 架构

### 2.1 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.25 + gin + gorm(postgres) + golang-jwt/v5 + gorilla/websocket + v2fly/v2ray-core v5（stats command 客户端） |
| 前端 | Vue3 + Vite + Tailwind + ECharts，`//go:embed all:web` 嵌入单二进制 |
| 数据库 | PostgreSQL（现有 singbox 库） |
| 部署 | 单二进制 + deb 包（nfpm）+ systemd |

### 2.2 目录结构

```
sing-monitor/
├── server/              Go 后端
│   ├── main.go          入口（-config/-version flag，注入 version）
│   ├── config/          config.yaml 解析（字段与原 singbox-monitor 完全兼容）
│   ├── models/          5 表模型 + 周期窗口计算（含 json tag）
│   ├── db/              PostgreSQL 连接 + AutoMigrate + 周期列回填
│   ├── api/             gin 路由/JWT/用户/入站/统计/系统操作/WebSocket
│   │   └── web/         go:embed 前端构建产物（git 忽略，CI 生成）
│   ├── collector/       sing-box gRPC 采集 → 增量日志 + 累计表
│   ├── control/         config.json 生成（备份→改→check→reload）+ 母配置导入
│   └── realtime/        WebSocket 实时速率/在线状态推送器
├── web/                 Vue3 前端源码（npm run build）
├── deploy/              systemd 单元、deb 脚本、部署/排查脚本
├── nfpm.yaml            deb 打包配置（占位符 __VERSION__/__ARCH__）
└── .github/workflows/   ci.yml（push）+ release.yml（tag）
```

### 2.3 数据库（兼容现有 schema）

5 张表（原样沿用）+ 新增列：

- `users`：email(unique)/uuid/password/flow/enable/traffic_limit/expire_at + **cycle_start/cycle_days（新增）**
- `inbound_nodes`：tag(unique)/type/listen/listen_port/enable/SNI/private_key/short_id 等 19 列
- `traffic_logs`：category/target_name/uplink_delta/downlink_delta/timestamp（10s 增量）
- `traffic_totals`：category+target_name(unique)/uplink_bytes/downlink_bytes/total_bytes（累计）
- `user_inbound_bindings`：user_id + inbound_id（多对多）

启动时 `AutoMigrate` 只加缺失列；存量用户自动回填 `cycle_start = created_at`、`cycle_days = 30`。

### 2.4 关键流程

**采集**：每 10s 查 sing-box gRPC（`QueryStats`，reset=true）→ 解析 `user>>>tag>>>traffic>>>uplink|downlink` → 增量写 `traffic_logs` + 累加 `traffic_totals` + 推送 realtime。

**配置生成**（用户/节点变更或点"一键重载"）：
1. 备份 `config.json` → `backup/config.json.bak_时间戳`
2. 解析（容忍行注释 JSONC）
3. **只更新 DB 中存在的入站节点**（按 tag 匹配）的 users 数组；DB 中 disable 的节点移除；**DB 没有的节点原样保留（绝不误删）**
4. 同步 `experimental.v2ray_api.stats.users`
5. 写临时文件 → `sing-box check` 通过才写回 → `reload_command`

**周期流量**：`CurrentCycleWindow(now)` = 锚点 + N×周期（自动滚动），大盘/用户列表按窗口 SUM 增量。

---

## 3. 版本演进

| 版本 | 内容 |
|---|---|
| v0.1.0/v0.1.1 | 初版周期流量功能 + deb 打包（后被重建替代，git 历史保留） |
| v0.2.0 | 按 singbox-monitor 功能完整重建（后端 + Vue3 前端 + 兼容库 + 配置生成） |
| v0.2.1 | 修复采集器 gRPC 服务名（见坑 1） |
| v0.2.2 | 修复 CI 测试（旧测试文件残留） |
| v0.2.3 | 修复 go:embed 子目录 + 新增版本接口 |
| v0.2.4 | vite base 顶层 + 模型 json tag + 在线判定 + 实时排序 + 图表修复 + 表格排序 + 动态采集灵敏度（**待发布**） |

---

## 4. 踩过的坑与修复（重点！）

### 坑 1：采集器 gRPC `Unimplemented: unknown service experimental.v2rayapi.StatsService`
- **现象**：部署后 `[Collector] QueryStats error: code = Unimplemented`
- **根因**：用户的 sing-box fork 在 `experimental/v2rayapi/stats.go` 第 22 行把服务名改写为 v2ray-api 兼容名 `v2ray.core.app.stats.command.StatsService`（为了兼容 v2rayN 等客户端）。我们用的官方 sing-box v1.14 客户端连的是 `experimental.v2rayapi.StatsService` → 不存在。
- **排查**：grpcurl + 手写精简 proto 直接探测两个服务名（`v2ray.core...` 返回真实数据，`experimental...` 报不存在）。
- **修复**：采集器改用 `github.com/v2fly/v2ray-core/v5/app/stats/command` 客户端（服务名精确匹配）。

### 坑 2：前端资源 404（第一重）
- **现象**：页面 HTML 能打开，JS/CSS 404。
- **根因**：`//go:embed web/*` 只嵌入第一层文件，`assets/` 子目录不进二进制。
- **修复**：`//go:embed all:web`（递归）。

### 坑 3：前端资源 404（第二重：vite base）
- **现象**：修复坑 2 后仍 404。
- **根因**：vite 的 `base: './'` 写在了 `build:` 里——**无效配置**（base 是顶层选项），构建产物 HTML 里是 `/assets/...` **绝对路径**，子路径反代（nginx `/control/`）下浏览器请求根路径 `/assets/` → 404。
- **排查**：`curl https://host/control/` 拿 HTML，看资源引用是 `/assets/` 还是 `./assets/`；再分别 curl 根路径与 /control/ 路径对比状态码。
- **修复**：`base: './'` 移到顶层；同时前端 API/WebSocket 全部用相对路径（`fetch('api/v1/...')`、`new URL('api/v1/ws/rt', location.href)`）。

### 坑 4：概览大盘/入站节点/历史统计全空
- **现象**：API 返回 count 正确但字段全 None。
- **根因**：gorm 模型没加 json tag → JSON 字段是大驼峰（`Category`/`TargetName`），前端读 snake_case（`category`/`target_name`）。
- **修复**：models 全部字段补 `json:"snake_case"` tag。

### 坑 5：在线状态不准（全员永远在线）
- **现象**：WebSocket 快照里零流量用户也显示在线。
- **根因**：realtime `Submit` 无条件刷新活跃时间戳；采集器对 gRPC 返回的所有 target（含增量为 0）都调用。
- **修复**：`Submit` 开头 `if up == 0 && down == 0 { return }`——只有实际流量才标记活跃，120s 后自动离线。
- **验证**：测试实例实测 2 用户在线（桐薇/beitou）与数据库流量记录完全吻合。

### 坑 6：CI 失败（go vet）
- **现象**：CI 的 vet 步骤失败。
- **根因**：重建时残留旧测试文件（引用已删除的 `models.SysStatLog`、旧 `CycleStart` 结构）；`go build` 不编译测试文件所以本地没发现。
- **修复**：删除过期测试，重写模型测试适配新结构（周期窗口/限额逻辑）。

### 坑 7：nfpm（deb 打包）系列
- nfpm **不支持环境变量替换**（`${VAR}`/模板均无效）→ nfpm.yaml 用 `__VERSION__/__ARCH__` 占位符 + CI 里 `sed` 替换
- 官方 `nfpm.goreleaser.com/install.sh` 返回 404 → CI 里从 GitHub Release 下载固定版本（v2.47.0）
- deb 脚本中文注释乱码 → postinstall/preremove 用纯 ASCII 注释

### 坑 8：实时监控其他细节
- **用户排序乱跳**：后端遍历 Go map 推送（顺序随机）→ publish 前 `sort.SliceStable` 按总速率降序 + 名称升序稳定排序。
- **速率图看不到波动**：①x 轴 `type:'time'` 但数据塞的是 `toLocaleTimeString` 字符串（ECharts time 轴要 epoch ms 时间戳）→ 改用 `Date.now()`；②y 轴大值压平小波动 → `scale:true` + 轴标签/悬浮框单位格式化（B/s→MB/s）+ dataZoom 缩放。
- **表格排序需求**：用户管理/历史统计/实时监控三张表都支持表头点击排序（可复用组件 `web/src/components/SortableTh.vue`，中文 `localeCompare('zh')`）。
- **动态采集灵敏度（实时监控）**：WS 订阅数 > 0 时采集器自动切高频（前端可选 1s/2s/5s/10s，通过 WS 控制消息 `{"action":"set_interval","interval_ms":N}` 调整）；断开/页面关闭自动恢复默认 10s。实现：collector ticker 运行时 `Reset` + realtime 订阅计数 + `pollMs` 原子变量。**注意：高频模式写库频率提升（数据量约 5 倍），仅实时页打开时生效**。

### 坑 9：PowerShell/脚本环境
- 本机 PowerShell 不支持 `&&`；curl 是 Invoke-WebRequest 别名（用 `curl.exe`）
- **写脚本时密码会被显示层脱敏成 `***` 破坏命令** → 排查脚本从 `/etc/sing-monitor/config.yaml` 动态读取密码（`grep -A6 '^postgres:' ... | awk`），避免硬编码

---

## 5. 服务器环境速查（rjp）

| 项 | 值 |
|---|---|
| SSH | `rjp`（rjp.wowoha.top:911，user，本机 `~/.ssh/config` 已配，密钥免密） |
| 面板 URL | `https://rjp.wowoha.top/control/`（nginx 子路径反代 → 127.0.0.1:8090） |
| 生产服务 | **`sing-monitor.service`**（deb 安装，`/lib/systemd/system/sing-monitor.service`）——**不是** v2rayapi.service（已被 deb 替换移除） |
| 生产二进制 | `/usr/bin/sing-monitor-server`（root 所有，替换需 sudo） |
| 生产配置 | `/etc/sing-monitor/config.yaml`（与原系统完全兼容；auth 密码/数据库密码都在里面） |
| 数据库 | PostgreSQL `singbox` 库：users 15 / inbound_nodes 10 / traffic_totals 10 / traffic_logs 14万+ |
| sing-box | `/usr/bin/sing-box`，v2ray_api 监听 127.0.0.1:8080（**legacy 服务名**）；进程 8/4 起未重启过 |
| 测试实例 | `~/singtest/sing-monitor-test -config config.test.yaml`（端口 8091，库 singbox_test，登录密码 testpass） |
| sudo | 无密码，部署命令需用户本人执行 |

### 常用运维命令

```bash
# 部署/升级（用户执行）
sudo cp ~/singtest/sing-monitor-fix /usr/bin/sing-monitor-server
sudo systemctl restart sing-monitor.service

# 回滚
sudo cp /usr/local/bin/sing-monitor.v020.bak /usr/local/bin/sing-monitor 2>/dev/null  # 旧路径备份
# 查看日志
journalctl -u sing-monitor.service -f
# 验证版本
/usr/bin/sing-monitor-server -version
curl -s http://127.0.0.1:8090/api/v1/version
# 探测 sing-box gRPC 服务
/tmp/gc/grpcurl -plaintext -proto /tmp/v2ray_stats_min.proto -import-path /tmp \
  -d '{"pattern":"user>>>","reset":true}' 127.0.0.1:8080 v2ray.core.app.stats.command.StatsService/QueryStats
```

### 配置生成安全策略提醒
- config.json 里有 **7 个节点不在数据库**（手动加的）：`quic`、`ipad_0b46tuic`、`bella_b86ftuic`、`chen_5c89tuic`、`yujia_aae6tuic`、`wantong_2ab9tuic`、`guangxizhuangxiu_d547tuic`
- 配置生成**不会删除它们**；如需面板管理，用面板右上角「同步母配置」导入数据库

---

## 6. 遗留事项

- [ ] 生产部署 v0.2.4（用户需执行 sudo 替换 + 重启；替换后验证 /control/ 三处显示 + 实时监控在线/排序/图表/灵敏度）
- [ ] 打 tag v0.2.4（含全部修复与功能）
- [ ] **排查教训：调试用测试实例必须及时杀掉**——测试实例与生产采集器共享 sing-box 8080 且都 reset=true，会互相抢计数器导致生产数据中断（本次已发生）
- [ ] config.json 的 7 个外部节点：决定是否「同步母配置」导入
- [ ] 工作区 `ruyizf_pay.ps1` 有明文商户密钥（已从 git 剔除，建议改环境变量读取）
- [ ] `singbox-monitor.bin`（旧系统二进制）备份在 Windows 工作区，勿删

## 7. 给后续 AI 的提示

1. **项目状态判断先看版本**：`-version` / `/api/v1/version` / 日志首行
2. **验证改动用 8091 测试实例**（不碰生产），二进制 `~/singtest/sing-monitor-fix` 上传后 cp 即可
3. **grpcurl 探测**是排查 gRPC 问题的利器（proto 在 `deploy/v2ray_stats_min.proto`）
4. **WebSocket 在线状态**用 `deploy/ws_check.py` 验证
5. **前端构建链路**：`web/` npm build → 产物复制到 `server/api/web/`（git 忽略）→ go build 嵌入 → CI 里同样流程
6. **改数据库模型必须加 json tag**，否则前端字段对不上（坑 4）
7. **服务器上所有密码从 `/etc/sing-monitor/config.yaml` 动态读取**，不要硬编码进脚本
