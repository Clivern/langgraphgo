package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/store"
)

const (
	testNode   = "test_node"
	testResult = "test_result"
)

func TestNewMemoryCheckpointStore(t *testing.T) {
	t.Parallel()

	ms := NewMemoryCheckpointStore()

	if ms == nil {
		t.Fatal("Expected store but got nil")
	}

	// Verify it implements the interface
	var _ store.CheckpointStore = ms
}

func TestMemoryCheckpointStore_SaveAndLoad(t *testing.T) {
	t.Parallel()

	ms := NewMemoryCheckpointStore()
	ctx := context.Background()

	checkpoint := &store.Checkpoint{
		ID:        "test_checkpoint_1",
		NodeName:  testNode,
		State:     "test_state",
		Timestamp: time.Now(),
		Version:   1,
		Metadata: map[string]interface{}{
			"execution_id": "exec_123",
		},
	}

	// Save checkpoint
	err := ms.Save(ctx, checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Load checkpoint
	loaded, err := ms.Load(ctx, "test_checkpoint_1")
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	if loaded.ID != checkpoint.ID {
		t.Errorf("Expected ID %s, got %s", checkpoint.ID, loaded.ID)
	}

	if loaded.NodeName != checkpoint.NodeName {
		t.Errorf("Expected NodeName %s, got %s", checkpoint.NodeName, loaded.NodeName)
	}

	if loaded.State != checkpoint.State {
		t.Errorf("Expected State %v, got %v", checkpoint.State, loaded.State)
	}

	if loaded.Version != checkpoint.Version {
		t.Errorf("Expected Version %d, got %d", checkpoint.Version, loaded.Version)
	}

	// Check metadata
	execID, ok := loaded.Metadata["execution_id"].(string)
	if !ok {
		t.Error("Expected execution_id to be a string")
	} else if execID != "exec_123" {
		t.Errorf("Expected execution_id exec_123, got %s", execID)
	}
}

func TestMemoryCheckpointStore_LoadNonExistent(t *testing.T) {
	t.Parallel()

	ms := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Try to load non-existing checkpoint
	_, err := ms.Load(ctx, "non_existing")
	if err == nil {
		t.Error("Expected error for non-existing checkpoint")
	}

	expectedError := "checkpoint not found: non_existing"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestMemoryCheckpointStore_SaveOverwrite(t *testing.T) {
	t.Parallel()

	ms := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Save initial checkpoint
	checkpoint1 := &store.Checkpoint{
		ID:        "test_checkpoint",
		NodeName:  "node1",
		State:     "state1",
		Timestamp: time.Now(),
		Version:   1,
	}

	err := ms.Save(ctx, checkpoint1)
	if err != nil {
		t.Fatalf("Failed to save initial checkpoint: %v", err)
	}

	// Save checkpoint with same ID (overwrite)
	checkpoint2 := &store.Checkpoint{
		ID:        "test_checkpoint",
		NodeName:  "node2",
		State:     "state2",
		Timestamp: time.Now(),
		Version:   2,
	}

	err = ms.Save(ctx, checkpoint2)
	if err != nil {
		t.Fatalf("Failed to overwrite checkpoint: %v", err)
	}

	// Load and verify it's the second checkpoint
	loaded, err := ms.Load(ctx, "test_checkpoint")
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}

	if loaded.NodeName != "node2" {
		t.Errorf("Expected NodeName 'node2', got '%s'", loaded.NodeName)
	}

	if loaded.State != "state2" {
		t.Errorf("Expected State 'state2', got '%v'", loaded.State)
	}

	if loaded.Version != 2 {
		t.Errorf("Expected Version 2, got %d", loaded.Version)
	}
}

func TestMemoryCheckpointStore_List(t *testing.T) {
	t.Parallel()

	ms := NewMemoryCheckpointStore()
	ctx := context.Background()
	executionID := "exec_123"
	threadID := "thread_456"

	// Save multiple checkpoints
	checkpoints := []*store.Checkpoint{
		{
			ID:       "checkpoint_1",
			NodeName: "node1",
			Metadata: map[string]interface{}{
				"execution_id": executionID,
			},
			Version:   1,
			Timestamp: time.Now(),
		},
		{
			ID:       "checkpoint_2",
			NodeName: "node2",
			Metadata: map[string]interface{}{
				"execution_id": executionID,
			},
			Version:   2,
			Timestamp: time.Now().Add(time.Hour),
		},
		{
			ID:       "checkpoint_3",
			NodeName: "node3",
			Metadata: map[string]interface{}{
				"thread_id": threadID,
			},
			Version:   1,
			Timestamp: time.Now().Add(2 * time.Hour),
		},
		{
			ID:       "checkpoint_4",
			NodeName: "node4",
			Metadata: map[string]interface{}{
				"execution_id": "different_exec",
			},
			Version:   1,
			Timestamp: time.Now().Add(3 * time.Hour),
		},
		{
			ID:       "checkpoint_5",
			NodeName: "node5",
			// No metadata - should not be included in any list by execution_id or thread_id
			Version:   1,
			Timestamp: time.Now().Add(4 * time.Hour),
		},
	}

	for _, checkpoint := range checkpoints {
		err := ms.Save(ctx, checkpoint)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}
	}

	tests := []struct {
		name         string
		executionID  string
		expectedLen  int
		expectedIDs  []string
		expectSorted bool
	}{
		{
			name:        "list by execution_id",
			executionID: executionID,
			expectedLen: 2,
			expectedIDs: []string{"checkpoint_1", "checkpoint_2"},
			expectSorted: true, // Should be sorted by version ascending
		},
		{
			name:        "list by thread_id",
			executionID: threadID,
			expectedLen: 1,
			expectedIDs: []string{"checkpoint_3"},
			expectSorted: true,
		},
		{
			name:        "list non-existing execution",
			executionID: "non_existing",
			expectedLen: 0,
			expectedIDs: []string{},
			expectSorted: true,
		},
		{
			name:        "list different execution",
			executionID: "different_exec",
			expectedLen: 1,
			expectedIDs: []string{"checkpoint_4"},
			expectSorted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			listed, err := ms.List(ctx, tt.executionID)

			if err != nil {
				t.Fatalf("Failed to list checkpoints: %v", err)
			}

			if len(listed) != tt.expectedLen {
				t.Errorf("Expected %d checkpoints, got %d", tt.expectedLen, len(listed))
			}

			// Verify correct checkpoints returned
			ids := make(map[string]bool)
			for _, checkpoint := range listed {
				ids[checkpoint.ID] = true
			}

			for _, expectedID := range tt.expectedIDs {
				if !ids[expectedID] {
					t.Errorf("Expected checkpoint ID %s not found in results", expectedID)
				}
			}

			// Verify sorting order if sorting is expected and there are multiple items
			if tt.expectSorted && len(listed) > 1 {
				for i := 1; i < len(listed); i++ {
					if listed[i-1].Version > listed[i].Version {
						t.Error("Checkpoints should be sorted by version ascending")
						break
					}
				}
			}
		})
	}
}

func TestMemoryCheckpointStore_Delete(t *testing.T) {
	t.Parallel()

	ms := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Save multiple checkpoints
	checkpoints := []*store.Checkpoint{
		{ID: "checkpoint_1", Version: 1},
		{ID: "checkpoint_2", Version: 1},
		{ID: "checkpoint_3", Version: 1},
	}

	for _, checkpoint := range checkpoints {
		err := ms.Save(ctx, checkpoint)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}
	}

	// Delete one checkpoint
	err := ms.Delete(ctx, "checkpoint_2")
	if err != nil {
		t.Fatalf("Failed to delete checkpoint: %v", err)
	}

	// Verify checkpoint is deleted
	_, err = ms.Load(ctx, "checkpoint_2")
	if err == nil {
		t.Error("Expected checkpoint_2 to be deleted")
	}

	// Verify other checkpoints still exist
	_, err = ms.Load(ctx, "checkpoint_1")
	if err != nil {
		t.Errorf("Expected checkpoint_1 to still exist: %v", err)
	}

	_, err = ms.Load(ctx, "checkpoint_3")
	if err != nil {
		t.Errorf("Expected checkpoint_3 to still exist: %v", err)
	}

	// Delete non-existing checkpoint (should not error)
	err = ms.Delete(ctx, "non_existing")
	if err != nil {
		t.Errorf("Expected no error when deleting non-existing checkpoint: %v", err)
	}
}

func TestMemoryCheckpointStore_Clear(t *testing.T) {
	t.Parallel()

	ms := NewMemoryCheckpointStore()
	ctx := context.Background()
	executionID := "exec_123"
	differentExecutionID := "exec_456"

	// Save multiple checkpoints for different executions
	checkpoints := []*store.Checkpoint{
		{
			ID: "checkpoint_1",
			Metadata: map[string]interface{}{
				"execution_id": executionID,
			},
			Version: 1,
		},
		{
			ID: "checkpoint_2",
			Metadata: map[string]interface{}{
				"execution_id": executionID,
			},
			Version: 2,
		},
		{
			ID: "checkpoint_3",
			Metadata: map[string]interface{}{
				"execution_id": differentExecutionID,
			},
			Version: 1,
		},
	}

	for _, checkpoint := range checkpoints {
		err := ms.Save(ctx, checkpoint)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}
	}

	// Verify checkpoints exist before clear
	listed, err := ms.List(ctx, executionID)
	if err != nil {
		t.Fatalf("Failed to list checkpoints before clear: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("Expected 2 checkpoints before clear, got %d", len(listed))
	}

	// Clear checkpoints for specific execution
	err = ms.Clear(ctx, executionID)
	if err != nil {
		t.Fatalf("Failed to clear checkpoints: %v", err)
	}

	// Verify checkpoints for executionID are cleared
	listed, err = ms.List(ctx, executionID)
	if err != nil {
		t.Fatalf("Failed to list checkpoints after clear: %v", err)
	}
	if len(listed) != 0 {
		t.Errorf("Expected 0 checkpoints after clear, got %d", len(listed))
	}

	// Verify checkpoints for different executionID still exist
	listed, err = ms.List(ctx, differentExecutionID)
	if err != nil {
		t.Fatalf("Failed to list checkpoints for different execution: %v", err)
	}
	if len(listed) != 1 {
		t.Errorf("Expected 1 checkpoint for different execution, got %d", len(listed))
	}

	// Verify individual checkpoints
	_, err = ms.Load(ctx, "checkpoint_1")
	if err == nil {
		t.Error("Expected checkpoint_1 to be cleared")
	}

	_, err = ms.Load(ctx, "checkpoint_2")
	if err == nil {
		t.Error("Expected checkpoint_2 to be cleared")
	}

	_, err = ms.Load(ctx, "checkpoint_3")
	if err != nil {
		t.Errorf("Expected checkpoint_3 to still exist: %v", err)
	}
}

func TestMemoryCheckpointStore_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	ms := NewMemoryCheckpointStore()
	ctx := context.Background()

	// Test concurrent saves
	done := make(chan bool, 10)
	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			checkpoint := &store.Checkpoint{
				ID:       fmt.Sprintf("checkpoint_%d", id),
				NodeName: fmt.Sprintf("node_%d", id),
				Metadata: map[string]interface{}{
					"execution_id": fmt.Sprintf("exec_%d", id),
				},
				Version:   1,
				Timestamp: time.Now(),
			}

			if err := ms.Save(ctx, checkpoint); err != nil {
				errs <- fmt.Errorf("failed to save checkpoint %d: %v", id, err)
				return
			}

			// Try to load it immediately
			if _, err := ms.Load(ctx, checkpoint.ID); err != nil {
				errs <- fmt.Errorf("failed to load checkpoint %d: %v", id, err)
				return
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// OK
		case err := <-errs:
			t.Errorf("Error in goroutine: %v", err)
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}

	// Verify all checkpoints were saved
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("checkpoint_%d", i)
		_, err := ms.Load(ctx, id)
		if err != nil {
			t.Errorf("Expected checkpoint %s to exist", id)
		}
	}
}