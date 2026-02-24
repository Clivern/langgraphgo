# PiAgent 实现总结

PiAgent 是在 `/Users/smallnest/ai/langgraphgo/prebuilt/pi_agent.go` 中实现的 Agent，灵感来自 pi-mono/packages/agent。

## 核心组件

### 1. PiAgentState - 与 pi-mono 的 AgentState 一致的状态结构

```go
type PiAgentState struct {
    // 核心状态
    SystemPrompt  string
    Model         string
    ThinkingLevel string  // off, minimal, low, medium, high, xhigh
    Messages      []llms.MessageContent
    Tools         []tools.Tool

    // 流式状态
    IsStreaming   bool
    StreamMessage *llms.MessageContent

    // 工具执行
    PendingToolCalls map[string]bool
    Error            error

    // 消息队列
    SteeringQueue     []llms.MessageContent
    SteeringMode      MessageQueueMode
    FollowUpQueue     []llms.MessageContent
    FollowUpMode      MessageQueueMode

    // 会话信息
    SessionKey string
}
```

### 2. PiAgentEvent - 完整的事件类型系统

| 事件类型 | 说明 |
|---------|------|
| `EventAgentStart` | Agent 开始 |
| `EventAgentEnd` | Agent 结束 |
| `EventTurnStart` | Turn 开始（一次 LLM 调用 + 工具执行） |
| `EventTurnEnd` | Turn 结束 |
| `EventMessageStart` | 消息开始 |
| `EventMessageUpdate` | 消息更新（流式） |
| `EventMessageEnd` | 消息结束 |
| `EventToolExecutionStart` | 工具执行开始 |
| `EventToolExecutionUpdate` | 工具执行更新（流式） |
| `EventToolExecutionEnd` | 工具执行结束 |

### 3. PiAgent - 高级 API

| 方法 | 说明 |
|------|------|
| `Subscribe(fn)` | 订阅事件，返回取消订阅函数 |
| `Prompt(msg)` | 发送提示并执行 |
| `PromptWithStream(msg)` | 发送提示并返回事件流 |
| `Steer(msg)` | 添加引导消息（中断） |
| `FollowUp(msg)` | 添加后续消息 |
| `Abort()` | 中断执行 |
| `Reset()` | 重置状态 |
| `WaitForIdle(ctx)` | 等待空闲 |

### 4. 构建器模式

- 使用 langgraphgo 的 `StateGraph` 构建执行图
- **agent 节点**：调用 LLM 生成响应
- **tools 节点**：执行工具调用
- **条件边**：根据是否有工具调用决定流向

## 使用示例

```go
import "github.com/smallnest/langgraphgo/prebuilt"

// 创建 PiAgent
agent, err := prebuilt.NewPiAgent(
    model,
    []tools.Tool{...},
    prebuilt.WithPiSystemPrompt("You are a helpful assistant"),
    prebuilt.WithPiMaxIterations(10),
    prebuilt.WithPiSteeringMode(prebuilt.QueueModeAll),
)

// 订阅事件
unsubscribe := agent.Subscribe(func(event prebuilt.PiAgentEvent) {
    switch event.Type {
    case prebuilt.EventAgentStart:
        fmt.Println("Agent started")
    case prebuilt.EventToolExecutionStart:
        fmt.Printf("Tool %s started\n", event.ToolName)
    case prebuilt.EventToolExecutionEnd:
        fmt.Printf("Tool %s ended: %v\n", event.ToolName, event.ToolResult)
    }
})
defer unsubscribe()

// 发送提示
msg := llms.TextParts(llms.ChatMessageTypeHuman, "What's the weather today?")
err = agent.Prompt(ctx, msg)

// 或使用流式执行
eventChan, errorChan, cancel := agent.PromptWithStream(ctx, msg)
defer cancel()

for {
    select {
    case event := <-eventChan:
        fmt.Printf("Event: %v\n", event.Type)
    case err := <-errorChan:
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }
        return
    }
}
```

## 与 pi-mono 的对应关系

| pi-mono (TypeScript) | langgraphgo (Go) |
|---------------------|------------------|
| `Agent` class | `PiAgent` struct |
| `AgentState` interface | `PiAgentState` struct |
| `AgentEvent` type | `PiAgentEvent` struct |
| `agentLoop()` function | `buildPiAgentGraph()` + StateRunnable |
| `EventStream` | `InvokeWithListener()` + channel |
| `steer()` / `followUp()` | `Steer()` / `FollowUp()` |
| `QueueMode` ("all" / "one-at-a-time") | `MessageQueueMode` (QueueModeAll / QueueModeOneAtATime) |
| `subscribe(fn)` | `Subscribe(func) func()` |
| `prompt(message)` | `Prompt(msg MessageContent)` |
| `abort()` | `Abort()` |
| `reset()` | `Reset()` |
| `waitForIdle()` | `WaitForIdle(ctx) error` |

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                         PiAgent                              │
│  (高级API，管理状态、队列、事件订阅)                         │
└──────────────────────┬──────────────────────────────────────┘
                       │ 调用
                       ▼
┌─────────────────────────────────────────────────────────────┐
│              StateRunnable[*PiAgentState]                    │
│  (langgraphgo 的状态图执行器)                                │
└──────────────────────┬──────────────────────────────────────┘
                       │
       ┌───────────────┼───────────────┐
       ▼               ▼               ▼
┌──────────────┐ ┌──────────────┐ ┌──────────────┐
│ agent 节点   │ │ tools 节点   │ │ 条件边       │
│ (LLM调用)    │ │ (工具执行)   │ │ (决策流向)   │
└──────────────┘ └──────────────┘ └──────────────┘
```

## 执行流程

```
1. 用户调用 Prompt() 发送消息
   ↓
2. agent 节点：
   - 应用 transformContext (可选)
   - 应用 convertToLLM (可选)
   - 调用 LLM 生成响应
   - 检查是否有工具调用
   ↓
3. 条件边判断：
   - 有工具调用 → 转到 tools 节点
   - 无工具调用 → 结束 (END)
   ↓
4. tools 节点：
   - 执行每个工具调用
   - 发送工具执行事件
   - 检查 steering 消息
   - 返回工具结果
   ↓
5. 回到 agent 节点继续（如果有更多工具调用）
```

## 消息队列机制

### Steering 消息（引导消息）
- 用于中断正在运行的 agent
- 通过 `Steer()` 方法添加
- 在工具执行后检查，如果有则跳过剩余工具
- 支持两种模式：
  - `QueueModeAll`: 一次性返回所有消息
  - `QueueModeOneAtATime`: 每次返回一条消息

### FollowUp 消息（后续消息）
- 在 agent 完成后处理
- 通过 `FollowUp()` 方法添加
- 用于添加后续任务或追问
- 支持两种模式（同上）

## 配置选项

| 选项 | 说明 |
|------|------|
| `WithPiSystemPrompt(prompt)` | 设置系统提示词 |
| `WithPiThinkingLevel(level)` | 设置思考级别 |
| `WithPiSteeringMode(mode)` | 设置引导消息模式 |
| `WithPiFollowUpMode(mode)` | 设置后续消息模式 |
| `WithPiStreamMode(mode)` | 设置流模式 |
| `WithPiConvertToLLM(fn)` | 设置消息转换函数 |
| `WithPiTransformContext(fn)` | 设置上下文转换函数 |
| `WithPiMaxIterations(max)` | 设置最大迭代次数 |

## 与 CreateAgent 的区别

| 特性 | CreateAgent | PiAgent |
|------|-------------|---------|
| 状态类型 | `map[string]any` | `*PiAgentState` (结构化) |
| 事件系统 | 无 | 完整的 `PiAgentEvent` |
| 消息队列 | 无 | Steering + FollowUp |
| 流式支持 | `StreamingRunnable` | `PromptWithStream()` |
| API 风格 | 函数式 | 面向对象 |
| 灵感来源 | LangChain | pi-mono/packages/agent |
