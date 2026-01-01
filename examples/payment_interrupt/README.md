# Payment Processing with Dynamic Interrupt

This example demonstrates the fix for [Issue #67](https://github.com/smallnest/langgraphgo/issues/67), which ensures that state modifications made before calling `graph.Interrupt()` are correctly preserved.

## The Problem (Issue #67)

Before the fix, when a node modified state and then called `graph.Interrupt()`, the state modifications were lost. For example:

```go
func paymentNode(ctx context.Context, state OrderState) (OrderState, error) {
    // Modify state
    state.PaymentStatus = "pending_payment"  // ❌ This was lost!
    state.TransactionID = "TXN-123"          // ❌ This was lost!

    // Then interrupt
    _, err := graph.Interrupt(ctx, "Confirm payment?")
    return state, err
}
```

The `GraphInterrupt.State` would contain the state **before** the modifications, not **after**.

## The Solution

After the fix, state modifications are automatically preserved:

```go
func paymentNode(ctx context.Context, state OrderState) (OrderState, error) {
    // Modify state
    state.PaymentStatus = "pending_payment"  // ✅ Now preserved!
    state.TransactionID = "TXN-123"          // ✅ Now preserved!

    // Interrupt - state changes are saved
    _, err := graph.Interrupt(ctx, "Confirm payment?")
    return state, err  // State is correctly saved even with error
}
```

## Scenario

This example simulates an e-commerce payment flow:

1. **Initialize Payment**: Create a new payment session
2. **Process Payment**:
   - Update state to `"pending_payment"`
   - Generate a transaction ID
   - **Interrupt** to request user confirmation
3. **User Confirmation**: Simulate user approving the payment
4. **Resume Execution**: Continue with confirmed payment
5. **Finalize Order**: Complete the order

## Key Demonstration Points

The example shows:

1. ✅ State modifications before `Interrupt()` are preserved
2. ✅ Transaction IDs and payment status persist across interruption
3. ✅ Resume correctly continues with the updated state
4. ✅ The fix works with typed state (struct), not just `map[string]any`

## Running the Example

```bash
cd examples/payment_interrupt
go run main.go
```

## Expected Output

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
   Payment Status: pending_payment          ✅ PRESERVED!
   Transaction ID: TXN-ORD-2024-001-001    ✅ PRESERVED!
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

## What Was Fixed

Three changes were made to fix this issue:

1. **`executeNodeWithRetry()`**: For `NodeInterrupt` errors, return the node's result along with the error instead of a zero value
2. **`executeNodesParallel()`**: Save the result even when a `NodeInterrupt` error occurs
3. **`InvokeWithConfig()`**: Merge state updates before checking for interrupt errors

## Use Cases

This pattern is useful for:

- 💳 Payment confirmations
- 📧 Email verification steps
- ✅ User approval workflows
- 🔐 Two-factor authentication
- 📝 Form validation with user corrections
- 🎫 Reservation confirmations

## Related Examples

- `examples/dynamic_interrupt/` - Basic dynamic interrupt usage
- `examples/human_in_the_loop/` - Human-in-the-loop patterns
- `examples/time_travel/` - State snapshots and time travel

## Technical Details

The fix ensures that when `graph.Interrupt(ctx, value)` is called:

1. The node's return value (including state modifications) is preserved
2. The `GraphInterrupt.State` contains the **updated** state
3. Resume operations continue with the correct state
4. Works with both typed (`StateGraph[T]`) and untyped (`StateGraph[map[string]any]`) graphs
