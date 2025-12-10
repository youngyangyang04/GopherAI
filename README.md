## 📡 端口总览

| 模块 | 端口 | 说明 | 来源 |
| --- | --- | --- | --- |
| Go 后端服务 | `9090` | Gin API / SSE 服务入口 | `config/config.toml` → `[mainConfig] port` |
| Vue 前端 DevServer | `8080` | `vue-cli-service serve` 默认访问地址 | `vue-frontend/vue.config.js` |
| MySQL | `3307` | 主业务数据库 `GopherAI` | `config/config.toml` → `[mysqlConfig] port` |
| Redis | `6380` | 验证码、缓存等 | `config/config.toml` → `[redisConfig] port` |
| RabbitMQ | `5672` | 异步消息/任务队列 | `config/config.toml` → `[rabbitmqConfig] port` |

# Go项目推荐：AI应用服务平台（GopherAI）

GopherAI 是一个基于 Go + Vue3 的 AI 应用服务平台，聚合了多会话聊天、图像识别、流式输出等典型 AI 能力，配套 MySQL / Redis / RabbitMQ 等基础设施，可直接作为全栈示例或生产级骨架使用。

## 🚀 核心特性

- **多会话 AI 助手**：Gin + GORM + Redis 维护用户上下文，RabbitMQ 异步写入历史消息，前端通过 SSE 实时接收回复。
- **图像识别链路**：提供图片上传、预处理、ONNXRuntime 推理到分类标签输出的全流程代码，便于扩展更多视觉模型。
- **高并发友好架构**：通用中间件（日志、认证、限流）、RabbitMQ 解耦写入、Redis 缓存加速，确保响应速度和扩展能力。
- **AI 工厂模式**：统一的 Provider 接入模式，便于新增不同类型的模型或第三方服务。
- **全栈实现**：Vue3 + Element Plus 构建的管理面板覆盖登录注册、验证码校验、聊天与图像识别等场景。

## 🧱 技术栈

- **后端**：Go 1.20+、Gin、GORM、RabbitMQ、Redis、MySQL、ONNXRuntime
- **前端**：Vue3、Vue Router、Element Plus、Axios
- **基础设施**：MySQL 8、Redis 6、RabbitMQ 3、Nginx（可选）

## 🗺 系统架构

![image](https://file1.kamacoder.com/i/web/2025-11-20_09-34-09.jpg)

> 架构图覆盖 Web 层、业务服务、AI 推理、消息队列与数据层，展示了从请求进入、AI 处理到结果落库和前端展示的完整链路。

## 📁 主要目录

- `common/`：数据库、Redis、RabbitMQ 等通用客户端初始化
- `controller/`、`router/`：HTTP API、SSE 推送接口
- `service/`、`dao/`：业务逻辑与数据访问
- `vue-frontend/`：Vue3 前端代码
- `config/`：TOML 配置和运行所需脚本

## ⚙️ 环境与配置

1. 根据 `config/config.toml` 设置数据库、Redis、RabbitMQ、邮件等连接信息。
2. 如果需要本地推理，确保安装 ONNXRuntime 依赖，并设置 `config/env.sh` 中的 `LD_LIBRARY_PATH`。
3. 保证上表列出的端口未被占用，或在配置文件中调整后同步更新 README。

## 🚀 快速开始

```bash
# 1. 初始化依赖
go mod download

# 2. 启动后端
go run main.go
```

前端：

```bash
cd vue-frontend
npm install
npm run serve
```

确保 MySQL、Redis、RabbitMQ 已启动并与配置文件保持一致。
