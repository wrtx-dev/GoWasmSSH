# GoWasmSSH - 基于WebAssembly的浏览器SSH客户端

[![Apache 2.0 License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

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

### 构建步骤

1. **克隆项目**
   ```bash
   git clone https://github.com/wrtx-dev/GoWasmSSH.git
   cd gowasmssh
   ```

2. **构建 WASM 模块**
   ```bash
   # 编译 Go 代码为 WASM
   GOOS=js GOARCH=wasm go build -o ssh-page/public/ssh.wasm
   ```

3. **安装前端依赖**
   ```bash
   cd ssh-page
   pnpm install
   ```

4. **启动开发服务器**
   ```bash
   pnpm dev
   ```

5. **构建生产版本**
   ```bash
   pnpm build
   ```

## 🚀 快速开始

1. 打开浏览器访问 `http://localhost:3000`
2. 点击"连接"按钮
3. 填写 SSH 连接信息：
   - 主机地址
   - 端口（默认 22）
   - 用户名
   - 密码或私钥文件
   - 代理地址（可选）
4. 点击"连接"建立 SSH 会话
5. 使用文件夹图标打开 SFTP 文件浏览器

## 🔧 配置说明

### 代理设置

项目支持通过 WebSocket 代理连接 SSH
服务器。可以在连接对话框中设置代理地址，格式为：

```
ws://your-proxy-server:port
```

但是需要注意的是，受浏览器安全策略的影响，如果部署在https环境下，则只支持wss协议。

### 环境变量

前端可以通过修改`public/config.json`文件来设置默认代理地址。

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
