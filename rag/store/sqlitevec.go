package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	sqlitevec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smallnest/langgraphgo/rag"
)

// SQLiteVecVectorStore is a vector store implementation using sqlite-vec
// It provides persistent vector storage with SQLite backend using CGO
type SQLiteVecVectorStore struct {
	db             *sql.DB
	embedder       rag.Embedder
	tableName      string
	dimension      int
	mu             sync.RWMutex
	collectionName string
}

// SQLiteVecConfig contains configuration for SQLiteVecVectorStore
type SQLiteVecConfig struct {
	// DBPath is the path to the SQLite database file
	// If empty, uses in-memory storage
	DBPath string

	// CollectionName is the name of the collection/table to use
	// If empty, uses "default"
	CollectionName string

	// Embedder is the embedder to use for generating embeddings
	Embedder rag.Embedder

	// Dimension is the dimension of the vectors
	// If 0, attempts to detect from embedder
	Dimension int
}

// NewSQLiteVecVectorStore creates a new SQLiteVecVectorStore with the given configuration
func NewSQLiteVecVectorStore(config SQLiteVecConfig) (*SQLiteVecVectorStore, error) {
	var db *sql.DB
	var err error

	// Register sqlite-vec extension for all new connections
	sqlitevec.Auto()

	dsn := config.DBPath
	if dsn == "" {
		dsn = ":memory:"
	}

	// Open database using mattn/go-sqlite3 with sqlite-vec extension
	db, err = sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if config.Embedder == nil {
		return nil, fmt.Errorf("embedder is required")
	}

	collectionName := config.CollectionName
	if collectionName == "" {
		collectionName = "default"
	}

	dimension := config.Dimension
	if dimension == 0 {
		dimension = config.Embedder.GetDimension()
	}

	if dimension <= 0 {
		return nil, fmt.Errorf("invalid dimension: %d", dimension)
	}

	s := &SQLiteVecVectorStore{
		db:             db,
		embedder:       config.Embedder,
		collectionName: collectionName,
		dimension:      dimension,
	}

	// Create table with vec0 for vector operations
	tableName := sanitizeTableName(collectionName)
	s.tableName = tableName

	if err := s.initSchema(ctxBackground()); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return s, nil
}

// NewSQLiteVecVectorStoreSimple creates a new SQLiteVecVectorStore with simple parameters
// For in-memory storage, pass an empty string for dbPath
func NewSQLiteVecVectorStoreSimple(dbPath string, embedder rag.Embedder) (*SQLiteVecVectorStore, error) {
	return NewSQLiteVecVectorStore(SQLiteVecConfig{
		DBPath:   dbPath,
		Embedder: embedder,
	})
}

// initSchema creates the necessary tables for vector storage
func (s *SQLiteVecVectorStore) initSchema(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create virtual table for vector search using vec0 with auxiliary columns
	// Use double quotes for table name to handle SQL keywords
	// #nosec G201 - table names are sanitized and not user input
	createVecTableSQL := fmt.Sprintf(`
		CREATE VIRTUAL TABLE IF NOT EXISTS "%s" USING vec0(
			embedding float[%d],
			id TEXT PRIMARY KEY,
			content TEXT,
			metadata TEXT,
			created_at INTEGER,
			updated_at INTEGER
		)
	`, s.tableName, s.dimension)

	if _, err := s.db.ExecContext(ctx, createVecTableSQL); err != nil {
		return fmt.Errorf("failed to create vec table: %w", err)
	}

	return nil
}

// Add adds documents to the sqlite-vec vector store
func (s *SQLiteVecVectorStore) Add(ctx context.Context, documents []rag.Document) error {
	if len(documents) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// #nosec G201 - table names are sanitized and not user input
	insertSQL := fmt.Sprintf(`
		INSERT OR REPLACE INTO "%s"(id, embedding, content, metadata, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, s.tableName)

	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now().Unix()

	for _, doc := range documents {
		// Generate embedding if not provided
		var embedding []float32
		if len(doc.Embedding) > 0 {
			embedding = doc.Embedding
		} else {
			var err error
			embedding, err = s.embedder.EmbedDocument(ctx, doc.Content)
			if err != nil {
				return fmt.Errorf("failed to generate embedding for %s: %w", doc.ID, err)
			}
		}

		// Validate embedding dimension
		if len(embedding) != s.dimension {
			return fmt.Errorf("embedding dimension mismatch for %s: expected %d, got %d",
				doc.ID, s.dimension, len(embedding))
		}

		// Serialize metadata
		var metadataJSON []byte
		if len(doc.Metadata) > 0 {
			var err error
			metadataJSON, err = json.Marshal(doc.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal metadata for %s: %w", doc.ID, err)
			}
		}
		// Convert nil to empty string for sqlite compatibility
		metadataStr := ""
		if metadataJSON != nil {
			metadataStr = string(metadataJSON)
		}

		// Serialize embedding
		embeddingBlob, err := sqlitevec.SerializeFloat32(embedding)
		if err != nil {
			return fmt.Errorf("failed to serialize embedding for %s: %w", doc.ID, err)
		}

		// Insert document
		createdAt := doc.CreatedAt.Unix()
		updatedAt := doc.UpdatedAt.Unix()
		if createdAt == 0 {
			createdAt = now
		}
		if updatedAt == 0 {
			updatedAt = now
		}

		if _, err := stmt.ExecContext(ctx, doc.ID, embeddingBlob, doc.Content, metadataStr, createdAt, updatedAt); err != nil {
			return fmt.Errorf("failed to insert document %s: %w", doc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Search performs similarity search in the sqlite-vec vector store
func (s *SQLiteVecVectorStore) Search(ctx context.Context, query []float32, k int) ([]rag.DocumentSearchResult, error) {
	return s.SearchWithFilter(ctx, query, k, nil)
}

// SearchWithFilter performs similarity search with metadata filters
func (s *SQLiteVecVectorStore) SearchWithFilter(ctx context.Context, query []float32, k int, filter map[string]any) ([]rag.DocumentSearchResult, error) {
	if k <= 0 {
		return nil, fmt.Errorf("k must be positive")
	}

	if len(query) != s.dimension {
		return nil, fmt.Errorf("query dimension mismatch: expected %d, got %d", s.dimension, len(query))
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Serialize query
	queryBlob, err := sqlitevec.SerializeFloat32(query)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize query: %w", err)
	}

	// Build the SQL query for vector search using sqlite-vec
	// Note: Filtering is done in-memory after vector search for simplicity
	// sqlite-vec's vec0 table has limitations on filtering auxiliary columns
	// #nosec G201 - table names are sanitized and not user input
	querySQL := fmt.Sprintf(`
		SELECT
			id,
			content,
			metadata,
			created_at,
			updated_at,
			distance
		FROM "%s"
		WHERE embedding MATCH ?
		ORDER BY distance
		LIMIT ?
	`, s.tableName)

	args := []any{queryBlob, k * 10} // Get more results for filtering

	rows, err := s.db.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []rag.DocumentSearchResult
	for rows.Next() {
		var id string
		var content string
		var metadataJSON sql.NullString
		var createdAt int64
		var updatedAt int64
		var distance float64

		if err := rows.Scan(&id, &content, &metadataJSON, &createdAt, &updatedAt, &distance); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var metadata map[string]any
		if metadataJSON.Valid && metadataJSON.String != "" {
			if err := json.Unmarshal([]byte(metadataJSON.String), &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata for %s: %w", id, err)
			}
		}

		// Apply metadata filter
		if len(filter) > 0 {
			matches := true
			for key, value := range filter {
				if metadataValue, ok := metadata[key]; !ok || metadataValue != value {
					matches = false
					break
				}
			}
			if !matches {
				continue
			}
		}

		// Convert distance to similarity (cosine similarity = 1 - distance for normalized vectors)
		similarity := 1.0 - distance

		results = append(results, rag.DocumentSearchResult{
			Document: rag.Document{
				ID:        id,
				Content:   content,
				Metadata:  metadata,
				CreatedAt: time.Unix(createdAt, 0),
				UpdatedAt: time.Unix(updatedAt, 0),
			},
			Score: similarity,
		})

		// Stop if we have enough results
		if len(results) >= k {
			break
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// Delete removes documents by their IDs
func (s *SQLiteVecVectorStore) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// #nosec G201 - table names are sanitized and not user input
	deleteSQL := fmt.Sprintf("DELETE FROM \"%s\" WHERE id = ?", s.tableName)
	stmt, err := tx.PrepareContext(ctx, deleteSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare delete: %w", err)
	}
	defer stmt.Close()

	for _, id := range ids {
		if _, err := stmt.ExecContext(ctx, id); err != nil {
			return fmt.Errorf("failed to delete document %s: %w", id, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Update updates documents in the sqlite-vec vector store
func (s *SQLiteVecVectorStore) Update(ctx context.Context, documents []rag.Document) error {
	return s.Add(ctx, documents)
}

// GetStats returns statistics about the sqlite-vec vector store
func (s *SQLiteVecVectorStore) GetStats(ctx context.Context) (*rag.VectorStoreStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// #nosec G201 - table names are sanitized and not user input
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM \"%s\"", s.tableName)
	var count int
	if err := s.db.QueryRowContext(ctx, countSQL).Scan(&count); err != nil {
		return nil, fmt.Errorf("failed to get document count: %w", err)
	}

	return &rag.VectorStoreStats{
		TotalDocuments: count,
		TotalVectors:   count,
		Dimension:      s.dimension,
		LastUpdated:    time.Now(),
	}, nil
}

// Close closes the sqlite-vec vector store and releases resources
func (s *SQLiteVecVectorStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetTableName returns the name of the table
func (s *SQLiteVecVectorStore) GetTableName() string {
	return s.tableName
}

// GetCollectionName returns the name of the collection
func (s *SQLiteVecVectorStore) GetCollectionName() string {
	return s.collectionName
}

// sanitizeTableName sanitizes the collection name for use as a table name
func sanitizeTableName(name string) string {
	// Replace any non-alphanumeric characters with underscores
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' {
			result = append(result, c)
		} else {
			result = append(result, '_')
		}
	}
	// Ensure it starts with a letter
	if len(result) > 0 && (result[0] >= '0' && result[0] <= '9') {
		result = append([]byte{'t'}, result...)
	}
	if len(result) == 0 {
		return "vec_store"
	}
	return string(result)
}

// ctxBackground returns a background context for initialization
func ctxBackground() context.Context {
	return context.Background()
}
