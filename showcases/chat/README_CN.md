# LangGraphGo 聊天应用程序

一个基于Web的复杂多会话聊天应用，集成了AI代理、工具支持和持久化本地存储。

## ✨ 特性

- 🔄 **多会话支持**：创建和管理多个独立的聊天会话
- 💾 **持久化存储**：所有对话自动保存到本地磁盘
- 🌐 **现代化Web界面**：简洁、响应式的Web UI，支持实时更新
- 🤖 **AI聊天代理**：先进的代理，具备会话历史管理功能
- 🔧 **工具集成**：支持Skills和MCP（模型上下文协议）工具
- 🔌 **多提供商支持**：兼容OpenAI、百度、Azure和任何OpenAI兼容API
- 🎨 **精美UI**：支持深色/浅色主题，流畅的动画效果
- 📝 **会话管理**：创建、查看、清除和删除会话
- ⚡ **热重载**：开发模式，自动代码重载
- 🐳 **Docker支持**：容器化部署就绪

## 🏗️ 架构

```
showcases/chat/
├── main.go                 # 应用程序入口点和服务器引导
├── pkg/                    # Go包
│   ├── chat/              # 聊天服务器和代理逻辑
│   │   └── chat.go        # 核心聊天功能
│   └── session/           # 会话管理
│       └── session.go     # 会话持久化
├── static/
│   ├── index.html        # Web前端
│   ├── style.css         # UI样式
│   └── script.js         # 前端逻辑
├── sessions/             # 本地会话存储（自动创建）
├── build/                # 构建输出目录
├── Makefile              # 构建自动化
├── Dockerfile            # Docker配置
├── .air.toml            # 热重载配置
├── go.mod
├── go.sum
├── .env                 # 配置（从.env.example创建）
└── README_CN.md
```

## 🚀 快速开始

### 选项1：使用Makefile（推荐）

```bash
# 克隆并导航到项目
cd showcases/chat

# 安装开发工具
make setup-dev

# 复制环境变量模板
cp .env.example .env

# 编辑.env并添加你的OpenAI API密钥
# OPENAI_API_KEY=sk-...

# 运行热重载（开发模式）
make dev

# 或正常运行
make run-dev
```

### 选项2：标准Go命令

```bash
cd showcases/chat

# 安装依赖
go mod download

# 复制环境变量模板
cp .env.example .env

# 编辑.env并添加你的OpenAI API密钥
# OPENAI_API_KEY=sk-...

# 构建并运行
go run main.go
```

服务器将在 `http://localhost:8080` 启动

## 🛠️ 开发工作流

### 使用Makefile

```bash
# 安装开发工具（air、golangci-lint等）
make setup-dev

# 运行热重载
make dev

# 运行所有检查（格式化、lint、vet、测试）
make check

# 构建生产版本
make build

# 构建所有平台
make build-all
```

### 常用Makefile目标

| 目标 | 描述 |
|------|------|
| `make dev` | 运行热重载 |
| `make run-dev` | 运行开发环境 |
| `make build` | 构建应用程序 |
| `make test` | 运行测试 |
| `make coverage` | 运行测试并生成覆盖率报告 |
| `make format` | 格式化代码 |
| `make vet` | 代码检查 |
| `make lint` | 代码规范检查 |
| `make docker-up` | 构建并运行Docker |
| `make clean` | 清理构建产物 |
| `make help` | 显示所有目标 |

## ⚙️ 配置

环境变量（在`.env`中）：

```env
# 必需：你的API密钥
OPENAI_API_KEY=your-api-key-here

# 可选：模型名称（默认：gpt-4o-mini）
OPENAI_MODEL=gpt-4o-mini

# 可选：OpenAI兼容API的Base URL
# 示例：
#   百度：https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions
#   Azure：https://your-resource.openai.azure.com/
#   Ollama：http://localhost:11434/v1
OPENAI_BASE_URL=

# 可选：服务器端口（默认：8080）
PORT=8080

# 可选：会话存储目录（默认：./sessions）
SESSION_DIR=./sessions

# 可选：每个会话最大消息数（默认：50）
MAX_HISTORY_SIZE=50

# 可选：技能目录（用于工具集成）
SKILLS_DIR=../../testdata/skills

# 可选：MCP配置路径
MCP_CONFIG_PATH=../../testdata/mcp/mcp.json

# 可选：聊天标题
CHAT_TITLE=LangGraphGo 聊天
```

### LLM提供商示例

**OpenAI**：
```env
OPENAI_API_KEY=sk-your-openai-key
OPENAI_MODEL=gpt-4o
```

**百度千帆**：
```env
OPENAI_API_KEY=your-baidu-token
OPENAI_BASE_URL=https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions
OPENAI_MODEL=ERNIE-Bot
```

**Azure OpenAI**：
```env
OPENAI_API_KEY=your-azure-key
OPENAI_BASE_URL=https://your-resource.openai.azure.com/
OPENAI_MODEL=your-deployment-name
```

**本地模型（Ollama、LM Studio）**：
```env
OPENAI_API_KEY=not-needed
OPENAI_BASE_URL=http://localhost:11434/v1
OPENAI_MODEL=llama2
```

## 📡 API端点

### 会话管理
- `POST /api/sessions/new` - 创建新会话
- `GET /api/sessions` - 列出所有会话
- `DELETE /api/sessions/:id` - 删除会话
- `GET /api/sessions/:id/history` - 获取会话消息
- `GET /api/client-id` - 获取当前客户端ID

### 聊天
- `POST /api/chat` - 发送消息
  ```json
  {
    "session_id": "uuid",
    "message": "你的消息",
    "user_settings": {
      "enable_skills": true,
      "enable_mcp": true
    }
  }
  ```
  响应：
  ```json
  {
    "response": "AI响应文本"
  }
  ```

### 工具
- `GET /api/mcp/tools?session_id=:id` - 列出可用的MCP工具
- `GET /api/tools/hierarchical?session_id=:id` - 获取分层结构的工具
- `GET /api/config` - 获取聊天配置

## 🧩 组件

### ChatAgent

`SimpleChatAgent`提供：
- 自动对话上下文管理
- 工具集成（Skills和MCP）
- 支持OpenAI兼容API
- 线程安全的会话历史
- 异步工具加载

### 会话管理

每个会话包括：
- 唯一的UUID标识符
- 完整的消息历史
- 持久化JSON存储
- 基于客户端的隔离
- 自动保存和加载

### 工具集成

应用支持两种类型的工具：

1. **Skills**：从`SKILLS_DIR`加载的预定义工具包
2. **MCP工具**：来自模型上下文协议服务器的动态工具

工具可以通过用户设置在每个会话中启用/禁用。

## 🐳 Docker部署

```bash
# 使用Docker Compose构建并运行
make docker-up

# 或手动：
docker build -t chat-app .
docker run -p 8080:8080 -e OPENAI_API_KEY=your-key chat-app
```

### Docker Compose

```yaml
version: '3.8'
services:
  chat:
    build: .
    ports:
      - "8080:8080"
    environment:
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - OPENAI_MODEL=gpt-4o-mini
    volumes:
      - ./sessions:/app/sessions
```

## 🧪 测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make coverage

# 运行特定测试
go test ./pkg/session -v
```

## 📦 构建

### 为当前平台构建
```bash
make build
```

### 跨平台构建
```bash
# 为所有平台构建
make build-all

# 为特定平台构建
make build-linux
make build-darwin
make build-windows
```

### 发布包
```bash
# 创建发布包
make release
```

输出将在`build/release/`中。

## 🔧 自定义

### 更改系统提示词

编辑`pkg/chat/chat.go`中的`NewSimpleChatAgent`函数：
```go
systemMsg := llms.MessageContent{
    Role:  llms.ChatMessageTypeSystem,
    Parts: []llms.ContentPart{llms.TextPart("你的自定义系统消息")},
}
```

### 添加自定义工具

1. 在你的技能目录中创建技能包
2. 按照示例中的技能包结构
3. 工具将自动加载

### 修改UI

编辑`static/`中的文件：
- `index.html` - 主要HTML结构
- `style.css` - 样式和主题
- `script.js` - 前端逻辑

## 🔍 开发

### 项目结构

- **main.go**：应用程序入口点、引导和优雅关闭
- **pkg/chat/**：核心聊天功能和HTTP处理器
- **pkg/session/**：会话持久化和管理
- **static/**：Web前端资源
- **Makefile**：构建自动化和开发工作流

### 添加功能

1. **新API端点**：添加到`pkg/chat/chat.go`
2. **新会话字段**：更新`pkg/session/session.go`
3. **前端更改**：修改`static/`文件
4. **配置**：添加到环境变量

### 代码质量

项目使用：
- `go fmt`用于格式化
- `go vet`用于静态分析
- `golangci-lint`用于全面的代码规范检查
- 对关键功能进行测试

运行`make check`以运行所有质量检查。

## 🐛 故障排除

### 常见问题

**"OPENAI_API_KEY environment variable not set"**
```bash
cp .env.example .env
# 编辑.env并添加你的密钥
```

**端口已被占用**
```bash
PORT=3000 make run-dev
```

**工具未加载**
- 检查`SKILLS_DIR`环境变量
- 验证MCP配置路径
- 检查日志中的错误消息

**构建错误**
```bash
make clean
make deps
make build
```

### 调试模式

启用详细日志：
```env
LOG_LEVEL=debug
```

## 📈 性能

- **会话加载**：延迟加载会话历史
- **工具初始化**：异步后台加载
- **内存管理**：基于LRU的会话缓存
- **并发请求**：基于goroutine的请求处理

## 🔒 安全

- 无用户认证（单用户模式）
- 仅本地存储（无云依赖）
- 输入验证和清理
- API访问的CORS配置

## 🗺️ 路线图

- [ ] 流式聊天响应
- [ ] 带身份验证的多用户支持
- [ ] 会话导出/导入功能
- [ ] 高级工具管理UI
- [ ] 语音输入/输出支持
- [ ] 自定义工具的插件系统
- [ ] 实时协作功能

## 📄 许可证

本项目是LangGraphGo的一部分，遵循相同的许可证。

## 🔗 了解更多

- [LangGraphGo文档](https://github.com/smallnest/langgraphgo)
- [Makefile指南](./Makefile.README.md)
- [LangChain Go](https://github.com/tmc/langchaingo)
- [MCP规范](https://modelcontextprotocol.io/)