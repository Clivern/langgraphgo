# LangGraphGo Chat Application

A sophisticated web-based multi-session chat application with AI agent integration, tool support, and persistent local storage.

## ✨ Features

- 🔄 **Multi-Session Support**: Create and manage multiple independent chat sessions
- 💾 **Persistent Storage**: All conversations automatically saved to local disk
- 🌐 **Modern Web Interface**: Clean, responsive web UI with real-time updates
- 🤖 **AI Chat Agent**: Advanced agent with conversation history management
- 🔧 **Tool Integration**: Support for Skills and MCP (Model Context Protocol) tools
- 🔌 **Multi-Provider Support**: Works with OpenAI, Baidu, Azure, and any OpenAI-compatible API
- 🎨 **Beautiful UI**: Dark/light theme support with smooth animations
- 📝 **Session Management**: Create, view, clear, and delete sessions
- ⚡ **Hot Reload**: Development mode with automatic code reloading
- 🐳 **Docker Support**: Containerized deployment ready

## 🏗️ Architecture

```
showcases/chat/
├── main.go                 # Application entry point and server bootstrap
├── pkg/                    # Go packages
│   ├── chat/              # Chat server and agent logic
│   │   └── chat.go        # Core chat functionality
│   └── session/           # Session management
│       └── session.go     # Session persistence
├── static/
│   ├── index.html        # Web frontend
│   ├── style.css         # UI styles
│   └── script.js         # Frontend logic
├── sessions/             # Local session storage (auto-created)
├── build/                # Build output directory
├── Makefile              # Build automation
├── Dockerfile            # Docker configuration
├── .air.toml            # Hot reload configuration
├── go.mod
├── go.sum
├── .env                 # Configuration (create from .env.example)
└── README.md
```

## 🚀 Quick Start

### Option 1: Using Makefile (Recommended)

```bash
# Clone and navigate to the project
cd showcases/chat

# Install development tools
make setup-dev

# Copy environment template
cp .env.example .env

# Edit .env and add your OpenAI API key
# OPENAI_API_KEY=sk-...

# Run with hot reload (development mode)
make dev

# Or run normally
make run-dev
```

### Option 2: Standard Go Commands

```bash
cd showcases/chat

# Install dependencies
go mod download

# Copy environment template
cp .env.example .env

# Edit .env and add your OpenAI API key
# OPENAI_API_KEY=sk-...

# Build and run
go run main.go
```

The server will start at `http://localhost:8080`

## 🛠️ Development Workflow

### Using Makefile

```bash
# Install development tools (air, golangci-lint, etc.)
make setup-dev

# Run with hot reload
make dev

# Run all checks (format, lint, vet, test)
make check

# Build for production
make build

# Build for all platforms
make build-all
```

### Common Makefile Targets

| Target | Description |
|--------|-------------|
| `make dev` | Run with hot reload |
| `make run-dev` | Run with dev environment |
| `make build` | Build the application |
| `make test` | Run tests |
| `make coverage` | Run tests with coverage |
| `make format` | Format code |
| `make vet` | Vet code |
| `make lint` | Lint code |
| `make docker-up` | Build and run Docker |
| `make clean` | Clean build artifacts |
| `make help` | Show all targets |

## ⚙️ Configuration

Environment variables (in `.env`):

```env
# Required: Your API key
OPENAI_API_KEY=your-api-key-here

# Optional: Model name (default: gpt-4o-mini)
OPENAI_MODEL=gpt-4o-mini

# Optional: Base URL for OpenAI-compatible APIs
# Examples:
#   Baidu: https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions
#   Azure: https://your-resource.openai.azure.com/
#   Ollama: http://localhost:11434/v1
OPENAI_BASE_URL=

# Optional: Server port (default: 8080)
PORT=8080

# Optional: Session storage directory (default: ./sessions)
SESSION_DIR=./sessions

# Optional: Maximum messages per session (default: 50)
MAX_HISTORY_SIZE=50

# Optional: Skills directory (for tool integration)
SKILLS_DIR=../../testdata/skills

# Optional: MCP configuration path
MCP_CONFIG_PATH=../../testdata/mcp/mcp.json

# Optional: Chat title
CHAT_TITLE=LangGraphGo Chat
```

### LLM Provider Examples

**OpenAI**:
```env
OPENAI_API_KEY=sk-your-openai-key
OPENAI_MODEL=gpt-4o
```

**Baidu Qianfan**:
```env
OPENAI_API_KEY=your-baidu-token
OPENAI_BASE_URL=https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/completions
OPENAI_MODEL=ERNIE-Bot
```

**Azure OpenAI**:
```env
OPENAI_API_KEY=your-azure-key
OPENAI_BASE_URL=https://your-resource.openai.azure.com/
OPENAI_MODEL=your-deployment-name
```

**Local Models (Ollama, LM Studio)**:
```env
OPENAI_API_KEY=not-needed
OPENAI_BASE_URL=http://localhost:11434/v1
OPENAI_MODEL=llama2
```

## 📡 API Endpoints

### Sessions
- `POST /api/sessions/new` - Create a new session
- `GET /api/sessions` - List all sessions
- `DELETE /api/sessions/:id` - Delete a session
- `GET /api/sessions/:id/history` - Get session messages
- `GET /api/client-id` - Get current client ID

### Chat
- `POST /api/chat` - Send a message
  ```json
  {
    "session_id": "uuid",
    "message": "your message",
    "user_settings": {
      "enable_skills": true,
      "enable_mcp": true
    }
  }
  ```
  Response:
  ```json
  {
    "response": "AI response text"
  }
  ```

### Tools
- `GET /api/mcp/tools?session_id=:id` - List available MCP tools
- `GET /api/tools/hierarchical?session_id=:id` - Get tools in hierarchical structure
- `GET /api/config` - Get chat configuration

## 🧩 Components

### ChatAgent

The `SimpleChatAgent` provides:
- Automatic conversation context management
- Tool integration (Skills and MCP)
- Support for OpenAI-compatible APIs
- Thread-safe conversation history
- Asynchronous tool loading

### Session Management

Each session includes:
- Unique UUID identifier
- Complete message history
- Persistent JSON storage
- Client-based isolation
- Automatic saving and loading

### Tool Integration

The application supports two types of tools:

1. **Skills**: Pre-defined tool packages loaded from `SKILLS_DIR`
2. **MCP Tools**: Dynamic tools from Model Context Protocol servers

Tools can be enabled/disabled per session via user settings.

## 🐳 Docker Deployment

```bash
# Build and run with Docker Compose
make docker-up

# Or manually:
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

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific test
go test ./pkg/session -v
```

## 📦 Building

### Build for Current Platform
```bash
make build
```

### Cross-Platform Builds
```bash
# Build for all platforms
make build-all

# Build for specific platforms
make build-linux
make build-darwin
make build-windows
```

### Release Packages
```bash
# Create release packages
make release
```

Outputs will be in `build/release/`.

## 🔧 Customization

### Change System Prompt

Edit `pkg/chat/chat.go` in the `NewSimpleChatAgent` function:
```go
systemMsg := llms.MessageContent{
    Role:  llms.ChatMessageTypeSystem,
    Parts: []llms.ContentPart{llms.TextPart("Your custom system message here")},
}
```

### Add Custom Tools

1. Create a skill package in your skills directory
2. Follow the skill package structure from the examples
3. Tools will be automatically loaded

### Modify UI

Edit files in `static/`:
- `index.html` - Main HTML structure
- `style.css` - Styles and themes
- `script.js` - Frontend logic

## 🔍 Development

### Project Structure

- **main.go**: Application entry point, bootstrap, and graceful shutdown
- **pkg/chat/**: Core chat functionality and HTTP handlers
- **pkg/session/**: Session persistence and management
- **static/**: Web frontend assets
- **Makefile**: Build automation and development workflow

### Adding Features

1. **New API endpoints**: Add to `pkg/chat/chat.go`
2. **New session fields**: Update `pkg/session/session.go`
3. **Frontend changes**: Modify `static/` files
4. **Configuration**: Add to environment variables

### Code Quality

The project uses:
- `go fmt` for formatting
- `go vet` for static analysis
- `golangci-lint` for comprehensive linting
- Tests for critical functionality

Run `make check` to run all quality checks.

## 🐛 Troubleshooting

### Common Issues

**"OPENAI_API_KEY environment variable not set"**
```bash
cp .env.example .env
# Edit .env and add your key
```

**Port already in use**
```bash
PORT=3000 make run-dev
```

**Tools not loading**
- Check `SKILLS_DIR` environment variable
- Verify MCP configuration path
- Check logs for error messages

**Build errors**
```bash
make clean
make deps
make build
```

### Debug Mode

Enable verbose logging:
```env
LOG_LEVEL=debug
```

## 📈 Performance

- **Session Loading**: Lazy loading of session history
- **Tool Initialization**: Asynchronous background loading
- **Memory Management**: LRU-based session caching
- **Concurrent Requests**: Goroutine-based request handling

## 🔒 Security

- No user authentication (single-user mode)
- Local storage only (no cloud dependencies)
- Input validation and sanitization
- CORS configuration for API access

## 🗺️ Roadmap

- [ ] Streaming chat responses
- [ ] Multi-user support with authentication
- [ ] Session export/import functionality
- [ ] Advanced tool management UI
- [ ] Voice input/output support
- [ ] Plugin system for custom tools
- [ ] Real-time collaboration features

## 📄 License

This project is part of LangGraphGo and follows the same license.

## 🔗 Learn More

- [LangGraphGo Documentation](https://github.com/smallnest/langgraphgo)
- [Makefile Guide](./Makefile.README.md)
- [LangChain Go](https://github.com/tmc/langchaingo)
- [MCP Specification](https://modelcontextprotocol.io/)