# PiAgent Example

This example demonstrates the usage of `PiAgent`, inspired by [pi-mono/packages/agent](https://github.com/mariozechner/pi-mono/tree/main/packages/agent).

## Prerequisites

- Go 1.21+
- OpenAI API key

## Running

```bash
cd examples/pi_agent
export OPENAI_API_KEY="your-key-here"
go run main.go
```

## What It Demonstrates

- **Agent Creation** - Creating an agent with tools
- **Event Subscription** - Listening to agent lifecycle events
- **Tool Execution** - Agent using tools during execution
- **State Management** - Accessing agent state after execution

## Example Output

```
=== Example: Simple Calculation ===
=== Agent Started ===
--- Turn Started ---
Tool Started: calculator
Tool Completed: calculator
--- Turn Ended ---
=== Agent Ended ===

=== Example: Multiple Calculations ===
=== Agent Started ===
--- Turn Started ---
Tool Started: calculator
Tool Completed: calculator
--- Turn Ended ---
=== Agent Ended ===

=== Final State ===
Total messages: 8
Tools: 1
Thinking Level: off
```

## Code Overview

```go
// Create OpenAI LLM
model, err := openai.New()

// Create agent with tools
agent, err := prebuilt.NewPiAgent(
    model,
    []tools.Tool{calculator},
    prebuilt.WithPiSystemPrompt("You are a helpful assistant."),
)

// Subscribe to events
agent.Subscribe(func(event prebuilt.PiAgentEvent) {
    // Handle events
})

// Send prompt
msg := llms.TextParts(llms.ChatMessageTypeHuman, "What is 25 + 17?")
agent.Prompt(ctx, msg)
```

## Tool Implementation

The example includes a simple `CalculatorTool` that demonstrates:

```go
type CalculatorTool struct{}

func (t *CalculatorTool) Name() string {
    return "calculator"
}

func (t *CalculatorTool) Call(ctx context.Context, input string) (string, error) {
    // Execute calculation
    return result, nil
}

func (t *CalculatorTool) Schema() map[string]any {
    // Return JSON Schema for parameters
    return map[string]any{
        "type": "object",
        "properties": map[string]any{
            "expression": map[string]any{
                "type": "string",
                "description": "The expression to evaluate",
            },
        },
        "required": []string{"expression"},
    }
}
```

## Event Types

| Event | Description |
|-------|-------------|
| `EventAgentStart` | Agent started |
| `EventAgentEnd` | Agent ended |
| `EventTurnStart` | Turn started (LLM call + tool execution) |
| `EventTurnEnd` | Turn ended |
| `EventToolExecutionStart` | Tool started |
| `EventToolExecutionEnd` | Tool ended |
