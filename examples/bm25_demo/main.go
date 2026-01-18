package main

import (
	"context"
	"fmt"
	"log"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/retriever"
	"github.com/smallnest/langgraphgo/rag/tokenizer"
)

func main() {
	fmt.Println("=== BM25 Retriever Demo ===\n")

	// Example 1: Basic BM25 retrieval
	fmt.Println("1. Basic BM25 Retrieval")
	fmt.Println("-----------------------")
	basicExample()

	// Example 2: BM25 with score threshold
	fmt.Println("\n2. BM25 with Score Threshold")
	fmt.Println("----------------------------")
	scoreThresholdExample()

	// Example 3: BM25 with custom tokenizer
	fmt.Println("\n3. BM25 with Custom Tokenizer")
	fmt.Println("------------------------------")
	customTokenizerExample()

	// Example 4: Hybrid retrieval (BM25 + Vector)
	fmt.Println("\n4. Hybrid Retrieval Setup")
	fmt.Println("-------------------------")
	hybridRetrievalExample()

	// Example 5: Dynamic document management
	fmt.Println("\n5. Dynamic Document Management")
	fmt.Println("-------------------------------")
	dynamicDocumentExample()

	// Example 6: Parameter tuning
	fmt.Println("\n6. Parameter Tuning")
	fmt.Println("-------------------")
	parameterTuningExample()
}

func basicExample() {
	docs := []rag.Document{
		{
			ID:      "doc1",
			Content: "LangGraph is a framework for building stateful, multi-actor applications with LLMs",
			Metadata: map[string]any{
				"title": "LangGraph Overview",
				"type":  "documentation",
			},
		},
		{
			ID:      "doc2",
			Content: "LangChain provides integrations and interfaces for working with LLMs",
			Metadata: map[string]any{
				"title": "LangChain Overview",
				"type":  "documentation",
			},
		},
		{
			ID:      "doc3",
			Content: "RAG combines retrieval systems with generation capabilities for better LLM responses",
			Metadata: map[string]any{
				"title": "RAG Overview",
				"type":  "tutorial",
			},
		},
		{
			ID:      "doc4",
			Content: "BM25 is a ranking function used in information retrieval to estimate document relevance",
			Metadata: map[string]any{
				"title": "BM25 Overview",
				"type":  "reference",
			},
		},
	}

	config := retriever.DefaultBM25Config()
	config.K = 2

	bm25Retriever, err := retriever.NewBM25Retriever(docs, config)
	if err != nil {
		log.Fatalf("Failed to create BM25 retriever: %v", err)
	}

	ctx := context.Background()
	query := "framework for building LLM applications"

	fmt.Printf("Query: %s\n", query)
	results, err := bm25Retriever.Retrieve(ctx, query)
	if err != nil {
		log.Fatalf("Failed to retrieve: %v", err)
	}

	fmt.Printf("Found %d results:\n", len(results))
	for i, result := range results {
		fmt.Printf("  %d. [%s] %s\n", i+1, result.Metadata["title"], result.Content)
	}
}

func scoreThresholdExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "Machine learning algorithms learn patterns from data"},
		{ID: "doc2", Content: "Deep learning uses neural networks with multiple layers"},
		{ID: "doc3", Content: "Cooking is the art of preparing food using heat"},
		{ID: "doc4", Content: "Neural networks are inspired by biological neurons"},
		{ID: "doc5", Content: "Baking involves cooking with dry heat in an oven"},
	}

	config := retriever.DefaultBM25Config()
	config.K = 10
	config.ScoreThreshold = 0.5

	bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

	ctx := context.Background()
	query := "neural networks learning"

	retrievalConfig := &rag.RetrievalConfig{
		K:              10,
		ScoreThreshold: config.ScoreThreshold,
	}

	fmt.Printf("Query: %s (threshold: %.1f)\n", query, config.ScoreThreshold)
	results, _ := bm25Retriever.RetrieveWithConfig(ctx, query, retrievalConfig)

	fmt.Printf("Found %d results above threshold:\n", len(results))
	for i, result := range results {
		fmt.Printf("  %d. [Score: %.4f] %s\n", i+1, result.Score, result.Document.Content)
	}
}

func customTokenizerExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "Go语言支持并发编程"},
		{ID: "doc2", Content: "Python是一种易学的编程语言"},
		{ID: "doc3", Content: "JavaScript用于前端开发"},
	}

	// Use Chinese tokenizer
	chineseTokenizer := tokenizer.NewChineseTokenizer()

	config := retriever.DefaultBM25Config()
	config.K = 2

	bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(docs, config, chineseTokenizer)

	ctx := context.Background()
	query := "编程语言"

	fmt.Printf("Query: %s\n", query)
	results, _ := bm25Retriever.Retrieve(ctx, query)

	fmt.Printf("Found %d results:\n", len(results))
	for i, result := range results {
		fmt.Printf("  %d. %s\n", i+1, result.Content)
	}
}

func hybridRetrievalExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "Vector search uses embeddings for semantic similarity"},
		{ID: "doc2", Content: "BM25 uses term frequency for keyword matching"},
		{ID: "doc3", Content: "Hybrid search combines both approaches"},
	}

	config := retriever.DefaultBM25Config()
	config.K = 3

	bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

	ctx := context.Background()
	query := "search similarity matching"

	retrievalConfig := &rag.RetrievalConfig{K: 3}
	results, _ := bm25Retriever.RetrieveWithConfig(ctx, query, retrievalConfig)

	fmt.Printf("BM25 results for '%s':\n", query)
	for i, result := range results {
		fmt.Printf("  %d. [Score: %.4f] %s\n", i+1, result.Score, result.Document.Content)
	}

	fmt.Println("\nTo use hybrid retrieval, combine BM25Retriever with VectorRetriever:")
	fmt.Println("  hybridRetriever := retriever.NewHybridRetriever(")
	fmt.Println("    []rag.Retriever{bm25Retriever, vectorRetriever},")
	fmt.Println("    []float64{0.5, 0.5},")
	fmt.Println("    rag.RetrievalConfig{K: 3},")
	fmt.Println("  )")
}

func dynamicDocumentExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "Initial document"},
	}

	config := retriever.DefaultBM25Config()
	bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

	fmt.Printf("Initial document count: %d\n", bm25Retriever.GetDocumentCount())

	// Add documents
	newDocs := []rag.Document{
		{ID: "doc2", Content: "Added document about AI"},
		{ID: "doc3", Content: "Added document about ML"},
	}
	bm25Retriever.AddDocuments(newDocs)
	fmt.Printf("After adding: %d\n", bm25Retriever.GetDocumentCount())

	// Update document
	bm25Retriever.UpdateDocument(rag.Document{
		ID:      "doc1",
		Content: "Updated document with new content",
	})
	fmt.Println("Updated doc1")

	// Delete document
	bm25Retriever.DeleteDocument("doc2")
	fmt.Printf("After deleting doc2: %d\n", bm25Retriever.GetDocumentCount())

	ctx := context.Background()
	results, _ := bm25Retriever.Retrieve(ctx, "new content")
	fmt.Printf("Query 'new content' found %d results\n", len(results))
}

func parameterTuningExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "The quick brown fox"},
		{ID: "doc2", Content: "Quick movement is fast"},
		{ID: "doc3", Content: "Fast and agile animals"},
	}

	query := "quick fast"

	fmt.Printf("Query: %s\n\n", query)

	// Test different k1 values
	fmt.Println("Testing k1 parameter (term frequency saturation):")
	k1Values := []float64{0.5, 1.5, 2.5}

	for _, k1 := range k1Values {
		config := retriever.DefaultBM25Config()
		config.K = 3
		config.K1 = k1

		bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

		ctx := context.Background()
		results, _ := bm25Retriever.RetrieveWithConfig(ctx, query, &rag.RetrievalConfig{K: 3})

		fmt.Printf("  k1=%.1f: ", k1)
		for _, result := range results {
			fmt.Printf("[%s:%.2f] ", result.Document.ID, result.Score)
		}
		fmt.Println()
	}

	// Get statistics
	bm25Retriever, _ := retriever.NewBM25Retriever(docs, retriever.DefaultBM25Config())
	stats := bm25Retriever.GetStats()

	fmt.Printf("\nIndex Statistics:\n")
	fmt.Printf("  Documents: %v\n", stats["num_documents"])
	fmt.Printf("  Unique terms: %v\n", stats["num_unique_terms"])
	fmt.Printf("  Avg doc length: %.1f\n", stats["avg_doc_length"])
}
