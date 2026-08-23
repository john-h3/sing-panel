# Sing Box 管理面板

一个用于管理 Sing Box 内核的 Web 管理面板。

## 功能

- 嵌入 sing-box 内核（libbox），开箱即用，无需单独下载内核
- 完整的 sing-box 配置管理（Inbound / Outbound / Ruleset / 路由规则 / DNS / 服务 / HTTP 客户端 / Experimental）
- **定制化 Outbound**：支持 fork 提供的 `fallback` 和 `loadbalance`，可拖拽调整出站顺序；LoadBalance 支持轮询、随机、加权、最少连接和一致性哈希策略
- **定制化功能开关**：开关位于面板 `app_config` 中，不会下发给 sing-box；关闭时定制化 Outbound 仍可创建、编辑和删除，但不能启用，已有相关状态会自动关闭
- **TProxy 透明代理**：启用 tproxy inbound 后，自动通过 nftables/netlink 系统调用配置防火墙规则（无需安装 iptables/nft/ip 命令行工具），客户端网关指向本机即可透明代理；停止内核时自动清理规则
- 支持从订阅链接导入配置、GitHub GEO 规则树自动刷新
- 实时状态监控（内存 / 协程 / GC / 运行时长 / 构建时间）
- 健康检查接口：根据嵌入式 sing-box 内核是否运行返回 HTTP 200 / 400
- 日志管理：统一查看面板、Gin 和 sing-box 日志，支持实时 SSE 推送、暂停/恢复和级别筛选
- 数据库导出 / 导入
- **多实例管理**：在单个面板上管理多个面板实例，基于导出/导入实现配置同步，实时对比各实例配置是否一致

## 技术栈

- **后端**: Go + Gin + bbolt
- **前端**: Vue 3 + Vite + Element Plus

## 快速开始

### 开发模式

**启动后端:**
```bash
cd backend
go mod tidy
go run .
```

**启动前端:**
```bash
cd frontend
npm install
npm run dev
```

访问 http://localhost:3000

### 生产构建

```bash
chmod +x build.sh
./build.sh                 # 交互式选择目标架构
./build.sh linux/arm64     # 直接指定目标架构，跳过交互
./build/sing-panel         # 当前平台产物；交叉编译产物为 build/sing-panel-<os>-<arch>
```

支持的目标架构可用 `./build.sh --help` 查看（linux amd64/arm64/armv7、darwin、windows）。

`build.sh` 会根据前端源码、依赖锁文件和 Vite 配置的内容缓存前端构建结果；输入未变化时会复用 `frontend/dist`，跳过 `npm install` 和前端打包。前端重新构建时使用 npm 本地缓存，并关闭 audit 和 fund 请求以缩短构建时间。需要强制重新构建前端时，删除 `build/frontend.sha256` 和 `frontend/dist/` 即可。

默认只生成 gzip 压缩资源。如需额外生成 Brotli 资源，使用 `./build.sh linux/arm64 --brotli`；Brotli 模式会单独缓存，不会与默认模式混用。

### 启动脚本

根目录提供了 `start.sh`（开发环境用），自动选择 `build/` 下当前平台的编译产物并启动：

```bash
./start.sh                    # 默认监听 :8080，数据目录 ./data
./start.sh --listen :3000     # 指定端口
./start.sh --data-dir /etc/sing-panel   # 指定数据目录
./start.sh install            # 安装为系统服务（参数透传给二进制）
```

- 数据目录（存放 `sing-panel.db`）默认为项目根目录下 `data/`，可通过 `--data-dir` 指定
- 安装为系统服务时，`--data-dir` 与 `--listen` 会写入服务文件

### 手动启动

```bash
cd backend && go run . --data-dir ./data   # 开发模式
```

### 部署

`deploy_test.sh` 和 `deploy_prod.sh` 会先调用 `build.sh linux/arm64` 构建 ARM64 版本，再将新二进制上传为临时文件，停止旧服务后原子替换实际的 OpenRC 执行文件；停止时会校验进程并在必要时使用 TERM/KILL，启动后会检查服务进程是否仍在运行。通过 `start.sh install` 安装的 OpenRC 服务会主动关闭日志 SSE 长连接，并配置 `TERM/10/KILL/5` 停止重试，避免服务因优雅退出超时而被误判为停止失败。部署脚本默认使用脚本内配置的远端地址，执行前请确认 SSH 免密登录和目标环境配置正确。

如果面板页面在后端重启或重新部署期间保持打开，首次访问尚未加载过的页面可能遇到旧版懒加载资源失效；前端会自动刷新一次并重新加载最新资源，通常无需手动刷新浏览器。

> 部署脚本匹配 `deploy_*.sh`，默认由 Git 忽略，不会被纳入版本跟踪。

### 健康检查

面板默认监听 `:8080`，提供无需认证的健康检查接口：

```bash
curl -i http://127.0.0.1:8080/health
```

返回规则：

- `200 OK`：嵌入式 sing-box 内核正在运行
- `400 Bad Request`：嵌入式 sing-box 内核未运行

该接口只检查内核运行状态，不检查第三方依赖；适合负载均衡器、OpenRC 或外部监控探测。

## 多实例管理

在「系统设置 → 多实例管理」页面：

1. **添加实例**：填写远端面板的地址（可选同步令牌），即可将其纳入管理
2. **一致性检测**：点击「全部检查」，面板会拉取远端导出的配置并计算指纹，与本机对比，标记每个实例为「一致 / 不一致 / 不可达」
3. **配置同步**：
   - **推送**：将本机配置同步到指定实例
   - **拉取**：用远端实例配置覆盖本机配置
   - **推送全部**：将本机配置同步到所有实例
4. **同步令牌**（可选）：面板可设置一个令牌，设置后其他面板访问导出/导入/面板信息接口必须携带该令牌（`X-Sync-Token` 头），令牌存于本机状态中，不会被配置同步覆盖

一致性判断基于导出的配置数据：`config` 与 `singbox` 两个 bucket（排除 GEO 规则树缓存）；各面板本机的运行状态、managed instances、同步令牌不参与比对。

## 定制化 Outbound

面板支持当前 fork 中提供的两个定制化 Outbound：

- **Fallback**：按配置顺序尝试出站，直到连接成功；支持拖拽调整尝试顺序。
- **LoadBalance**：按策略选择出站；支持 `round_robin`、`random`、`weighted_round_robin`、`weighted_random`、`least_connections` 和 `consistent_hash`，并可配置每个出站的权重。

在「Singbox → Experimental」中开启「定制化功能」后，已启用的定制化 Outbound 才会导出给内核。该开关存储在面板自身的 `config/app_config` 中，而不是 sing-box 的 `experimental` 配置中。关闭开关时，面板会自动禁用所有 `fallback` 和 `loadbalance`，并清理它们在路由中的引用。

## TProxy 透明代理

1. 在「Singbox → Inbound」中新增类型为 **TProxy** 的入站，配置监听地址（默认 `0.0.0.0`）、监听端口（如 `5678`）和网络（`tcp` / `udp` / 自动 TCP+UDP）
2. 保存后启动 sing-box 内核，面板会自动通过 **nftables + netlink 系统调用**写入 TPROXY 规则与路由规则：
   - 拦截所有转发流量到 tproxy 端口，并排除本地目的、广播/组播、已标记流量，避免环路与 NAT 会话泄漏
   - 支持多 tproxy 入站、IPv4/IPv6
3. 局域网客户端将**默认网关**指向本机 IP 即可透明代理
4. 停止内核或面板退出时，自动清理全部规则

> **依赖**：需以 root 运行或具备 `CAP_NET_ADMIN`；不依赖 `iptables`/`nft`/`ip` 命令行工具（内核需支持 nf_tables，Linux 4.17+ 默认具备）。若需要调优 UDP NAT 会话开销，可在 tproxy 表单中调整「UDP 超时」。

## API 接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /health | 内核健康检查，运行返回 200，未运行返回 400 |
| GET | /api/panel/info | 本面板基本信息（供实例管理使用） |
| GET | /api/kernel/status | 获取当前内核状态 |
| GET | /api/kernel/system | 获取系统信息 |
| GET | /api/kernel/monitor | 获取运行时监控数据 |
| GET/PUT | /api/config | 读取/更新面板配置 |
| GET/PUT | /api/singbox | 读取/保存 sing-box 配置 |
| GET/POST/PUT/DELETE | /api/singbox/inbounds | Inbound 配置管理 |
| GET/POST/PUT/DELETE | /api/singbox/outbounds | Outbound 配置管理 |
| GET/POST/PUT/DELETE | /api/singbox/rulesets | Ruleset 配置管理 |
| GET/POST/PUT/DELETE | /api/singbox/route-rules | 路由规则管理 |
| GET/PUT | /api/singbox/route-config | 路由配置管理 |
| GET/PUT | /api/singbox/dns | DNS 配置管理 |
| GET/POST/PUT/DELETE | /api/singbox/services | 服务管理 |
| GET/POST/PUT/DELETE | /api/singbox/http-clients | HTTP 客户端管理 |
| GET/PUT | /api/singbox/experimental | Experimental 配置 |
| POST | /api/singbox/import | 从订阅链接导入配置 |
| GET/POST | /api/process/... | sing-box 进程控制（start/stop/restart/status） |
| GET | /api/stats/service | 服务统计信息 |
| GET | /api/system/init | 检测初始化系统（systemd/openrc） |
| POST | /api/system/restart-service | 重启 sing-panel 系统服务 |
| POST | /api/system/reboot-machine | 重启宿主机操作系统 |
| GET/POST/PUT/DELETE | /api/instances | 多实例管理（增删改查） |
| GET | /api/instances/status | 检查全部实例状态与配置一致性 |
| POST | /api/instances/:id/sync | 同步配置（push / pull） |
| POST | /api/instances/sync-all | 推送本机配置到全部实例 |
| PUT | /api/instances/sync-token | 设置同步令牌 |
| GET | /api/instances/local-info | 本机面板信息与配置指纹 |
| GET | /api/db/export | 导出数据库（受同步令牌保护） |
| POST | /api/db/import | 导入数据库（受同步令牌保护） |
| GET | /api/logs | 查询内存日志，支持 `limit`、`after`、`level`、`source` 参数 |
| GET | /api/logs/stream | 通过 SSE 接收实时内存日志 |
| DELETE | /api/logs | 清空内存日志 |

### 日志管理

「设置 → 日志管理」页面支持：

- 调整面板和嵌入式 sing-box 的统一日志级别；
- 查看面板、Gin、sing-box 的内存日志；
- 暂停实时日志接收，恢复后按日志序号补齐尚未被环形缓冲区覆盖的日志；
- 按来源和最低日志级别筛选；
- 开启或关闭“过滤检测 API”。该开关位于内存日志筛选区，与自动滚动并列，切换立即生效，状态保存在浏览器 localStorage 中（换浏览器或清除站点数据后会恢复默认开启）；
- 开启过滤时，前端仅隐藏 `/health` 的 Gin 访问记录，后端仍会保留日志。页面先过滤再显示最近 100 条，原始内存日志最多保留 2048 条。

日志文件默认写入 `/var/log/sing-panel.log`，按大小滚动并保留压缩归档；无权限写入默认目录时，应用会回退到数据目录下的 `sing-panel.log`。

## 项目结构

```
sing_panel/
├── backend/           # Go 后端
│   ├── main.go       # 入口文件
│   ├── handlers/     # HTTP 处理器
│   ├── services/     # 业务逻辑
│   └── models/       # 数据模型
├── frontend/         # Vue 前端
│   ├── src/
│   │   ├── views/    # 页面组件
│   │   ├── api/      # API 调用
│   │   └── router/   # 路由配置
│   └── package.json
└── build.sh          # 构建脚本
```

> sing-box 内核以 Go 模块依赖引入。面板保留官方模块路径的 import，并通过 `backend/go.mod` 中的远程 `replace` 使用 fork，例如：
>
> ```go
> require github.com/sagernet/sing-box v1.14.0-beta.15
>
> replace github.com/sagernet/sing-box => github.com/john-h3/sing-box v0.0.0-20260823071457-150e69f5c9f8
> ```
>
> fork 更新后，在 `backend/` 目录执行 `go mod edit -replace=github.com/sagernet/sing-box=github.com/john-h3/sing-box@<提交哈希>`，再运行 `GOPROXY=direct go mod tidy` 和 `go test ./...`。不要执行 `go get github.com/sagernet/sing-box@testing`，它会拉取官方仓库而不是 fork。
