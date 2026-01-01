# 支付处理与动态中断

本示例演示了 [Issue #67](https://github.com/smallnest/langgraphgo/issues/67) 的修复，确保在调用 `graph.Interrupt()` 之前对状态的修改能够被正确保存。

## 问题描述 (Issue #67)

在修复之前，当节点修改状态后调用 `graph.Interrupt()` 时，状态修改会丢失。例如：

```go
func paymentNode(ctx context.Context, state OrderState) (OrderState, error) {
    // 修改状态
    state.PaymentStatus = "pending_payment"  // ❌ 这个修改会丢失！
    state.TransactionID = "TXN-123"          // ❌ 这个修改会丢失！

    // 然后中断
    _, err := graph.Interrupt(ctx, "确认支付？")
    return state, err
}
```

`GraphInterrupt.State` 中包含的是修改**之前**的状态，而不是修改**之后**的状态。

## 解决方案

修复后，状态修改会被自动保存：

```go
func paymentNode(ctx context.Context, state OrderState) (OrderState, error) {
    // 修改状态
    state.PaymentStatus = "pending_payment"  // ✅ 现在会被保存！
    state.TransactionID = "TXN-123"          // ✅ 现在会被保存！

    // 中断 - 状态修改已保存
    _, err := graph.Interrupt(ctx, "确认支付？")
    return state, err  // 即使有错误，状态也会被正确保存
}
```

## 场景说明

本示例模拟电商支付流程：

1. **初始化支付**：创建新的支付会话
2. **处理支付**：
   - 更新状态为 `"pending_payment"`（待支付）
   - 生成交易ID
   - **中断**以请求用户确认
3. **用户确认**：模拟用户批准支付
4. **恢复执行**：继续已确认的支付
5. **完成订单**：完成订单流程

## 关键演示点

本示例展示了：

1. ✅ `Interrupt()` 之前的状态修改被正确保存
2. ✅ 交易ID和支付状态在中断期间持久化
3. ✅ 恢复操作能够正确继续使用更新后的状态
4. ✅ 修复对类型化状态（struct）有效，不仅限于 `map[string]any`

## 运行示例

```bash
cd examples/payment_interrupt
go run main.go
```

## 预期输出

```
=== Payment Processing with Dynamic Interrupt Demo ===
This example demonstrates Issue #67 fix:
State modifications before Interrupt() are now correctly preserved.

🛒 Starting order: ORD-2024-001 for customer CUST-123
============================================================

--- Step 1: Initial Execution ---
📝 [init_payment] Initializing payment...
   Status: initialized

💳 [process_payment] Processing payment...
   Created transaction: TXN-ORD-2024-001-001
   Status updated to: pending_payment
   Amount: $99.99
   ⏸️  Interrupting to request user confirmation...

⚠️  Graph Interrupted!
   Node: process_payment
   Question: Please confirm payment of $99.99 via Credit Card

📊 State at Interruption:
   Order ID: ORD-2024-001
   Payment Status: pending_payment          ✅ 已保存！
   Transaction ID: TXN-ORD-2024-001-001    ✅ 已保存！
   Amount: $99.99

✅ SUCCESS: State modifications before Interrupt() were preserved!
   This confirms Issue #67 is fixed.

--- Step 2: User Confirmation ---
💬 Simulating user confirming payment...

--- Step 3: Resuming Execution ---
   ✅ Payment confirmed and completed

📦 [finalize_order] Finalizing order...
   Order ORD-2024-001 is ready for shipment

📊 Final State:
   Order ID: ORD-2024-001
   Payment Status: paid
   Transaction ID: TXN-ORD-2024-001-001
   Amount: $99.99

🎉 Order completed successfully!

============================================================
Demo completed!
```

## 修复内容

为了修复这个问题，进行了三处修改：

1. **`executeNodeWithRetry()`**：对于 `NodeInterrupt` 错误，返回节点的实际结果和错误，而不是零值
2. **`executeNodesParallel()`**：即使发生 `NodeInterrupt` 错误，也保存节点的结果
3. **`InvokeWithConfig()`**：在检查中断错误之前先合并状态更新

## 使用场景

此模式适用于：

- 💳 支付确认
- 📧 邮箱验证步骤
- ✅ 用户审批工作流
- 🔐 双因素认证
- 📝 需要用户修正的表单验证
- 🎫 预订确认

## 相关示例

- `examples/dynamic_interrupt/` - 基本的动态中断用法
- `examples/human_in_the_loop/` - 人机交互模式
- `examples/time_travel/` - 状态快照和时间旅行

## 技术细节

修复确保了当调用 `graph.Interrupt(ctx, value)` 时：

1. 节点的返回值（包括状态修改）被保留
2. `GraphInterrupt.State` 包含**更新后**的状态
3. 恢复操作使用正确的状态继续执行
4. 适用于类型化（`StateGraph[T]`）和非类型化（`StateGraph[map[string]any]`）图
