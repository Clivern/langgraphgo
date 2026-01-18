package retriever

import (
	"context"
	"testing"
	"time"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/tokenizer"
)

func TestBM25Retriever(t *testing.T) {
	t.Run("BasicRetrieval", func(t *testing.T) {
		docs := []rag.Document{
			{
				ID:      "doc1",
				Content: "The quick brown fox jumps over the lazy dog",
				Metadata: map[string]any{
					"source": "test1",
				},
				CreatedAt: time.Now(),
			},
			{
				ID:      "doc2",
				Content: "A fast fox is quick and agile",
				Metadata: map[string]any{
					"source": "test2",
				},
				CreatedAt: time.Now(),
			},
			{
				ID:      "doc3",
				Content: "The dog is sleeping on the couch",
				Metadata: map[string]any{
					"source": "test3",
				},
				CreatedAt: time.Now(),
			},
		}

		config := DefaultBM25Config()
		config.K = 2

		retriever, err := NewBM25Retriever(docs, config)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		ctx := context.Background()
		results, err := retriever.Retrieve(ctx, "quick fox")
		if err != nil {
			t.Fatalf("failed to retrieve: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}

		// Should return documents containing "quick" or "fox"
		// Both doc1 and doc2 contain these words
		foundDoc1 := false
		foundDoc2 := false
		for _, result := range results {
			if result.ID == "doc1" {
				foundDoc1 = true
			}
			if result.ID == "doc2" {
				foundDoc2 = true
			}
		}

		if !foundDoc1 {
			t.Error("expected doc1 to be in results")
		}
		if !foundDoc2 {
			t.Error("expected doc2 to be in results")
		}
	})

	t.Run("RetrieveWithK", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "doc1", Content: "machine learning is a subset of artificial intelligence"},
			{ID: "doc2", Content: "deep learning uses neural networks"},
			{ID: "doc3", Content: "neural networks are inspired by biological neurons"},
			{ID: "doc4", Content: "artificial intelligence is transforming technology"},
		}

		config := DefaultBM25Config()
		retriever, err := NewBM25Retriever(docs, config)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		ctx := context.Background()
		results, err := retriever.RetrieveWithK(ctx, "neural networks", 3)
		if err != nil {
			t.Fatalf("failed to retrieve: %v", err)
		}

		if len(results) > 3 {
			t.Errorf("expected at most 3 results, got %d", len(results))
		}
	})

	t.Run("RetrieveWithConfig", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "doc1", Content: "golang programming language"},
			{ID: "doc2", Content: "python programming language"},
			{ID: "doc3", Content: "javascript web development"},
		}

		config := DefaultBM25Config()
		retriever, err := NewBM25Retriever(docs, config)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		ctx := context.Background()
		retrievalConfig := &rag.RetrievalConfig{
			K:              2,
			ScoreThreshold: 0.1,
		}

		results, err := retriever.RetrieveWithConfig(ctx, "programming language", retrievalConfig)
		if err != nil {
			t.Fatalf("failed to retrieve: %v", err)
		}

		for _, result := range results {
			if result.Score < retrievalConfig.ScoreThreshold {
				t.Errorf("result score %f below threshold %f", result.Score, retrievalConfig.ScoreThreshold)
			}
		}
	})

	t.Run("AddDocuments", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "doc1", Content: "initial document"},
		}

		config := DefaultBM25Config()
		retriever, err := NewBM25Retriever(docs, config)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		// Add more documents
		newDocs := []rag.Document{
			{ID: "doc2", Content: "new document about testing"},
			{ID: "doc3", Content: "another document"},
		}
		retriever.AddDocuments(newDocs)

		if retriever.GetDocumentCount() != 3 {
			t.Errorf("expected 3 documents, got %d", retriever.GetDocumentCount())
		}

		ctx := context.Background()
		results, err := retriever.Retrieve(ctx, "testing")
		if err != nil {
			t.Fatalf("failed to retrieve: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("expected results for 'testing'")
		}

		if results[0].ID != "doc2" {
			t.Errorf("expected doc2 as first result for 'testing', got %s", results[0].ID)
		}
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "doc1", Content: "document one"},
			{ID: "doc2", Content: "document two"},
			{ID: "doc3", Content: "document three"},
		}

		config := DefaultBM25Config()
		retriever, err := NewBM25Retriever(docs, config)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		// Delete document
		retriever.DeleteDocument("doc2")

		if retriever.GetDocumentCount() != 2 {
			t.Errorf("expected 2 documents after deletion, got %d", retriever.GetDocumentCount())
		}

		ctx := context.Background()
		results, err := retriever.Retrieve(ctx, "document")
		if err != nil {
			t.Fatalf("failed to retrieve: %v", err)
		}

		// Check that doc2 is not in results
		for _, result := range results {
			if result.ID == "doc2" {
				t.Error("deleted doc2 should not appear in results")
			}
		}
	})

	t.Run("UpdateDocument", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "doc1", Content: "original content"},
		}

		config := DefaultBM25Config()
		retriever, err := NewBM25Retriever(docs, config)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		// Update document
		updatedDoc := rag.Document{
			ID:      "doc1",
			Content: "updated content with new keywords",
		}
		retriever.UpdateDocument(updatedDoc)

		ctx := context.Background()
		results, err := retriever.Retrieve(ctx, "keywords")
		if err != nil {
			t.Fatalf("failed to retrieve: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("expected results for 'keywords' after update")
		}

		if results[0].Content != "updated content with new keywords" {
			t.Error("document content was not updated")
		}
	})

	t.Run("GetStats", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "doc1", Content: "test document one"},
			{ID: "doc2", Content: "test document two"},
		}

		config := DefaultBM25Config()
		retriever, err := NewBM25Retriever(docs, config)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		stats := retriever.GetStats()

		numDocs, ok := stats["num_documents"].(int)
		if !ok || numDocs != 2 {
			t.Errorf("expected num_documents to be 2, got %v", stats["num_documents"])
		}

		k1, ok := stats["k1"].(float64)
		if !ok || k1 != config.K1 {
			t.Errorf("expected k1 to be %f, got %v", config.K1, stats["k1"])
		}
	})

	t.Run("ScoreThreshold", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "doc1", Content: "machine learning and artificial intelligence"},
			{ID: "doc2", Content: "cooking recipes and food preparation"},
			{ID: "doc3", Content: "sports news and athletic events"},
		}

		config := DefaultBM25Config()
		retriever, err := NewBM25Retriever(docs, config)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		ctx := context.Background()
		retrievalConfig := &rag.RetrievalConfig{
			K:              10,
			ScoreThreshold: 1.0, // High threshold
		}

		results, err := retriever.RetrieveWithConfig(ctx, "machine learning", retrievalConfig)
		if err != nil {
			t.Fatalf("failed to retrieve: %v", err)
		}

		// All results should meet the threshold
		for _, result := range results {
			if result.Score < 1.0 {
				t.Errorf("result score %f below threshold 1.0", result.Score)
			}
		}
	})

	t.Run("Parameters", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "doc1", Content: "test document"},
		}

		config := DefaultBM25Config()
		retriever, err := NewBM25Retriever(docs, config)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		// Test SetK1
		retriever.SetK1(2.0)
		if retriever.GetK1() != 2.0 {
			t.Errorf("expected k1 to be 2.0, got %f", retriever.GetK1())
		}

		// Test SetB
		retriever.SetB(0.5)
		if retriever.GetB() != 0.5 {
			t.Errorf("expected b to be 0.5, got %f", retriever.GetB())
		}
	})
}

func TestBM25RetrieverWithCustomTokenizer(t *testing.T) {
	t.Run("CustomTokenizer", func(t *testing.T) {
		docs := []rag.Document{
			{ID: "doc1", Content: "The quick brown fox"},
			{ID: "doc2", Content: "Fast animals are quick"},
		}

		config := DefaultBM25Config()
		customTokenizer := tokenizer.DefaultRegexTokenizer()

		retriever, err := NewBM25RetrieverWithTokenizer(docs, config, customTokenizer)
		if err != nil {
			t.Fatalf("failed to create BM25 retriever: %v", err)
		}

		ctx := context.Background()
		results, err := retriever.Retrieve(ctx, "quick")
		if err != nil {
			t.Fatalf("failed to retrieve: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}
	})
}
