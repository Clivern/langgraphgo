package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/smallnest/langgraphgo/store"
)

// MemoryCheckpointStore provides in-memory checkpoint storage
type MemoryCheckpointStore struct {
	checkpoints    map[string]*store.Checkpoint // id -> checkpoint
	threadIndex    map[string][]string          // thread_id -> []checkpoint IDs
	executionIndex map[string][]string          // execution_id -> []checkpoint IDs
	mutex          sync.RWMutex
}

// NewMemoryCheckpointStore creates a new in-memory checkpoint store
func NewMemoryCheckpointStore() store.CheckpointStore {
	return &MemoryCheckpointStore{
		checkpoints:    make(map[string]*store.Checkpoint),
		threadIndex:    make(map[string][]string),
		executionIndex: make(map[string][]string),
	}
}

// Save implements CheckpointStore interface
func (m *MemoryCheckpointStore) Save(_ context.Context, checkpoint *store.Checkpoint) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Store checkpoint
	m.checkpoints[checkpoint.ID] = checkpoint

	// Update execution_id index
	if execID, ok := checkpoint.Metadata["execution_id"].(string); ok && execID != "" {
		m.executionIndex[execID] = append(m.executionIndex[execID], checkpoint.ID)
	}

	// Update thread_id index
	if threadID, ok := checkpoint.Metadata["thread_id"].(string); ok && threadID != "" {
		m.threadIndex[threadID] = append(m.threadIndex[threadID], checkpoint.ID)
	}

	return nil
}

// Load implements CheckpointStore interface
func (m *MemoryCheckpointStore) Load(_ context.Context, checkpointID string) (*store.Checkpoint, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	checkpoint, exists := m.checkpoints[checkpointID]
	if !exists {
		return nil, fmt.Errorf("checkpoint not found: %s", checkpointID)
	}

	return checkpoint, nil
}

// List implements CheckpointStore interface
func (m *MemoryCheckpointStore) List(_ context.Context, executionID string) ([]*store.Checkpoint, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var checkpoints []*store.Checkpoint
	for _, checkpoint := range m.checkpoints {
		// Filter by various ID fields that can be used for grouping
		execID, _ := checkpoint.Metadata["execution_id"].(string)
		threadID, _ := checkpoint.Metadata["thread_id"].(string)
		sessionID, _ := checkpoint.Metadata["session_id"].(string)
		workflowID, _ := checkpoint.Metadata["workflow_id"].(string)

		if execID == executionID || threadID == executionID || sessionID == executionID || workflowID == executionID {
			checkpoints = append(checkpoints, checkpoint)
		}
	}

	// Sort by version (ascending order) so latest is last
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Version < checkpoints[j].Version
	})

	return checkpoints, nil
}

// ListByThread returns all checkpoints for a specific thread_id
func (m *MemoryCheckpointStore) ListByThread(_ context.Context, threadID string) ([]*store.Checkpoint, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ids, exists := m.threadIndex[threadID]
	if !exists {
		return []*store.Checkpoint{}, nil
	}

	checkpoints := make([]*store.Checkpoint, 0, len(ids))
	for _, id := range ids {
		if cp, ok := m.checkpoints[id]; ok {
			checkpoints = append(checkpoints, cp)
		}
	}

	// Sort by version (ascending order)
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Version < checkpoints[j].Version
	})

	return checkpoints, nil
}

// GetLatestByThread returns the latest checkpoint for a thread_id
func (m *MemoryCheckpointStore) GetLatestByThread(_ context.Context, threadID string) (*store.Checkpoint, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	ids, exists := m.threadIndex[threadID]
	if !exists || len(ids) == 0 {
		return nil, fmt.Errorf("no checkpoints found for thread: %s", threadID)
	}

	var latest *store.Checkpoint
	for _, id := range ids {
		cp := m.checkpoints[id]
		if latest == nil || cp.Version > latest.Version {
			latest = cp
		}
	}

	return latest, nil
}

// Delete implements CheckpointStore interface
func (m *MemoryCheckpointStore) Delete(_ context.Context, checkpointID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	checkpoint, exists := m.checkpoints[checkpointID]
	if !exists {
		return nil
	}

	// Remove from indexes
	if execID, ok := checkpoint.Metadata["execution_id"].(string); ok {
		if ids, ok := m.executionIndex[execID]; ok {
			for i, id := range ids {
				if id == checkpointID {
					m.executionIndex[execID] = append(ids[:i], ids[i+1:]...)
					break
				}
			}
		}
	}

	if threadID, ok := checkpoint.Metadata["thread_id"].(string); ok {
		if ids, ok := m.threadIndex[threadID]; ok {
			for i, id := range ids {
				if id == checkpointID {
					m.threadIndex[threadID] = append(ids[:i], ids[i+1:]...)
					break
				}
			}
		}
	}

	delete(m.checkpoints, checkpointID)
	return nil
}

// Clear implements CheckpointStore interface
func (m *MemoryCheckpointStore) Clear(_ context.Context, executionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var idsToDelete []string

	// Find checkpoints to delete
	for id, checkpoint := range m.checkpoints {
		execID, _ := checkpoint.Metadata["execution_id"].(string)
		threadID, _ := checkpoint.Metadata["thread_id"].(string)
		sessionID, _ := checkpoint.Metadata["session_id"].(string)
		workflowID, _ := checkpoint.Metadata["workflow_id"].(string)

		if execID == executionID || threadID == executionID || sessionID == executionID || workflowID == executionID {
			idsToDelete = append(idsToDelete, id)
		}
	}

	// Delete from indexes and main map
	for _, id := range idsToDelete {
		checkpoint := m.checkpoints[id]

		// Remove from execution_index
		if execID, ok := checkpoint.Metadata["execution_id"].(string); ok {
			if ids, ok := m.executionIndex[execID]; ok {
				for i, cid := range ids {
					if cid == id {
						m.executionIndex[execID] = append(ids[:i], ids[i+1:]...)
						break
					}
				}
			}
		}

		// Remove from thread_index
		if threadID, ok := checkpoint.Metadata["thread_id"].(string); ok {
			if ids, ok := m.threadIndex[threadID]; ok {
				for i, cid := range ids {
					if cid == id {
						m.threadIndex[threadID] = append(ids[:i], ids[i+1:]...)
						break
					}
				}
			}
		}

		delete(m.checkpoints, id)
	}

	return nil
}
