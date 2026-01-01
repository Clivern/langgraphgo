package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/smallnest/langgraphgo/graph"
)

// OrderState represents the state of an order with payment processing
type OrderState struct {
	OrderID       string
	Amount        float64
	PaymentStatus string
	PaymentMethod string
	TransactionID string
	CustomerID    string
	Timestamp     string
}

func main() {
	fmt.Println("=== Payment Processing with Dynamic Interrupt Demo ===")
	fmt.Println("This example demonstrates Issue #67 fix:")
	fmt.Println("State modifications before Interrupt() are now correctly preserved.\n")

	// Create a typed state graph
	g := graph.NewStateGraph[OrderState]()

	// Node 1: Initialize payment
	g.AddNode("init_payment", "Initialize payment", func(ctx context.Context, state OrderState) (OrderState, error) {
		fmt.Println("📝 [init_payment] Initializing payment...")
		state.PaymentStatus = "initialized"
		state.Timestamp = "2024-01-01T10:00:00Z"
		fmt.Printf("   Status: %s\n", state.PaymentStatus)
		return state, nil
	})

	// Node 2: Process payment - this is where we modify state and then interrupt
	g.AddNode("process_payment", "Process payment and await confirmation", func(ctx context.Context, state OrderState) (OrderState, error) {
		fmt.Println("\n💳 [process_payment] Processing payment...")

		// CRITICAL: Modify state BEFORE interrupting
		// This simulates a payment system that creates a pending transaction
		state.PaymentStatus = "pending_payment"
		state.TransactionID = "TXN-" + state.OrderID + "-001"

		fmt.Printf("   Created transaction: %s\n", state.TransactionID)
		fmt.Printf("   Status updated to: %s\n", state.PaymentStatus)
		fmt.Printf("   Amount: $%.2f\n", state.Amount)

		// Now interrupt to get user confirmation
		// Before the fix (Issue #67), these state changes would be lost!
		fmt.Println("   ⏸️  Interrupting to request user confirmation...")

		confirmationMsg := fmt.Sprintf("Please confirm payment of $%.2f via %s",
			state.Amount, state.PaymentMethod)

		userResponse, err := graph.Interrupt(ctx, confirmationMsg)
		if err != nil {
			// Return the modified state along with the interrupt error
			return state, err
		}

		// If resumed and user confirmed
		if userResponse != nil {
			confirmed, ok := userResponse.(bool)
			if !ok || !confirmed {
				state.PaymentStatus = "cancelled"
				fmt.Println("   ❌ Payment cancelled by user")
				return state, nil
			}

			// User confirmed - complete the payment
			state.PaymentStatus = "paid"
			fmt.Println("   ✅ Payment confirmed and completed")
		}

		return state, nil
	})

	// Node 3: Finalize order
	g.AddNode("finalize_order", "Finalize order", func(ctx context.Context, state OrderState) (OrderState, error) {
		fmt.Println("\n📦 [finalize_order] Finalizing order...")
		if state.PaymentStatus == "paid" {
			fmt.Printf("   Order %s is ready for shipment\n", state.OrderID)
		} else {
			fmt.Printf("   Order %s requires manual review (status: %s)\n",
				state.OrderID, state.PaymentStatus)
		}
		return state, nil
	})

	// Build the graph
	g.SetEntryPoint("init_payment")
	g.AddEdge("init_payment", "process_payment")
	g.AddEdge("process_payment", "finalize_order")
	g.AddEdge("finalize_order", graph.END)

	runnable, err := g.Compile()
	if err != nil {
		log.Fatal(err)
	}

	// Initial state
	initialState := OrderState{
		OrderID:       "ORD-2024-001",
		Amount:        99.99,
		CustomerID:    "CUST-123",
		PaymentMethod: "Credit Card",
	}

	fmt.Printf("\n🛒 Starting order: %s for customer %s\n",
		initialState.OrderID, initialState.CustomerID)
	fmt.Println(strings.Repeat("=", 60))

	// ===== STEP 1: Initial Run (will interrupt) =====
	fmt.Println("\n--- Step 1: Initial Execution ---")
	result, err := runnable.Invoke(context.Background(), initialState)

	var graphInterrupt *graph.GraphInterrupt
	if errors.As(err, &graphInterrupt) {
		fmt.Println("\n⚠️  Graph Interrupted!")
		fmt.Printf("   Node: %s\n", graphInterrupt.Node)
		fmt.Printf("   Question: %s\n", graphInterrupt.InterruptValue)

		// IMPORTANT: Check that state was preserved
		interruptState, ok := graphInterrupt.State.(OrderState)
		if !ok {
			log.Fatalf("Expected OrderState, got %T", graphInterrupt.State)
		}

		fmt.Println("\n📊 State at Interruption:")
		fmt.Printf("   Order ID: %s\n", interruptState.OrderID)
		fmt.Printf("   Payment Status: %s\n", interruptState.PaymentStatus)
		fmt.Printf("   Transaction ID: %s\n", interruptState.TransactionID)
		fmt.Printf("   Amount: $%.2f\n", interruptState.Amount)

		// Verify the fix worked
		if interruptState.PaymentStatus != "pending_payment" {
			fmt.Println("\n❌ BUG: State was not preserved! Expected 'pending_payment', got:", interruptState.PaymentStatus)
			fmt.Println("   This was the issue reported in #67")
		} else {
			fmt.Println("\n✅ SUCCESS: State modifications before Interrupt() were preserved!")
			fmt.Println("   This confirms Issue #67 is fixed.")
		}

		// Simulate user input
		fmt.Println("\n--- Step 2: User Confirmation ---")
		fmt.Println("💬 Simulating user confirming payment...")
		userConfirmed := true

		// ===== STEP 2: Resume Execution =====
		fmt.Println("\n--- Step 3: Resuming Execution ---")
		config := &graph.Config{
			ResumeValue: userConfirmed,
		}

		// Resume with the interrupted state
		result, err = runnable.InvokeWithConfig(context.Background(), interruptState, config)
		if err != nil {
			log.Fatalf("Resume execution failed: %v", err)
		}

		fmt.Println("\n📊 Final State:")
		fmt.Printf("   Order ID: %s\n", result.OrderID)
		fmt.Printf("   Payment Status: %s\n", result.PaymentStatus)
		fmt.Printf("   Transaction ID: %s\n", result.TransactionID)
		fmt.Printf("   Amount: $%.2f\n", result.Amount)

		if result.PaymentStatus == "paid" {
			fmt.Println("\n🎉 Order completed successfully!")
		}

	} else if err != nil {
		log.Fatalf("Execution failed: %v", err)
	} else {
		fmt.Println("Execution finished without interrupt (unexpected)")
		fmt.Printf("Final state: %+v\n", result)
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("Demo completed!")
}
