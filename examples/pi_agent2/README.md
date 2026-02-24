# PiAgent OpenAI Example

This example demonstrates `PiAgent` with a real OpenAI LLM integration.

## Prerequisites

- Go 1.21+
- OpenAI API key

## Running

```bash
cd examples/pi_agent2
export OPENAI_API_KEY="your-key-here"
go run main.go
```

## What It Demonstrates

- **Real LLM Integration** - Using OpenAI's GPT model
- **Tool Usage** - Agent using tools during execution
- **Event Subscription** - Real-time event feedback with emoji indicators
- **Conversation History** - Accessing final messages

## Code Overview

```go
// Create OpenAI LLM
model, err := openai.New(
    openai.WithToken(os.Getenv("OPENAI_API_KEY")),
    openai.WithModel("gpt-4o-mini"),
)

// Create PiAgent with tools
agent, err := prebuilt.NewPiAgent(
    model,
    []tools.Tool{searchTool},
    prebuilt.WithPiSystemPrompt("You are a helpful research assistant."),
)

// Subscribe to events
agent.Subscribe(func(event prebuilt.PiAgentEvent) {
    // Handle events in real-time
})

// Send prompt
msg := llms.TextParts(llms.ChatMessageTypeHuman, "What's the latest news?")
agent.Prompt(ctx, msg)
```

## Event Indicators

| Event | Indicator |
|-------|-----------|
| Agent Start | 🤖 |
| Agent End | ✅ |
| Turn Start | 🔄 |
| Turn End | ✓ |
| Tool Start | 🔧 |
| Tool Done | ✓ |
| Tool Error | ❌ |

## Customization

### Change Model

```go
model, err := openai.New(
    openai.WithToken(os.Getenv("OPENAI_API_KEY")),
    openai.WithModel("gpt-4"),  // or "gpt-4-turbo", "gpt-3.5-turbo"
)
```

### Add More Tools

```go
agent, err := prebuilt.NewPiAgent(
    model,
    []tools.Tool{searchTool, calculatorTool, weatherTool},
    prebuilt.WithPiSystemPrompt("You are a helpful assistant with access to multiple tools."),
)
```

### Adjust Iterations

```go
agent, err := prebuilt.NewPiAgent(
    model,
    []tools.Tool{...},
    prebuilt.WithPiMaxIterations(10),  // Allow more tool calls
)
```
