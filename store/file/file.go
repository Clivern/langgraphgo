package file

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/smallnest/langgraphgo/store"
)

// FileCheckpointStore provides file-based checkpoint storage
type FileCheckpointStore struct {
	path  string
	mutex sync.RWMutex
}

// threadIndex represents the in-memory index for thread_id -> checkpoint IDs
type threadIndex struct {
	Threads map[string][]string // thread_id -> []checkpoint IDs
}

// NewFileCheckpointStore creates a new file-based checkpoint store
func NewFileCheckpointStore(path string) (store.CheckpointStore, error) {
	// Ensure directory exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	// Ensure index directory exists
	indexDir := filepath.Join(path, "by_thread")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %w", err)
	}

	return &FileCheckpointStore{
		path: path,
	}, nil
}

// Save implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) Save(_ context.Context, checkpoint *store.Checkpoint) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Create filename from ID
	filename := filepath.Join(f.path, fmt.Sprintf("%s.json", checkpoint.ID))

	data, err := json.Marshal(checkpoint)
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write checkpoint file: %w", err)
	}

	// Update thread_id index
	if threadID, ok := checkpoint.Metadata["thread_id"].(string); ok && threadID != "" {
		if err := f.addToThreadIndex(threadID, checkpoint.ID); err != nil {
			// Log error but don't fail the save
			_ = fmt.Errorf("failed to update thread index: %w", err)
		}
	}

	return nil
}

// Load implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) Load(_ context.Context, checkpointID string) (*store.Checkpoint, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	filename := filepath.Join(f.path, fmt.Sprintf("%s.json", checkpointID))

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("checkpoint not found: %s", checkpointID)
		}
		return nil, fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint store.Checkpoint
	err = json.Unmarshal(data, &checkpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	return &checkpoint, nil
}

// List implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) List(_ context.Context, executionID string) ([]*store.Checkpoint, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	files, err := os.ReadDir(f.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint directory: %w", err)
	}

	var checkpoints []*store.Checkpoint

	for _, file := range files {
		// Skip directories and non-JSON files
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(f.path, file.Name()))
		if err != nil {
			// Skip unreadable files
			continue
		}

		var checkpoint store.Checkpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			// Skip invalid files
			continue
		}

		// Filter by executionID, threadID, sessionID, or workflowID
		execID, _ := checkpoint.Metadata["execution_id"].(string)
		threadID, _ := checkpoint.Metadata["thread_id"].(string)
		sessionID, _ := checkpoint.Metadata["session_id"].(string)
		workflowID, _ := checkpoint.Metadata["workflow_id"].(string)

		if execID == executionID || threadID == executionID || sessionID == executionID || workflowID == executionID {
			checkpoints = append(checkpoints, &checkpoint)
		}
	}

	// Sort by version (ascending order) so latest is last
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Version < checkpoints[j].Version
	})

	return checkpoints, nil
}

// ListByThread returns all checkpoints for a specific thread_id using index
func (f *FileCheckpointStore) ListByThread(_ context.Context, threadID string) ([]*store.Checkpoint, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// Load thread index
	checkpointIDs, err := f.loadThreadIndex(threadID)
	if err != nil {
		// Fallback to scanning all files if index doesn't exist
		return f.listByThreadScan(threadID)
	}

	if len(checkpointIDs) == 0 {
		return []*store.Checkpoint{}, nil
	}

	var checkpoints []*store.Checkpoint
	for _, id := range checkpointIDs {
		filename := filepath.Join(f.path, fmt.Sprintf("%s.json", id))
		data, err := os.ReadFile(filename)
		if err != nil {
			// Skip unreadable files
			continue
		}

		var checkpoint store.Checkpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			// Skip invalid files
			continue
		}

		checkpoints = append(checkpoints, &checkpoint)
	}

	// Sort by version (ascending order)
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Version < checkpoints[j].Version
	})

	return checkpoints, nil
}

// GetLatestByThread returns the latest checkpoint for a thread_id
func (f *FileCheckpointStore) GetLatestByThread(ctx context.Context, threadID string) (*store.Checkpoint, error) {
	checkpoints, err := f.ListByThread(ctx, threadID)
	if err != nil {
		return nil, err
	}

	if len(checkpoints) == 0 {
		return nil, fmt.Errorf("no checkpoints found for thread: %s", threadID)
	}

	// Return the last one (highest version due to sorting)
	return checkpoints[len(checkpoints)-1], nil
}

// Delete implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) Delete(_ context.Context, checkpointID string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// Load checkpoint first to get thread_id
	filename := filepath.Join(f.path, fmt.Sprintf("%s.json", checkpointID))
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Already deleted
			return nil
		}
		return fmt.Errorf("failed to read checkpoint file: %w", err)
	}

	var checkpoint store.Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return fmt.Errorf("failed to unmarshal checkpoint: %w", err)
	}

	// Remove the checkpoint file
	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete checkpoint file: %w", err)
	}

	// Remove from thread index
	if threadID, ok := checkpoint.Metadata["thread_id"].(string); ok && threadID != "" {
		if err := f.removeFromThreadIndex(threadID, checkpointID); err != nil {
			// Log error but don't fail the delete
			_ = fmt.Errorf("failed to update thread index: %w", err)
		}
	}

	return nil
}

// Clear implements CheckpointStore interface for file storage
func (f *FileCheckpointStore) Clear(ctx context.Context, executionID string) error {
	checkpoints, err := f.List(ctx, executionID)
	if err != nil {
		return err
	}

	var errs []error
	for _, cp := range checkpoints {
		if err := f.Delete(ctx, cp.ID); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("failed to clear some checkpoints: %v", errs)
	}

	return nil
}

// Helper functions for thread index management

func (f *FileCheckpointStore) getThreadIndexPath(threadID string) string {
	return filepath.Join(f.path, "by_thread", fmt.Sprintf("%s.json", threadID))
}

func (f *FileCheckpointStore) loadThreadIndex(threadID string) ([]string, error) {
	indexPath := f.getThreadIndexPath(threadID)

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var index threadIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	if index.Threads == nil {
		return []string{}, nil
	}

	ids, ok := index.Threads[threadID]
	if !ok {
		return []string{}, nil
	}

	return ids, nil
}

func (f *FileCheckpointStore) addToThreadIndex(threadID, checkpointID string) error {
	indexPath := f.getThreadIndexPath(threadID)

	// Load existing index
	var index threadIndex
	if data, err := os.ReadFile(indexPath); err == nil {
		_ = json.Unmarshal(data, &index)
	}

	if index.Threads == nil {
		index.Threads = make(map[string][]string)
	}

	// Add checkpoint ID to index
	index.Threads[threadID] = append(index.Threads[threadID], checkpointID)

	// Write index back to disk
	data, err := json.Marshal(index)
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0600)
}

func (f *FileCheckpointStore) removeFromThreadIndex(threadID, checkpointID string) error {
	indexPath := f.getThreadIndexPath(threadID)

	// Load existing index
	var index threadIndex
	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if err := json.Unmarshal(data, &index); err != nil {
		return err
	}

	if index.Threads == nil {
		return nil
	}

	// Remove checkpoint ID from index
	ids, ok := index.Threads[threadID]
	if !ok {
		return nil
	}

	for i, id := range ids {
		if id == checkpointID {
			index.Threads[threadID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}

	// Write index back to disk
	data, err = json.Marshal(index)
	if err != nil {
		return err
	}

	return os.WriteFile(indexPath, data, 0600)
}

// listByThreadScan is a fallback method that scans all files
func (f *FileCheckpointStore) listByThreadScan(threadID string) ([]*store.Checkpoint, error) {
	files, err := os.ReadDir(f.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read checkpoint directory: %w", err)
	}

	var checkpoints []*store.Checkpoint

	for _, file := range files {
		// Skip directories and non-JSON files
		if file.IsDir() || filepath.Ext(file.Name()) != ".json" {
			continue
		}

		// Skip index directory
		if file.Name() == "by_thread" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(f.path, file.Name()))
		if err != nil {
			continue
		}

		var checkpoint store.Checkpoint
		if err := json.Unmarshal(data, &checkpoint); err != nil {
			continue
		}

		// Filter by thread_id
		if cpThreadID, ok := checkpoint.Metadata["thread_id"].(string); ok && cpThreadID == threadID {
			checkpoints = append(checkpoints, &checkpoint)
		}
	}

	// Sort by version (ascending order)
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Version < checkpoints[j].Version
	})

	return checkpoints, nil
}
