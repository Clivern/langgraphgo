# API Interrupt Demo

This example demonstrates how to build a **request-response HTTP API** that handles conversational agents with checkpoint-based interrupt handling.

## What This Example Demonstrates

### 1. Issue #70 Fix: Auto-Save Checkpoints on Dynamic Interrupts

**Before the fix:** When using `graph.Interrupt()` for dynamic interrupts (human-in-the-loop), checkpoints were NOT automatically saved, making it difficult to resume conversations.

**After the fix:** Checkpoints are now automatically saved when a node calls `graph.Interrupt()`, preserving the state at the interrupt point.

The fix in `graph/state_graph.go:275-361` ensures that:
- `OnGraphStep` callback is invoked BEFORE returning on `NodeInterrupt`
- This triggers `CheckpointListener.OnGraphStep` which saves the checkpoint
- Regular errors still don't trigger checkpoint saves (preserving existing behavior)

### 2. Request-Response API Pattern

This example shows how to build an HTTP API where:
- Each HTTP request is independent (stateless server)
- Conversation state is persisted via checkpoints
- The API automatically detects if a request should resume from an interrupt

### 3. Resume Detection Logic

The key insight is using **checkpoint metadata** to determine if a request should resume:

```go
// Check if the latest checkpoint has interrupt metadata
if event, ok := latestCP.Metadata["event"].(string); ok && event == "step" {
    if state, ok := latestCP.State.(OrderState); ok && state.IsInterrupt {
        isResuming = true
    }
}
```

When resuming:
```go
config = &graph.Config{
    Configurable: map[string]any{
        "thread_id": threadID,
    },
    ResumeValue: req.Content,    // User's response
    ResumeFrom:  []string{latestCP.NodeName},  // Interrupted node
}
```

## How to Run

1. **Start the server:**
   ```bash
   go run main.go
   ```

2. **Test the conversation flow:**

   **Step 1: Start a new order**
   ```bash
   curl -X POST http://localhost:8080/chat \
     -H 'Content-Type: application/json' \
     -d '{"session_id":"user123","content":"我想买iPhone 15"}'
   ```

   Response:
   ```json
   {
     "message": "您购买的 iPhone 15，价格：7999.00 元\n请确认是否支付？（回复`确认`以完成支付）",
     "order_status": "待支付",
     "is_interrupt": true,
     "needs_resume": true
   }
   ```

   **Step 2: Confirm payment (resume from interrupt)**
   ```bash
   curl -X POST http://localhost:8080/chat \
     -H 'Content-Type: application/json' \
     -d '{"session_id":"user123","content":"确认"}'
   ```

   Response:
   ```json
   {
     "message": "您购买的 iPhone 15，价格：7999.00 元，已发货！\n订单号：ORDuser1231736123456",
     "order_status": "已发货",
     "is_interrupt": false,
     "needs_resume": false
   }
   ```

## Architecture

### Graph Structure

```
┌─────────────────┐
│ order_receive   │  Extract product info from user input
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ inventory_check │  Check if product is in stock
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ price_calc      │  Calculate price
└────────┬────────┘
         │
         ▼
┌─────────────────────┐
│ payment_processing  │  ⚡ INTERRUPT: Wait for user confirmation
└────────┬────────────┘
         │
         ▼
┌─────────────────┐
│ warehouse_notify│  Ship the order
└────────┬────────┘
         │
         ▼
        END
```

### State Management

**OrderState** contains:
- `SessionId`: Links requests to the same conversation
- `UserInput`: Latest user message
- `ProductInfo`, `OrderId`, `Price`: Order details
- `OrderStatus`: Current order state
- `NextNode`: Where to resume after interrupt
- `IsInterrupt`: Flag indicating interrupt occurred

### Checkpoint Flow

```
┌─────────────────────────────────────────────────────────────┐
│  User Request 1: "我想买iPhone 15"                          │
│  → New conversation (no checkpoints exist)                  │
│  → Execute from entry point                                 │
│  → order_receive → inventory_check → price_calc             │
│  → payment_processing calls Interrupt()                     │
│  → ✨ State is saved to checkpoint automatically ✨          │
│  → Return interrupt response to user                        │
└─────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────┐
│  User Request 2: "确认"                                     │
│  → Existing checkpoint found with interrupt metadata        │
│  → isResuming = true                                        │
│  → Load state from checkpoint                               │
│  → Set ResumeValue = "确认"                                 │
│  → Set ResumeFrom = ["payment_processing"]                  │
│  → Resume execution from interrupted node                   │
│  → Complete payment → warehouse_notify → END                │
│  → Return final response                                    │
└─────────────────────────────────────────────────────────────┘
```

## Key Code Patterns

### 1. Thread ID Management

Use `thread_id` to link all requests in the same conversation:

```go
// Use session_id as thread_id
threadID := req.SessionID

config := &graph.Config{
    Configurable: map[string]any{
        "thread_id": threadID,
    },
}
```

### 2. Checkpoint-Based Resume Detection

```go
checkpoints, err := s.Store.List(ctx, threadID)
if len(checkpoints) > 0 {
    latestCP := checkpoints[len(checkpoints)-1]

    // Check if this is an interrupt that needs resuming
    if state, ok := latestCP.State.(OrderState); ok && state.IsInterrupt {
        isResuming = true
    }
}
```

### 3. Resume Configuration

```go
if isResuming {
    config = &graph.Config{
        Configurable: map[string]any{
            "thread_id": threadID,
        },
        ResumeValue: req.Content,    // User's response to interrupt
        ResumeFrom:  []string{latestCP.NodeName},  // Resume point
    }
}
```

### 4. Interrupt Handling in Response

```go
var graphInterrupt *graph.GraphInterrupt
if errors.As(err, &graphInterrupt) {
    // Checkpoint was auto-saved by the framework (Issue #70 fix)
    response := ChatResponse{
        Message:     fmt.Sprintf("%v", graphInterrupt.InterruptValue),
        IsInterrupt: true,
        NeedsResume: true,
    }
    // Send response to client
}
```

## Comparing Approaches

### Approach 1: Auto-Save with Checkpoint Metadata (Recommended)

**Pros:**
- Framework handles checkpoint save automatically
- No manual checkpoint management in API code
- Clean separation of concerns

**Cons:**
- Requires framework fix (Issue #70)
- Depends on checkpoint metadata

### Approach 2: Manual Checkpoint Save

**Pros:**
- Works without framework fix
- Full control over checkpoint metadata

**Cons:**
- More boilerplate code
- Risk of forgetting to save checkpoint

**Example (before Issue #70 fix):**
```go
if errors.As(err, &graphInterrupt) {
    // Manually save checkpoint
    interruptState, _ := graphInterrupt.State.(OrderState)
    s.Runnable.SaveCheckpoint(ctx, graphInterrupt.Node, interruptState)
    // Send response...
}
```

### Approach 3: Custom State Fields

**Pros:**
- Simple to understand
- No dependency on checkpoint system

**Cons:**
- State bloat
- Doesn't work across server restarts
- Loss of time-travel capability

## Related Issues

- **Issue #70**: Checkpoints not saved on dynamic interrupts
  - Fixed by calling `OnGraphStep` before returning on `NodeInterrupt`
  - Checkpoint is saved with the merged state at interrupt point

- **Issue #67**: State modifications before `Interrupt()` were lost
  - Previously fixed by preserving node result on `NodeInterrupt`

## Testing

You can test this example manually or write automated tests:

```bash
# Test 1: New conversation
curl -X POST http://localhost:8080/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"test001","content":"我想买MacBook Pro"}'

# Test 2: Resume from interrupt
curl -X POST http://localhost:8080/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"test001","content":"确认"}'

# Test 3: Cancel order
curl -X POST http://localhost:8080/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"test002","content":"我想买AirPods"}'

curl -X POST http://localhost:8080/chat \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"test002","content":"不买"}'
```

## Production Considerations

1. **Checkpoint Storage**: Use Redis, PostgreSQL, or SQLite for production
2. **Cleanup**: Implement checkpoint TTL or cleanup logic
3. **Concurrency**: Ensure checkpoint store is thread-safe
4. **Error Handling**: Handle edge cases (corrupted checkpoints, missing state, etc.)
5. **Security**: Validate session IDs and user inputs
6. **Scalability**: Consider sharding checkpoints by user/session
