# Sing Box 管理面板

一个用于管理 Sing Box 内核的 Web 管理面板。

## 功能

- 嵌入 sing-box 内核（libbox），开箱即用，无需单独下载内核
- 完整的 sing-box 配置管理（Inbound / Outbound / Ruleset / 路由规则 / DNS / 服务 / HTTP 客户端 / Experimental）
- **TProxy 透明代理**：启用 tproxy inbound 后，自动通过 nftables/netlink 系统调用配置防火墙规则（无需安装 iptables/nft/ip 命令行工具），客户端网关指向本机即可透明代理；停止内核时自动清理规则
- 支持从订阅链接导入配置、GitHub GEO 规则树自动刷新
- 实时状态监控（内存 / 协程 / GC / 运行时长）
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
./build.sh
./build/sing-panel
```

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
| GET/POST/PUT/DELETE | /api/instances | 多实例管理（增删改查） |
| GET | /api/instances/status | 检查全部实例状态与配置一致性 |
| POST | /api/instances/:id/sync | 同步配置（push / pull） |
| POST | /api/instances/sync-all | 推送本机配置到全部实例 |
| PUT | /api/instances/sync-token | 设置同步令牌 |
| GET | /api/instances/local-info | 本机面板信息与配置指纹 |
| GET | /api/db/export | 导出数据库（受同步令牌保护） |
| POST | /api/db/import | 导入数据库（受同步令牌保护） |

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

> sing-box 内核以 Go 模块依赖引入（`backend/go.mod` 中 require），源码可在模块缓存中查看：`$(go env GOMODCACHE)/github.com/sagernet/sing-box@v1.14.0-beta.15/`。如需修改内核，请 fork 后在 go.mod 中 replace 到自己的 fork。