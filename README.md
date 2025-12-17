# Go项目推荐：AI应用服务平台（GopherAI）

GopherAI 是一个基于 Go + Vue3 的 AI 应用服务平台，聚合了多会话聊天、图像识别、流式输出等典型 AI 能力，配套 MySQL / Redis / RabbitMQ 等基础设施，可直接作为全栈示例或生产级骨架使用。

## 🚀 核心特性

- **多会话 AI 助手**：Gin + GORM + Redis 维护用户上下文，RabbitMQ 异步写入历史消息，前端通过 SSE 实时接收回复。
- **图像识别链路**：提供图片上传、预处理、ONNXRuntime 推理到分类标签输出的全流程代码，便于扩展更多视觉模型。
- **RAG + 搜索增强**：后端通过 `usingGoogle` / `usingRAG` 开关动态接入谷歌搜索工具和火山引擎 VikingDB 检索器，结合大模型进行知识补全。
- **流式推理与消息落库**：支持会话首次创建与既有会话触发 SSE 流式回答，消息借助 RabbitMQ 异步入库，保证高并发下的实时体验与可靠性。
- **会话热启动管理**：服务启动时自动从 MySQL 预载会话/消息到内存态 AIHelper 管理器，支持历史会话秒级恢复和多模型策略。
- **高并发友好架构**：通用中间件（日志、认证、限流）、RabbitMQ 解耦写入、Redis 缓存加速，确保响应速度和扩展能力。
- **AI 工厂模式**：统一的 Provider 接入模式，便于新增不同类型的模型或第三方服务。
- **全栈实现**：Vue3 + Element Plus 构建的管理面板覆盖登录注册、验证码校验、聊天与图像识别等场景。

## 🧠 外部 AI 能力与第三方服务

| 能力 | 使用场景 | 配置入口 | 说明 |
| --- | --- | --- | --- |
| Qwen-Plus（灵积 DashScope OpenAI 兼容接口） | 主聊天模型（`modelType=1`），支持 Google / RAG 工具调用 | `config/env.sh` → `OPENAI_API_KEY` / `OPENAI_BASE_URL` / `OPENAI_MODEL_NAME` | 默认模型为 `qwen-plus`，通过 CloudWeGo EinO 对接 OpenAI Chat API 兼容层。 |
| Qwen3-VL-Plus | 图片理解 / 多模态问答 | `config/config.toml` → `[imageAIConfig]` | 依托 DashScope 兼容接口的多模态模型，`common/image` 已封装图片转 base64 的推理链路。 |
| 火山引擎 VikingDB 向量数据库 | RAG 知识检索工具 | `config/config.toml` → `[vikingDBConfig]` | `common/tools` 中使用 AK/SK 构建 `Retriever`，`usingRAG=true` 时在回答中附带“参考资料”引用。 |
| Google Custom Search JSON API | 外部实时搜索 | `config/config.toml` → `[googleConfig]` | 通过 CloudWeGo EinO ToolNode 暴露给 Qwen-Plus，开启 `usingGoogle=true` 后自动调用并将结果回注上下文。 |

## 🧱 技术栈

- **后端**：Go 1.20+、Gin、GORM、RabbitMQ、Redis、MySQL、ONNXRuntime
- **前端**：Vue3、Vue Router、Element Plus、Axios
- **基础设施**：MySQL 8、Redis 6、RabbitMQ 3、Nginx（可选）

## 🗺 系统架构

![image](https://file1.kamacoder.com/i/web/2025-11-20_09-34-09.jpg)

> 架构图覆盖 Web 层、业务服务、AI 推理、消息队列与数据层，展示了从请求进入、AI 处理到结果落库和前端展示的完整链路。

## 📡 端口总览

| 模块 | 端口 | 说明 | 来源 |
| --- | --- | --- | --- |
| Go 后端服务 | `9090` | Gin API / SSE 服务入口 | `config/config.toml` → `[mainConfig] port` |
| Vue 前端 DevServer | `8080` | `vue-cli-service serve` 默认访问地址 | `vue-frontend/vue.config.js` |
| MySQL | `3307` | 主业务数据库 `GopherAI` | `config/config.toml` → `[mysqlConfig] port` |
| Redis | `6380` | 验证码、缓存等 | `config/config.toml` → `[redisConfig] port` |
| RabbitMQ | `5672` | 异步消息/任务队列 | `config/config.toml` → `[rabbitmqConfig] port` |

## 📁 主要目录

- `common/`：数据库、Redis、RabbitMQ 等通用客户端初始化
- `controller/`、`router/`：HTTP API、SSE 推送接口
- `service/`、`dao/`：业务逻辑与数据访问
- `vue-frontend/`：Vue3 前端代码
- `config/`：TOML 配置和运行所需脚本

## ⚙️ 环境与配置

1. 根据 `config/config.toml` 设置数据库、Redis、RabbitMQ、邮件等连接信息，并补充 `[googleConfig]`、`[vikingDBConfig]`、`[imageAIConfig]` 中的密钥与集合信息。
2. 在 `config/env.sh` 中写入 DashScope（Qwen-Plus）兼容接口所需的 `OPENAI_API_KEY`、`OPENAI_BASE_URL`、`OPENAI_MODEL_NAME`，运行前执行 `source config/env.sh`。
3. 如果需要本地 ONNX 推理，确保安装 ONNXRuntime 依赖，并设置 `config/env.sh` 中的 `LD_LIBRARY_PATH`。
4. 保证上表列出的端口未被占用，或在配置文件中调整后同步更新 README。

## 🛠 能力开关示例

聊天接口统一支持以下 JSON 字段：

```json
{
  "question": "介绍最新的医学研究进展",
  "modelType": "1",
  "sessionId": "xxxx",        // 新会话可省略
  "usingGoogle": true,        // 调用 Google Search Tool
  "usingRAG": true            // 触发 VikingDB 检索并追加参考资料
}
```

- `POST /chat/send-new-session` / `/chat/send`：同步回答。
- `POST /chat/send-stream-new-session` / `/chat/send-stream`：通过 SSE 推送增量 token，并在结尾发送 `[DONE]`。
- 所有消息会先写入内存态 AIHelper，再异步投递到 `Message` 队列持久化到 MySQL。

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

> 特别感谢 Codex 的大力支持，协助我们快速完善文档与代码细节。
