package store

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/smallnest/langgraphgo/rag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSQLiteVecVectorStore_InMemory tests the in-memory sqlite-vec vector store
func TestSQLiteVecVectorStore_InMemory(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewSQLiteVecVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Add and Search", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "1", Content: "hello", Embedding: []float32{1, 0, 0}},
			{ID: "2", Content: "world", Embedding: []float32{0, 1, 0}},
		}
		err := store.Add(ctx, docs)
		assert.NoError(t, err)

		// Search for something close to "hello"
		results, err := store.Search(ctx, []float32{1, 0.1, 0}, 1)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "1", results[0].Document.ID)
		assert.GreaterOrEqual(t, results[0].Score, 0.89) // Adjust for floating point precision
	})

	t.Run("Search with Filter", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "3", Content: "filtered", Embedding: []float32{0, 0, 1}, Metadata: map[string]any{"type": "special"}},
		}
		err := store.Add(ctx, docs)
		assert.NoError(t, err)

		results, err := store.SearchWithFilter(ctx, []float32{0, 0, 1}, 10, map[string]any{"type": "special"})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "3", results[0].Document.ID)

		results, err = store.SearchWithFilter(ctx, []float32{0, 0, 1}, 10, map[string]any{"type": "none"})
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("Update and Delete", func(t *testing.T) {
		// First delete the existing document with ID "1"
		err := store.Delete(ctx, []string{"1"})
		assert.NoError(t, err)

		// Now add the updated document
		doc := rag.Document{ID: "1", Content: "updated", Embedding: []float32{1, 1, 1}}
		err = store.Update(ctx, []rag.Document{doc})
		assert.NoError(t, err)

		stats, _ := store.GetStats(ctx)
		countBefore := stats.TotalDocuments

		err = store.Delete(ctx, []string{"1"})
		assert.NoError(t, err)

		stats, _ = store.GetStats(ctx)
		assert.Equal(t, countBefore-1, stats.TotalDocuments)
	})

	t.Run("GetStats", func(t *testing.T) {
		stats, err := store.GetStats(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalDocuments, 0)
		assert.Equal(t, 3, stats.Dimension)
	})

	t.Run("Add without embedding", func(t *testing.T) {
		doc := rag.Document{ID: "4", Content: "no emb"}
		err := store.Add(ctx, []rag.Document{doc})
		assert.NoError(t, err)

		stats, _ := store.GetStats(ctx)
		assert.GreaterOrEqual(t, stats.TotalVectors, 1)
	})
}

// TestSQLiteVecVectorStore_Persistent tests the persistent sqlite-vec vector store
func TestSQLiteVecVectorStore_Persistent(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	// Create a temporary directory for the test
	tempDir := filepath.Join(os.TempDir(), "sqlitevec-test")
	dbPath := filepath.Join(tempDir, "test.db")
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	// Create the directory first
	err := os.MkdirAll(tempDir, 0o755)
	require.NoError(t, err)

	store1, err := NewSQLiteVecVectorStoreSimple(dbPath, embedder)
	require.NoError(t, err)

	t.Run("Add documents and verify persistence", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "1", Content: "persistent doc 1", Embedding: []float32{1, 0, 0}},
			{ID: "2", Content: "persistent doc 2", Embedding: []float32{0, 1, 0}},
		}
		err := store1.Add(ctx, docs)
		assert.NoError(t, err)

		// Verify documents are in the store
		results, err := store1.Search(ctx, []float32{1, 0, 0}, 2)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})

	// Close the first store
	err = store1.Close()
	assert.NoError(t, err)

	// Open a new store with the same database path
	store2, err := NewSQLiteVecVectorStoreSimple(dbPath, embedder)
	require.NoError(t, err)
	defer func() {
		_ = store2.Close()
	}()

	t.Run("Verify documents persist across store instances", func(t *testing.T) {
		results, err := store2.Search(ctx, []float32{1, 0, 0}, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		stats, err := store2.GetStats(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, stats.TotalDocuments, 2)
	})
}

// TestSQLiteVecVectorStore_WithConfig tests creating a store with full configuration
func TestSQLiteVecVectorStore_WithConfig(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 128}

	config := SQLiteVecConfig{
		DBPath:         "", // In-memory
		CollectionName: "test_collection",
		Embedder:       embedder,
	}

	store, err := NewSQLiteVecVectorStore(config)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Verify collection name", func(t *testing.T) {
		assert.Equal(t, "test_collection", store.GetCollectionName())
	})

	t.Run("Add and search with different dimensions", func(t *testing.T) {
		// Create embedding with 128 dimensions
		embedding := make([]float32, 128)
		for i := range embedding {
			embedding[i] = 0.1
		}

		docs := []rag.Document{
			{ID: "1", Content: "test doc with 128 dims", Embedding: embedding},
		}
		err := store.Add(ctx, docs)
		assert.NoError(t, err)

		results, err := store.Search(ctx, embedding, 1)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "1", results[0].Document.ID)
	})
}

// TestSQLiteVecVectorStore_ConcurrentOperations tests concurrent operations
func TestSQLiteVecVectorStore_ConcurrentOperations(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewSQLiteVecVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Concurrent adds", func(t *testing.T) {
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(idx int) {
				docs := []rag.Document{
					{ID: fmt.Sprintf("concurrent-%d", idx), Content: fmt.Sprintf("content %d", idx), Embedding: []float32{float32(idx) * 0.1, 0, 0}},
				}
				_ = store.Add(ctx, docs)
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all documents were added
		stats, _ := store.GetStats(ctx)
		assert.GreaterOrEqual(t, stats.TotalDocuments, 10)
	})
}

// TestSQLiteVecVectorStore_EmbeddingGeneration tests automatic embedding generation
func TestSQLiteVecVectorStore_EmbeddingGeneration(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewSQLiteVecVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Add document without embedding", func(t *testing.T) {
		doc := rag.Document{
			ID:      "auto-embed",
			Content: "this should be auto-embedded",
		}
		err := store.Add(ctx, []rag.Document{doc})
		assert.NoError(t, err)

		// Search with a query embedding
		results, err := store.Search(ctx, []float32{0.1, 0.1, 0.1}, 1)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)
	})
}

// TestSQLiteVecVectorStore_MetadataFiltering tests metadata filtering
func TestSQLiteVecVectorStore_MetadataFiltering(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewSQLiteVecVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Multiple metadata filters", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "1", Content: "doc 1", Embedding: []float32{1, 0, 0}, Metadata: map[string]any{"category": "tech", "year": "2023"}},
			{ID: "2", Content: "doc 2", Embedding: []float32{0, 1, 0}, Metadata: map[string]any{"category": "news", "year": "2023"}},
			{ID: "3", Content: "doc 3", Embedding: []float32{0, 0, 1}, Metadata: map[string]any{"category": "tech", "year": "2024"}},
		}
		err := store.Add(ctx, docs)
		assert.NoError(t, err)

		// Filter by category
		results, err := store.SearchWithFilter(ctx, []float32{1, 0, 0}, 10, map[string]any{"category": "tech"})
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 2)

		// Filter by both category and year
		results, err = store.SearchWithFilter(ctx, []float32{1, 0, 0}, 10, map[string]any{"category": "tech", "year": "2023"})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "1", results[0].Document.ID)
	})
}

// TestSQLiteVecVectorStore_EdgeCases tests edge cases
func TestSQLiteVecVectorStore_EdgeCases(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewSQLiteVecVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Search with k=0", func(t *testing.T) {
		_, err := store.Search(ctx, []float32{1, 0, 0}, 0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "k must be positive")
	})

	t.Run("Search on empty store", func(t *testing.T) {
		emptyStore, err := NewSQLiteVecVectorStoreSimple("", embedder)
		require.NoError(t, err)
		defer func() {
			_ = emptyStore.Close()
		}()

		results, err := emptyStore.Search(ctx, []float32{1, 0, 0}, 5)
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("Delete non-existent documents", func(t *testing.T) {
		err := store.Delete(ctx, []string{"non-existent-id"})
		assert.NoError(t, err) // Should not error
	})

	t.Run("Update non-existent document", func(t *testing.T) {
		doc := rag.Document{ID: "non-existent", Content: "updated", Embedding: []float32{1, 1, 1}}
		err := store.Update(ctx, []rag.Document{doc})
		assert.NoError(t, err) // Should add the document
	})

	t.Run("Empty document list", func(t *testing.T) {
		err := store.Add(ctx, []rag.Document{})
		assert.NoError(t, err)

		err = store.Update(ctx, []rag.Document{})
		assert.NoError(t, err)

		err = store.Delete(ctx, []string{})
		assert.NoError(t, err)
	})
}

// TestSQLiteVecVectorStore_DimensionMismatch tests dimension mismatch handling
func TestSQLiteVecVectorStore_DimensionMismatch(t *testing.T) {
	ctx := context.Background()
	embedder := &mockEmbedder{dim: 3}

	store, err := NewSQLiteVecVectorStoreSimple("", embedder)
	require.NoError(t, err)
	defer func() {
		_ = store.Close()
	}()

	t.Run("Add document with wrong dimension", func(t *testing.T) {
		wrongDimEmbedding := []float32{1, 0} // Only 2 dimensions
		docs := []rag.Document{
			{ID: "wrong-dim", Content: "wrong dimension", Embedding: wrongDimEmbedding},
		}
		err := store.Add(ctx, docs)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "embedding dimension mismatch")
	})

	t.Run("Search with wrong dimension", func(t *testing.T) {
		wrongDimQuery := []float32{1, 0} // Only 2 dimensions
		_, err := store.Search(ctx, wrongDimQuery, 5)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "query dimension mismatch")
	})
}

// TestSanitizeTableName tests the table name sanitization
func TestSanitizeTableName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"alnum", "test123", "test123"},
		{"special chars", "test-collection!", "test_collection_"},
		{"starts with number", "123test", "t123test"},
		{"empty", "", "vec_store"},
		{"only special", "-@#", "vec_store"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeTableName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSQLiteVecVectorStore_TableName tests table name handling
func TestSQLiteVecVectorStore_TableName(t *testing.T) {
	embedder := &mockEmbedder{dim: 3}

	t.Run("custom collection name", func(t *testing.T) {
		store, err := NewSQLiteVecVectorStore(SQLiteVecConfig{
			DBPath:         "",
			CollectionName: "my-collection",
			Embedder:       embedder,
		})
		require.NoError(t, err)
		defer func() {
			_ = store.Close()
		}()

		assert.Equal(t, "my-collection", store.GetCollectionName())
		assert.Equal(t, "my_collection", store.GetTableName())
	})
}
