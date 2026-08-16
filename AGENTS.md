# Agent 指南

## 项目结构

- `backend/` — Go 后端（Gin），入口 `backend/main.go`
- `frontend/` — Vue 3 + Element Plus 前端（Vite 构建，产物嵌入 Go 二进制）
- `build.sh` — 完整构建脚本（npm build → 复制 dist → go build）

## 常用命令

```bash
# 后端构建（带 sing-box 特性标签）
cd backend && go build -tags "with_utls with_quic with_gvisor with_clash_api" -o ../build/sing-panel .

# 前端构建
cd frontend && npm run build
```

## sing-box 内核源码位置

内核以 Go 模块依赖引入（`backend/go.mod` 的 require），**不是**仓库内 submodule。
查看内核源码请去模块缓存：

```bash
go env GOMODCACHE
# 例如: $GOMODCACHE/github.com/sagernet/sing-box@v1.14.0-beta.15/
```

需要修改内核逻辑时，fork sing-box 到自己的仓库，再在 `backend/go.mod` 中
`replace github.com/sagernet/sing-box => <fork 路径>`。

## 关键说明

- 后端 API 路由在 `backend/main.go` 的 `api := router.Group("/api")` 中注册
- 业务逻辑在 `backend/services/`，HTTP 处理器在 `backend/handlers/`
- 数据库为 bbolt（`backend/services/db.go`），按 bucket 存储配置
- 前端页面在 `frontend/src/views/`，API 封装在 `frontend/src/api/`
- 静态资源由后端在构建期嵌入（`//go:embed frontend/dist/*`），修改前端后需
  重新执行 build.sh 全量构建
- `/assets/*` 由 `precompressedFileServer` 提供预压缩（.br/.gz）静态服务，
  修改后需同步更新 `backend/frontend/dist` 并重新构建
