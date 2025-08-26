# GoWasmSSH - 基于WebAssembly的浏览器SSH客户端

[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue)](https://golang.org/)
[![Node.js Version](https://img.shields.io/badge/Node.js-18%2B-green)](https://nodejs.org/)

## 📖 项目简介

GoWasmSSH 是一个基于 WebAssembly (WASM) 和 Go 语言开发的浏览器端 SSH
客户端。它允许用户直接在浏览器中建立 SSH
连接，无需安装任何本地客户端软件，提供了完整的终端模拟和 SFTP 文件传输功能。

## ✨ 核心特性

- **🖥️ 完整的终端模拟**: 基于 xterm.js 提供完整的终端体验
- **📁 SFTP 文件管理**: 内置 SFTP 客户端，支持文件上传下载和目录浏览
- **🔐 多种认证方式**: 支持密码认证和 SSH 密钥认证（含密码短语）
- **🌐 WebSocket 代理**: 通过 WebSocket 代理连接 SSH 服务器
- **📱 响应式界面**: 使用 Preact + TailwindCSS 构建的现代化界面
- **🔒 安全提示**: 支持 SSH 指纹验证和安全确认
- **⚡ 高性能**: 使用 Vite 构建，开发体验优秀

## 🛠️ 技术栈

### 后端 (Go WASM)

- **Go 1.23+** - 编译为 WebAssembly
- **golang.org/x/crypto/ssh** - SSH 协议实现
- **github.com/pkg/sftp** - SFTP 客户端
- **syscall/js** - JavaScript 互操作

### 前端

- **Preact** - 轻量级 React 替代
- **xterm.js** - 终端模拟器
- **TailwindCSS** - 实用优先的 CSS 框架
- **Vite** - 现代化构建工具
- **DaisyUI** - TailwindCSS 组件库

## 📦 安装与使用

### 前提条件

- Go 1.23+
- Node.js 18+
- pnpm (推荐) 或 npm

### 快速开始

1. **克隆项目**
   ```bash
   git clone https://github.com/wrtx-dev/GoWasmSSH.git
   cd gowasmssh
   ```

2. **构建 WASM 模块**
   ```bash
   # 使用 Makefile 自动构建（推荐）
   make client

   # 或者手动构建
   GOOS=js GOARCH=wasm go build -o webpage/public/ssh.wasm
   cp $(go env GOROOT)/misc/wasm/wasm_exec.js webpage/public/
   ```

3. **安装前端依赖**
   ```bash
   cd webpage
   pnpm install  # 或 npm install
   ```

4. **启动开发服务器**
   ```bash
   pnpm dev
   ```

5. **打开浏览器访问** `http://localhost:3000`

### 构建生产版本

```bash
# 构建前端
cd webpage
pnpm build

# 构建完整的静态文件到 dist 目录
```

## 🚀 使用说明

### 连接 SSH 服务器

1. 打开浏览器访问应用
2. 点击"连接"按钮
3. 填写 SSH 连接信息：
   - **主机地址**: SSH 服务器地址
   - **端口**: SSH 端口（默认 22）
   - **用户名**: SSH 用户名
   - **认证方式**: 选择密码或私钥
   - **代理地址**: WebSocket 代理地址（可选）
4. 点击"连接"建立 SSH 会话
5. 使用文件夹图标打开 SFTP 文件浏览器

### SFTP 文件管理

- **上传文件**: 拖拽文件到文件浏览器或点击上传按钮
- **下载文件**: 右键点击文件选择下载
- **创建目录**: 点击新建文件夹按钮
- **删除文件**: 右键点击文件选择删除
- **重命名**: 右键点击文件选择重命名

## 🔧 配置说明

### 代理服务器

项目包含一个内置的 WebSocket 代理服务器，可以将 WebSocket 连接转换为 TCP 连接：

```bash
# 构建并启动代理服务器
make all
./gowasmssh

# 或者指定监听地址和端口
./gowasmssh -listen 0.0.0.0 -port 9090
```

### 代理设置说明

项目支持通过 WebSocket 代理连接 SSH
服务器。可以在连接对话框中设置代理地址，格式为：

```
ws://your-proxy-server:port
```

**注意**: 受浏览器安全策略的影响，如果部署在 HTTPS 环境下，则只支持 WSS 协议。

## 🌐 部署指南

### 静态文件部署

1. 构建生产版本：
   ```bash
   make client
   cd webpage
   pnpm build
   ```

2. 将 `webpage/dist` 目录部署到任何静态文件服务器（Nginx、Apache、CDN 等）

3. 确保代理服务器可访问，或修改配置使用其他代理服务

### Cloudflare Workers 部署

项目提供了 Cloudflare Workers 部署选项，可以作为 WebSocket 代理服务器使用：

1. **安装依赖**：
   ```bash
   cd cf-workers
   pnpm install  # 或 npm install
   ```

2. **本地开发**：
   ```bash
   pnpm dev
   ```

3. **部署到 Cloudflare**：
   ```bash
   pnpm deploy
   ```

4. **配置前端使用 Workers 代理**： 在连接设置中使用 Workers 的 URL
   作为代理地址：
   ```
   wss://your-worker-name.your-subdomain.workers.dev/ws
   ```

**优势**：

- 全球分布式部署，低延迟
- 无需维护服务器基础设施
- 自动 HTTPS 支持
- 内置 DDoS 防护

### Docker 部署示例

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN GOOS=js GOARCH=wasm go build -o webpage/public/ssh.wasm

FROM node:18-alpine AS frontend
WORKDIR /app
COPY --from=builder /app/webpage .
RUN npm install -g pnpm
RUN pnpm install
RUN pnpm build

FROM nginx:alpine
COPY --from=frontend /app/dist /usr/share/nginx/html
```

## 🛠️ 开发指南

### 项目结构

```
gowasmssh/
├── main.go                 # WASM 入口文件
├── server.go              # 代理服务器
├── go.mod                 # Go 模块定义
├── Makefile              # 构建脚本
├── js/                   # JavaScript 互操作代码
│   ├── js.go
│   ├── promise.go
│   └── ws.go
├── ssh/                  # SSH 相关实现
│   ├── ssh.go
│   └── sftp.go
├── package/              # 内部包
│   └── server/           # 服务器实现
└── webpage/              # 前端代码
    ├── package.json
    ├── vite.config.js
    ├── public/           # 静态资源
    │   ├── ssh.wasm
    │   ├── wasm_exec.js
    └── src/              # 前端源码
        ├── index.jsx     # 主入口
        ├── style.css     # 样式文件
        └── hooks/        # React Hooks
            └── useTerm.js
└── cf-workers/           # Cloudflare Workers 代理服务器
    ├── src/
    │   └── index.ts      # Workers 入口文件
    ├── public/           # 静态资源（用于 Assets 功能）
    ├── wrangler.jsonc    # Workers 配置
    └── package.json      # Node.js 依赖
```

### 开发命令

```bash
# 开发模式（自动重新构建）
make client && cd webpage && pnpm dev

# 构建所有
make all

# 清理构建文件
make clean

# 启动代理服务器
make serve

# Cloudflare Workers 开发
cd cf-workers && pnpm dev

# Cloudflare Workers 部署
cd cf-workers && pnpm deploy
```

## 📝 代码参考说明

本项目在开发过程中参考了以下开源项目的实现思路和代码：

1. **https://github.com/hullarb/dom**
   - WebAssembly 与 JavaScript 互操作的最佳实践
   - Go 在浏览器环境中的异常处理机制
   - Promise 和异步操作的封装模式

2. **https://github.com/hullarb/ssheasy**
   - SSH 协议在 Web 环境中的实现方案
   - WebSocket 传输层的设计模式
   - 终端会话管理和数据流处理

这些参考项目的优秀实践为本项目的开发提供了重要的技术指导和灵感。

## ⚖️ 开源协议

本项目采用 **Apache 2.0 开源协议**。

## 👍 项目优势

1. **跨平台兼容**: 纯浏览器解决方案，无需安装客户端
2. **安全性**: 支持 SSH 指纹验证和密钥认证
3. **功能完整**: 同时提供终端和 SFTP 功能
4. **现代化界面**: 基于 Preact 和 TailwindCSS 的响应式 UI
5. **性能优化**: 使用 Vite 构建，开发体验优秀
6. **易于部署**: 静态文件部署，支持 CDN 分发

## ⚠️ 局限性

1. **网络限制**: 需要 WebSocket 代理或 CORS 支持
2. **性能开销**: WASM 模块加载需要一定时间
3. **功能限制**: 某些高级 SSH 功能可能受限
4. **浏览器兼容**: 需要现代浏览器支持 WASM
5. **文件传输**: 大文件传输可能受内存限制

## 🐛 已知问题

- 某些特殊终端字符可能渲染不正确
- 连接断开后的重连机制需要完善
- SFTP 大文件上传下载进度显示
- 移动端适配需要优化

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

### 开发规范

- 遵循 Go 代码规范
- 使用 Prettier 格式化前端代码
- 提交信息使用英文描述
- 确保所有测试通过

## 📞 联系方式

- 项目主页: https://github.com/wrtx-dev/GoWasmSSH
- Issues: https://github.com/wrtx-dev/GoWasmSSH/issues
- 邮箱: your-email@example.com

## 🙏 致谢

感谢所有为本项目提供灵感和代码参考的开源项目开发者，特别是：

- [hullarb/dom](https://github.com/hullarb/dom)
- [hullarb/ssheasy](https://github.com/hullarb/ssheasy)
- [xterm.js](https://xtermjs.org/) 团队
- [Go WebAssembly](https://github.com/golang/go/wiki/WebAssembly) 社区

---

⭐ 如果这个项目对您有帮助，请给它一个 Star！

## 🔄 更新日志

### v0.1.0 (2024-08-25)

- 初始版本发布
- 支持基本的 SSH 终端功能
- 支持 SFTP 文件管理
- 提供 WebSocket 代理服务器
- 现代化前端界面

## 🚧 开发计划

- [ ] 会话管理（多标签页支持）
- [ ] 连接历史记录
- [ ] 主题切换
- [ ] 移动端优化
- [ ] 性能优化
- [ ] 更多认证方式支持
