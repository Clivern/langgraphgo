package store

import (
	"context"
	"time"
)

// Checkpoint represents a saved state at a specific point in execution
type Checkpoint struct {
	ID        string         `json:"id"`
	NodeName  string         `json:"node_name"`
	State     any            `json:"state"`
	Metadata  map[string]any `json:"metadata"`
	Timestamp time.Time      `json:"timestamp"`
	Version   int            `json:"version"`
}

// CheckpointStore defines the interface for checkpoint persistence
type CheckpointStore interface {
	// Save stores a checkpoint
	Save(ctx context.Context, checkpoint *Checkpoint) error

	// Load retrieves a checkpoint by ID
	Load(ctx context.Context, checkpointID string) (*Checkpoint, error)

	// List returns all checkpoints for a given execution
	List(ctx context.Context, executionID string) ([]*Checkpoint, error)

	// ListByThread returns all checkpoints for a specific thread_id.
	// Returns checkpoints sorted by version (ascending).
	ListByThread(ctx context.Context, threadID string) ([]*Checkpoint, error)

	// GetLatestByThread returns the latest checkpoint for a thread_id.
	// Returns the checkpoint with the highest version.
	GetLatestByThread(ctx context.Context, threadID string) (*Checkpoint, error)

	// Delete removes a checkpoint
	Delete(ctx context.Context, checkpointID string) error

	// Clear removes all checkpoints for an execution
	Clear(ctx context.Context, executionID string) error
}
