# Sing Box 管理面板

一个用于管理 Sing Box 内核的 Web 管理面板。

## 功能

- 下载 Sing Box 内核（支持 latest、stable 和自定义链接）
- 查看可用版本列表
- 安装指定版本
- 删除已安装的内核
- 实时下载进度显示

## 技术栈

- **后端**: Go + Gin
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
./sing-panel
```

访问 http://localhost:8080

## API 接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | /api/kernel/status | 获取当前内核状态 |
| GET | /api/kernel/versions | 获取可用版本列表 |
| POST | /api/kernel/download | 下载内核 |
| POST | /api/kernel/stop | 停止下载 |
| DELETE | /api/kernel | 删除内核 |
| POST | /api/kernel/switch | 切换版本 |

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
