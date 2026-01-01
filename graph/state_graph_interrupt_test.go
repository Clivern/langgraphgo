package graph

import (
	"context"
	"errors"
	"testing"
)

func TestStateGraph_Interrupt(t *testing.T) {
	// Create a StateGraph
	g := NewStateGraph[map[string]any]()

	// Add node that uses Interrupt
	g.AddNode("node1", "Node with interrupt", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		// Use the Interrupt function
		resumeValue, err := Interrupt(ctx, "waiting for input")
		if err != nil {
			return nil, err
		}
		// If we resumed, return the resume value
		if resumeValue != nil {
			return map[string]any{"value": resumeValue}, nil
		}
		return map[string]any{"value": "default"}, nil
	})

	g.AddEdge("node1", END)
	g.SetEntryPoint("node1")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// First execution should interrupt
	_, err = runnable.Invoke(context.Background(), map[string]any{"initial": true})

	// Verify we got an interrupt error
	var graphInterrupt *GraphInterrupt
	if err == nil {
		t.Fatal("Expected interrupt error, got nil")
	}

	// Check if it's a NodeInterrupt wrapped in error or GraphInterrupt
	var nodeInterrupt *NodeInterrupt
	if !errors.As(err, &nodeInterrupt) {
		// Try GraphInterrupt
		if !errors.As(err, &graphInterrupt) {
			t.Fatalf("Expected NodeInterrupt or GraphInterrupt error, got: %v", err)
		}
	}

	if graphInterrupt != nil {
		if graphInterrupt.InterruptValue != "waiting for input" {
			t.Errorf("Expected interrupt value 'waiting for input', got: %v", graphInterrupt.InterruptValue)
		}
		t.Logf("Successfully interrupted with GraphInterrupt, value: %v", graphInterrupt.InterruptValue)
	} else {
		if nodeInterrupt.Value != "waiting for input" {
			t.Errorf("Expected interrupt value 'waiting for input', got: %v", nodeInterrupt.Value)
		}
		t.Logf("Successfully interrupted with NodeInterrupt, value: %v", nodeInterrupt.Value)
	}

	t.Log("StateGraph Interrupt test passed!")
}

func TestStateGraph_InterruptWithStateUpdate(t *testing.T) {
	// This test verifies that state modifications made before calling Interrupt
	// are preserved in the GraphInterrupt.State
	g := NewStateGraph[map[string]any]()

	g.AddNode("payment_node", "Payment processing node", func(ctx context.Context, state map[string]any) (map[string]any, error) {
		// Simulate updating state before interrupting
		// e.g., setting status to "pending_payment"
		state["payment_status"] = "pending_payment"
		state["amount"] = 100

		// Then interrupt to ask for user confirmation
		_, err := Interrupt(ctx, "Please confirm payment of $100")
		if err != nil {
			// When interrupting, return the updated state
			return state, err
		}

		// If resumed, mark as paid
		state["payment_status"] = "paid"
		return state, nil
	})

	g.AddEdge("payment_node", END)
	g.SetEntryPoint("payment_node")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// First execution should interrupt
	initialState := map[string]any{"user_id": "123"}
	result, err := runnable.Invoke(context.Background(), initialState)

	// Verify we got an interrupt error
	var graphInterrupt *GraphInterrupt
	if err == nil {
		t.Fatal("Expected interrupt error, got nil")
	}

	if !errors.As(err, &graphInterrupt) {
		t.Fatalf("Expected GraphInterrupt error, got: %v", err)
	}

	// Verify the interrupt value
	if graphInterrupt.InterruptValue != "Please confirm payment of $100" {
		t.Errorf("Expected interrupt value 'Please confirm payment of $100', got: %v", graphInterrupt.InterruptValue)
	}

	// CRITICAL: Verify that state modifications are preserved
	interruptState, ok := graphInterrupt.State.(map[string]any)
	if !ok {
		t.Fatalf("Expected state to be map[string]any, got: %T", graphInterrupt.State)
	}

	// Check that the state updates made before Interrupt() are present
	if interruptState["payment_status"] != "pending_payment" {
		t.Errorf("Expected payment_status to be 'pending_payment', got: %v", interruptState["payment_status"])
	}

	if interruptState["amount"] != 100 {
		t.Errorf("Expected amount to be 100, got: %v", interruptState["amount"])
	}

	// Also check that result has the updated state
	if result["payment_status"] != "pending_payment" {
		t.Errorf("Result: Expected payment_status to be 'pending_payment', got: %v", result["payment_status"])
	}

	if result["amount"] != 100 {
		t.Errorf("Result: Expected amount to be 100, got: %v", result["amount"])
	}

	t.Log("StateGraph Interrupt with state update test passed!")
}

// PaymentState is a value type for testing
type PaymentState struct {
	UserID        string
	PaymentStatus string
	Amount        int
}

func TestStateGraph_InterruptWithValueTypeState(t *testing.T) {
	// This test uses a value type (struct) instead of map (reference type)
	// to properly test if state updates are preserved during interrupt
	g := NewStateGraph[PaymentState]()

	g.AddNode("payment_node", "Payment processing node", func(ctx context.Context, state PaymentState) (PaymentState, error) {
		// Modify the state (creates a new copy since it's a value type)
		state.PaymentStatus = "pending_payment"
		state.Amount = 100

		// Then interrupt
		_, err := Interrupt(ctx, "Please confirm payment of $100")
		if err != nil {
			// Return the updated state
			return state, err
		}

		// If resumed, mark as paid
		state.PaymentStatus = "paid"
		return state, nil
	})

	g.AddEdge("payment_node", END)
	g.SetEntryPoint("payment_node")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	// First execution should interrupt
	initialState := PaymentState{UserID: "user123"}
	result, err := runnable.Invoke(context.Background(), initialState)

	// Verify we got an interrupt error
	var graphInterrupt *GraphInterrupt
	if err == nil {
		t.Fatal("Expected interrupt error, got nil")
	}

	if !errors.As(err, &graphInterrupt) {
		t.Fatalf("Expected GraphInterrupt error, got: %v", err)
	}

	t.Logf("GraphInterrupt.State: %+v", graphInterrupt.State)
	t.Logf("Result: %+v", result)

	// CRITICAL: Verify that state modifications are preserved
	interruptState, ok := graphInterrupt.State.(PaymentState)
	if !ok {
		t.Fatalf("Expected state to be PaymentState, got: %T", graphInterrupt.State)
	}

	// This should FAIL with the current bug - state updates are lost
	if interruptState.PaymentStatus != "pending_payment" {
		t.Errorf("BUG CONFIRMED: Expected payment_status to be 'pending_payment', got: %v", interruptState.PaymentStatus)
	}

	if interruptState.Amount != 100 {
		t.Errorf("BUG CONFIRMED: Expected amount to be 100, got: %v", interruptState.Amount)
	}

	// Also check result
	if result.PaymentStatus != "pending_payment" {
		t.Errorf("Result BUG: Expected payment_status to be 'pending_payment', got: %v", result.PaymentStatus)
	}

	if result.Amount != 100 {
		t.Errorf("Result BUG: Expected amount to be 100, got: %v", result.Amount)
	}
}
