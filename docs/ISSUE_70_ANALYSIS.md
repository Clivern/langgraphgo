# Issue #70 分析：中断和 Checkpoint 问题

## 问题 1: 自动保存的 checkpoint 状态不是最新的

### 根本原因

通过分析 `graph/state_graph.go:260-340` 的执行流程，发现了问题所在:

**执行流程:**
1. 执行节点 (line 262)
2. 处理结果 (line 265)
3. **合并状态** (line 269) - 包含中断前的状态修改
4. **检查 NodeInterrupt** (line 276-288) - 如果发现中断，立即返回 GraphInterrupt
5. *(只有没有中断时才继续)*
6. **调用 OnGraphStep 回调** (line 315-328) - **这里才保存 checkpoint**

**问题:**
- 当使用 `graph.Interrupt()` 动态中断时，代码在第 4 步就返回了
- 第 6 步的 `OnGraphStep` 回调**永远不会被调用**
- 因此使用 `AutoSave: true` 时，**不会自动保存包含中断状态的 checkpoint**

**对比 InterruptAfter:**
- 使用 `InterruptAfter` 时，检查发生在第 6 步之后（line 330-339）
- 所以 `OnGraphStep` 会被调用，checkpoint 会被自动保存

### 解决方案

有两种方法可以解决这个问题:

#### 方法 1: 手动保存 checkpoint (推荐用于当前版本)

在捕获到 `GraphInterrupt` 后，手动保存 checkpoint:

```go
result, err := h.Runnable.InvokeWithConfig(context.Background(), lastState, config)

var graphInterrupt *graph.GraphInterrupt
if errors.As(err, &graphInterrupt) {
    // 获取中断时的状态
    interruptState, _ := graphInterrupt.State.(workflow.OrderState)

    // 手动保存 checkpoint
    threadID := config.Configurable["thread_id"].(string)
    h.Runnable.SaveCheckpoint(context.Background(), graphInterrupt.Node, interruptState)

    // 更新 listener 的 threadID 以确保保存到正确的 thread
    // (如果使用 threadID 而不是 executionID)

    SendResponse(c, nil, graphInterrupt.InterruptValue)
    return
}
```

#### 方法 2: 修复框架代码 (需要提交 PR)

在 `graph/state_graph.go` 的 `InvokeWithConfig` 中，在返回 `GraphInterrupt` 之前调用 `OnGraphStep`:

```go
// 在 line 278-288 之间，修改为:
if errors.As(err, &nodeInterrupt) {
    // 在返回前调用 OnGraphStep 以保存 checkpoint
    if config != nil && len(config.Callbacks) > 0 {
        for _, cb := range config.Callbacks {
            if gcb, ok := cb.(GraphCallbackHandler); ok {
                gcb.OnGraphStep(ctx, nodeInterrupt.Node, state)
            }
        }
    }

    // Return GraphInterrupt with the merged state
    return state, &GraphInterrupt{
        Node:           nodeInterrupt.Node,
        State:          state,
        InterruptValue: nodeInterrupt.Value,
        NextNodes:      []string{nodeInterrupt.Node},
    }
}
```

---

## 问题 2: 如何在请求-响应模式下判断是否需要传递 ResumeValue 和 ResumeFrom

### 核心概念

在请求-响应 API 场景中，每次请求都是独立的，需要从持久化的 checkpoint 中恢复状态。关键是要区分：

1. **首次请求** - 开始新的对话流程
2. **恢复请求** - 从中断点继续执行

### 推荐方案：使用 Checkpoint 元数据判断

最可靠的方法是检查最新 checkpoint 的元数据：

```go
func (h *Handler) Demo2(c *gin.Context) {
    var req demoRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        SendResponse(c, errno.ErrBind, nil)
        return
    }

    ctx := context.Background()
    threadID := req.SessionId // 使用 session ID 作为 thread_id

    // 获取该 thread 的最新 checkpoint
    checkpoints, err := h.Runnable.ListCheckpoints(ctx)
    if err != nil {
        log.Printf("Failed to list checkpoints: %v", err)
    }

    var config *graph.Config
    var initialState workflow.OrderState
    var isResuming bool

    if len(checkpoints) > 0 {
        // 找到该 thread 的最新 checkpoint
        var latestCP *graph.Checkpoint
        for _, cp := range checkpoints {
            if cp.Metadata["thread_id"] == threadID {
                if latestCP == nil || cp.Version > latestCP.Version {
                    latestCP = cp
                }
            }
        }

        if latestCP != nil {
            // 检查 checkpoint 是否包含中断信息
            if event, ok := latestCP.Metadata["event"].(string); ok && event == "interrupt" {
                isResuming = true
                // 从 checkpoint 恢复状态
                initialState, _ = agent.ConvertStateToStruct[workflow.OrderState](latestCP.State)
                // 设置用户新输入
                initialState.UserInput = req.Content

                // 配置恢复参数
                config = &graph.Config{
                    Configurable: map[string]any{
                        "thread_id": threadID,
                    },
                    ResumeValue: req.Content, // 用户的响应作为 resume value
                    ResumeFrom:  []string{latestCP.NodeName}, // 从中断的节点恢复
                }
            }
        }
    }

    if !isResuming {
        // 首次请求，初始化新状态
        initialState = workflow.InitState(req.SessionId)
        initialState.UserInput = req.Content

        config = &graph.Config{
            Configurable: map[string]any{
                "thread_id": threadID,
            },
        }
    }

    // 执行图
    result, err := h.Runnable.InvokeWithConfig(ctx, initialState, config)

    var graphInterrupt *graph.GraphInterrupt
    if errors.As(err, &graphInterrupt) {
        // 保存中断状态到 checkpoint (手动保存，因为 AutoSave 不会保存)
        interruptState, _ := graphInterrupt.State.(workflow.OrderState)

        // 手动保存 checkpoint 并标记为中断
        checkpoint := &store.Checkpoint{
            ID:        fmt.Sprintf("checkpoint_%s_%d", threadID, time.Now().UnixNano()),
            NodeName:  graphInterrupt.Node,
            State:     interruptState,
            Timestamp: time.Now(),
            Version:   getNextVersion(checkpoints, threadID),
            Metadata: map[string]any{
                "thread_id": threadID,
                "event":     "interrupt", // 标记为中断事件
                "interrupt_value": graphInterrupt.InterruptValue,
            },
        }
        h.Runnable.GetCheckpointStore().Save(ctx, checkpoint)

        SendResponse(c, nil, graphInterrupt.InterruptValue)
        return
    }

    if err != nil {
        SendResponse(c, errno.InternalServerError, nil)
        return
    }

    SendResponse(c, nil, result.Message)
}

func getNextVersion(checkpoints []*graph.Checkpoint, threadID string) int {
    maxVersion := 0
    for _, cp := range checkpoints {
        if cp.Metadata["thread_id"] == threadID && cp.Version > maxVersion {
            maxVersion = cp.Version
        }
    }
    return maxVersion + 1
}
```

### 替代方案：使用自定义状态字段

如果不想依赖 checkpoint 元数据，可以在状态中添加标记字段：

```go
type OrderState struct {
    OrderID       string
    ProductInfo   string
    Price         float64
    OrderStatus   string
    Message       string
    UserInput     string
    UpdateAt      time.Time
    SessionId     int

    // 添加中断状态管理字段
    IsInterrupted bool   `json:"is_interrupted"`
    InterruptNode string `json:"interrupt_node,omitempty"`
}
```

然后在中断处理中设置这些字段：

```go
g.AddNode("payment_processing", "payment_processing", func(ctx context.Context, state OrderState) (OrderState, error) {
    state.OrderStatus = "待支付"

    confirmMsg := fmt.Sprintf("您购买的 %s，价格：%.2f 元\n请确认是否支付？（回复`确认`以完成支付）",
        state.ProductInfo, state.Price)

    payInfo, err := graph.Interrupt(ctx, confirmMsg)
    if err != nil {
        // 标记为中断状态
        state.IsInterrupted = true
        state.InterruptNode = "payment_processing"
        return state, err
    }

    // 恢复时，清除中断标记
    state.IsInterrupted = false
    state.InterruptNode = ""

    // 处理用户确认...
    payInfoStr, ok := payInfo.(string)
    if !ok || !strings.Contains(strings.ToLower(payInfoStr), "确认") {
        state.Message = "您已取消支付，订单已关闭"
        state.OrderStatus = "已取消"
        state.OrderId = ""
        return state, nil
    }

    state.OrderStatus = "已支付"
    state.UpdateAt = time.Now()
    return state, nil
})
```

API handler:

```go
if lastState.IsInterrupted {
    // 这是恢复请求
    config = &graph.Config{
        Configurable: map[string]any{
            "thread_id": threadID,
        },
        ResumeValue: req.Content,
        ResumeFrom:  []string{lastState.InterruptNode},
    }
} else {
    // 这是新请求
    config = &graph.Config{
        Configurable: map[string]any{
            "thread_id": threadID,
        },
    }
}
```

---

## 总结与建议

### 关于问题 1 (Checkpoint 不保存)

**短期解决方案:**
- 在捕获 `GraphInterrupt` 后手动调用 `SaveCheckpoint`
- 在 checkpoint 元数据中标记 `"event": "interrupt"`

**长期解决方案:**
- 向项目提交 PR，在 `state_graph.go` 中修复，使动态中断也能触发 `OnGraphStep` 回调

### 关于问题 2 (如何判断是否需要 Resume)

**推荐方案:**
- 使用 checkpoint 元数据中的 `"event"` 字段判断
- 元数据包含 `"event": "interrupt"` 表示需要恢复
- 使用 `checkpoint.NodeName` 作为 `ResumeFrom` 的值
- 使用用户输入作为 `ResumeValue`

**关键要点:**
1. 每次请求都应该传递相同的 `thread_id`（如 session ID）
2. 检查该 thread 的最新 checkpoint
3. 如果 checkpoint 表示中断，则设置 `ResumeValue` 和 `ResumeFrom`
4. 手动保存中断时的 checkpoint，并在元数据中标记

### 代码改进建议

您的 `BuildGraph()` 代码整体结构良好，但建议：

1. 在每个可能中断的节点中，中断前的状态修改已经能正确保存（Issue #67 已修复）
2. 在 API handler 中手动保存中断 checkpoint
3. 使用 checkpoint 元数据来判断是否需要恢复，而不是自定义状态字段（更可靠）
4. 考虑添加 checkpoint 清理逻辑，避免过期的 checkpoint 累积

希望这些分析和建议能帮助您解决问题！
