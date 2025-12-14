package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/store"
)

const (
	testNode   = "test_node"
	testResult = "test_result"
)

func TestNewFileCheckpointStore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "valid path",
			path:        t.TempDir(),
			expectError: false,
		},
		{
			name:        "relative path",
			path:        "./test_checkpoints",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Clean up after test if using relative path
			if tt.path == "./test_checkpoints" {
				defer os.RemoveAll(tt.path)
			}

			fs, err := NewFileCheckpointStore(tt.path)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Expected no error but got: %v", err)
			}

			if fs == nil {
				t.Fatal("Expected store but got nil")
			}

			// Verify directory was created
			if _, err := os.Stat(tt.path); os.IsNotExist(err) {
				t.Errorf("Expected directory %s to be created", tt.path)
			}
		})
	}
}

func TestFileCheckpointStore_Save(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	fs, err := NewFileCheckpointStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file checkpoint store: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		name       string
		checkpoint *store.Checkpoint
		expectErr  bool
	}{
		{
			name: "valid checkpoint",
			checkpoint: &store.Checkpoint{
				ID:        "test_checkpoint_1",
				NodeName:  testNode,
				State:     "test_state",
				Timestamp: time.Now(),
				Version:   1,
				Metadata: map[string]interface{}{
					"execution_id": "exec_123",
				},
			},
			expectErr: false,
		},
		{
			name: "checkpoint with complex state",
			checkpoint: &store.Checkpoint{
				ID:        "test_checkpoint_2",
				NodeName:  "complex_node",
				State: map[string]interface{}{
					"key1": "value1",
					"key2": 42,
					"key3": []string{"a", "b", "c"},
				},
				Timestamp: time.Now(),
				Version:   2,
				Metadata: map[string]interface{}{
					"execution_id": "exec_456",
					"thread_id":    "thread_789",
				},
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := fs.Save(ctx, tt.checkpoint)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			// Verify file was created
			filename := filepath.Join(tempDir, tt.checkpoint.ID+".json")
			if _, err := os.Stat(filename); os.IsNotExist(err) {
				t.Errorf("Expected checkpoint file %s to be created", filename)
			}

			// Verify file content
			data, err := os.ReadFile(filename)
			if err != nil {
				t.Errorf("Failed to read checkpoint file: %v", err)
				return
			}

			var savedCheckpoint store.Checkpoint
			err = json.Unmarshal(data, &savedCheckpoint)
			if err != nil {
				t.Errorf("Failed to unmarshal checkpoint: %v", err)
				return
			}

			if savedCheckpoint.ID != tt.checkpoint.ID {
				t.Errorf("Expected ID %s, got %s", tt.checkpoint.ID, savedCheckpoint.ID)
			}

			if savedCheckpoint.NodeName != tt.checkpoint.NodeName {
				t.Errorf("Expected NodeName %s, got %s", tt.checkpoint.NodeName, savedCheckpoint.NodeName)
			}

			if savedCheckpoint.Version != tt.checkpoint.Version {
				t.Errorf("Expected Version %d, got %d", tt.checkpoint.Version, savedCheckpoint.Version)
			}
		})
	}
}

func TestFileCheckpointStore_Load(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	fs, err := NewFileCheckpointStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file checkpoint store: %v", err)
	}

	ctx := context.Background()

	// Create a checkpoint file manually for testing
	checkpoint := &store.Checkpoint{
		ID:        "test_checkpoint",
		NodeName:  testNode,
		State:     "test_state",
		Timestamp: time.Now().UTC(),
		Version:   1,
		Metadata: map[string]interface{}{
			"execution_id": "exec_123",
		},
	}

	// Save checkpoint first
	err = fs.Save(ctx, checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	tests := []struct {
		name         string
		checkpointID string
		expectErr    bool
		expectFound  bool
	}{
		{
			name:        "existing checkpoint",
			checkpointID: "test_checkpoint",
			expectErr:   false,
			expectFound: true,
		},
		{
			name:        "non-existing checkpoint",
			checkpointID: "non_existing",
			expectErr:   true,
			expectFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			loaded, err := fs.Load(ctx, tt.checkpointID)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if loaded != nil {
					t.Errorf("Expected nil checkpoint but got %v", loaded)
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			if loaded == nil && tt.expectFound {
				t.Error("Expected checkpoint but got nil")
				return
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
		})
	}
}

func TestFileCheckpointStore_List(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	fs, err := NewFileCheckpointStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file checkpoint store: %v", err)
	}

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
		err := fs.Save(ctx, checkpoint)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}
	}

	// Create a corrupted file that should be skipped
	corruptedFile := filepath.Join(tempDir, "corrupted.json")
	err = os.WriteFile(corruptedFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create corrupted file: %v", err)
	}

	// Create a non-JSON file that should be skipped
	nonJSONFile := filepath.Join(tempDir, "readme.txt")
	err = os.WriteFile(nonJSONFile, []byte("readme content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create non-JSON file: %v", err)
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

			listed, err := fs.List(ctx, tt.executionID)

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

func TestFileCheckpointStore_Delete(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	fs, err := NewFileCheckpointStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file checkpoint store: %v", err)
	}

	ctx := context.Background()

	// Create checkpoint
	checkpoint := &store.Checkpoint{
		ID:        "test_checkpoint",
		NodeName:  testNode,
		State:     "test_state",
		Timestamp: time.Now(),
		Version:   1,
	}

	// Save checkpoint
	err = fs.Save(ctx, checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Verify file exists
	filename := filepath.Join(tempDir, checkpoint.ID+".json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("Checkpoint file should exist before deletion")
	}

	tests := []struct {
		name        string
		checkpointID string
		expectErr   bool
	}{
		{
			name:        "delete existing checkpoint",
			checkpointID: "test_checkpoint",
			expectErr:   false,
		},
		{
			name:        "delete non-existing checkpoint",
			checkpointID: "non_existing",
			expectErr:   false, // Should not error, just silently succeed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := fs.Delete(ctx, tt.checkpointID)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
				return
			}

			// For the existing checkpoint, verify file is deleted
			if tt.checkpointID == "test_checkpoint" {
				if _, err := os.Stat(filename); !os.IsNotExist(err) {
					t.Errorf("Checkpoint file should be deleted")
				}
			}
		})
	}
}

func TestFileCheckpointStore_Clear(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	fs, err := NewFileCheckpointStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file checkpoint store: %v", err)
	}

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
		err := fs.Save(ctx, checkpoint)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}
	}

	// Verify checkpoints exist before clear
	listed, err := fs.List(ctx, executionID)
	if err != nil {
		t.Fatalf("Failed to list checkpoints before clear: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("Expected 2 checkpoints before clear, got %d", len(listed))
	}

	// Clear checkpoints for specific execution
	err = fs.Clear(ctx, executionID)
	if err != nil {
		t.Fatalf("Failed to clear checkpoints: %v", err)
	}

	// Verify checkpoints for executionID are cleared
	listed, err = fs.List(ctx, executionID)
	if err != nil {
		t.Fatalf("Failed to list checkpoints after clear: %v", err)
	}
	if len(listed) != 0 {
		t.Errorf("Expected 0 checkpoints after clear, got %d", len(listed))
	}

	// Verify checkpoints for different executionID still exist
	listed, err = fs.List(ctx, differentExecutionID)
	if err != nil {
		t.Fatalf("Failed to list checkpoints for different execution: %v", err)
	}
	if len(listed) != 1 {
		t.Errorf("Expected 1 checkpoint for different execution, got %d", len(listed))
	}
}

func TestFileCheckpointStore_FilePermissions(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	fs, err := NewFileCheckpointStore(tempDir)
	if err != nil {
		t.Fatalf("Failed to create file checkpoint store: %v", err)
	}

	ctx := context.Background()

	checkpoint := &store.Checkpoint{
		ID:        "permission_test",
		NodeName:  testNode,
		State:     "test_state",
		Timestamp: time.Now(),
		Version:   1,
	}

	// Save checkpoint
	err = fs.Save(ctx, checkpoint)
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}

	// Check file permissions
	filename := filepath.Join(tempDir, checkpoint.ID+".json")
	fileInfo, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to stat checkpoint file: %v", err)
	}

	// On Unix systems, expect 0600 permissions (user read/write only)
	// Note: This test might behave differently on Windows
	if fileInfo.Mode().Perm()&0o777 != 0o600 {
		// Only check on Unix-like systems
		if os.Getenv("GOOS") != "windows" {
			t.Errorf("Expected file permissions 0600, got %o", fileInfo.Mode().Perm()&0o777)
		}
	}
}