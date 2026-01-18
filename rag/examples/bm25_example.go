package examples

import (
	"context"
	"fmt"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/retriever"
	"github.com/smallnest/langgraphgo/rag/tokenizer"
)

// BM25BasicExample demonstrates basic BM25 retrieval
func BM25BasicExample() {
	// Create sample documents
	docs := []rag.Document{
		{
			ID:      "doc1",
			Content: "Machine learning is a subset of artificial intelligence that enables systems to learn from data.",
			Metadata: map[string]any{
				"title":    "Introduction to ML",
				"category": "AI",
			},
		},
		{
			ID:      "doc2",
			Content: "Deep learning is a type of machine learning that uses neural networks with multiple layers.",
			Metadata: map[string]any{
				"title":    "Deep Learning Basics",
				"category": "AI",
			},
		},
		{
			ID:      "doc3",
			Content: "Natural language processing enables computers to understand and generate human language.",
			Metadata: map[string]any{
				"title":    "NLP Overview",
				"category": "AI",
			},
		},
		{
			ID:      "doc4",
			Content: "Computer vision allows machines to interpret and make decisions based on visual data.",
			Metadata: map[string]any{
				"title":    "Computer Vision",
				"category": "AI",
			},
		},
		{
			ID:      "doc5",
			Content: "Reinforcement learning is an area of machine learning concerned with how agents take actions in an environment.",
			Metadata: map[string]any{
				"title":    "Reinforcement Learning",
				"category": "AI",
			},
		},
	}

	// Create BM25 retriever with default configuration
	config := retriever.DefaultBM25Config()
	config.K = 3 // Retrieve top 3 documents

	bm25Retriever, err := retriever.NewBM25Retriever(docs, config)
	if err != nil {
		fmt.Printf("Error creating BM25 retriever: %v\n", err)
		return
	}

	// Query for documents
	ctx := context.Background()
	query := "neural networks machine learning"

	results, err := bm25Retriever.Retrieve(ctx, query)
	if err != nil {
		fmt.Printf("Error retrieving documents: %v\n", err)
		return
	}

	// Display results
	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Found %d results:\n\n", len(results))

	for i, result := range results {
		fmt.Printf("%d. ID: %s\n", i+1, result.ID)
		if title, ok := result.Metadata["title"]; ok {
			fmt.Printf("   Title: %s\n", title)
		}
		fmt.Printf("   Content: %s\n\n", result.Content)
	}
}

// BM25WithScoreThresholdExample demonstrates BM25 retrieval with score threshold
func BM25WithScoreThresholdExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "Python is a high-level programming language"},
		{ID: "doc2", Content: "Golang is a statically typed programming language"},
		{ID: "doc3", Content: "JavaScript is used for web development"},
		{ID: "doc4", Content: "Rust focuses on safety and performance"},
		{ID: "doc5", Content: "TypeScript adds types to JavaScript"},
	}

	config := retriever.DefaultBM25Config()
	config.K = 10
	config.ScoreThreshold = 0.5 // Only return documents with score >= 0.5

	bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

	ctx := context.Background()
	query := "programming language development"

	// Use RetrieveWithConfig to get scores
	retrievalConfig := &rag.RetrievalConfig{
		K:              config.K,
		ScoreThreshold: config.ScoreThreshold,
	}
	results, _ := bm25Retriever.RetrieveWithConfig(ctx, query, retrievalConfig)

	fmt.Printf("Query: %s (Score Threshold: %.1f)\n", query, config.ScoreThreshold)
	fmt.Printf("Found %d results above threshold:\n\n", len(results))

	for i, result := range results {
		fmt.Printf("%d. Score: %.4f - %s\n", i+1, result.Score, result.Document.Content)
	}
}

// BM25WithCustomTokenizerExample demonstrates BM25 with custom tokenization
func BM25WithCustomTokenizerExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "Go语言是一门静态类型的编程语言"},
		{ID: "doc2", Content: "Python是一种动态类型的高级编程语言"},
		{ID: "doc3", Content: "JavaScript是Web开发的核心语言"},
	}

	// Use Chinese tokenizer for Chinese text
	chineseTokenizer := tokenizer.NewChineseTokenizer()

	config := retriever.DefaultBM25Config()
	config.K = 2

	bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(docs, config, chineseTokenizer)

	ctx := context.Background()
	query := "编程语言"

	results, _ := bm25Retriever.Retrieve(ctx, query)

	fmt.Printf("Query: %s\n", query)
	fmt.Printf("Found %d results:\n\n", len(results))

	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Content)
	}
}

// BM25HybridRetrievalExample demonstrates hybrid retrieval using BM25 and Vector retrievers
func BM25HybridRetrievalExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "LangGraph is a framework for building stateful, multi-actor applications with LLMs"},
		{ID: "doc2", Content: "LangChain is a framework for developing applications powered by language models"},
		{ID: "doc3", Content: "RAG (Retrieval-Augmented Generation) combines retrieval with generation"},
		{ID: "doc4", Content: "Vector stores enable semantic search using embeddings"},
		{ID: "doc5", Content: "BM25 is a ranking function for information retrieval"},
	}

	// Create BM25 retriever
	bm25Config := retriever.DefaultBM25Config()
	bm25Config.K = 5
	bm25Retriever, _ := retriever.NewBM25Retriever(docs, bm25Config)

	// Create a mock vector retriever for demonstration
	// In practice, you would use a real vector store and embedder
	// vectorRetriever := retriever.NewVectorRetriever(vectorStore, embedder, config)

	// Create hybrid retriever combining BM25 and vector retrieval
	// hybridRetriever := retriever.NewHybridRetriever(
	//     []rag.Retriever{bm25Retriever, vectorRetriever},
	//     []float64{0.5, 0.5}, // Equal weights
	//     rag.RetrievalConfig{K: 3},
	// )

	ctx := context.Background()
	query := "framework for building applications with language models"

	// Use BM25 retriever
	fmt.Printf("BM25 Retrieval:\n")
	fmt.Printf("Query: %s\n\n", query)

	bm25Results, _ := bm25Retriever.RetrieveWithConfig(ctx, query, &rag.RetrievalConfig{
		K:              3,
		ScoreThreshold: 0.0,
		IncludeScores:  true,
	})

	for i, result := range bm25Results {
		fmt.Printf("%d. [Score: %.4f] %s\n", i+1, result.Score, result.Document.Content)
	}

	// Use hybrid retriever
	// hybridResults, _ := hybridRetriever.Retrieve(ctx, query)
	// fmt.Printf("\nHybrid Retrieval Results:\n")
	// for i, result := range hybridResults {
	//     fmt.Printf("%d. %s\n", i+1, result.Content)
	// }
}

// BM25DynamicUpdateExample demonstrates dynamic document management
func BM25DynamicUpdateExample() {
	// Initial documents
	docs := []rag.Document{
		{ID: "doc1", Content: "Initial document about AI"},
		{ID: "doc2", Content: "Initial document about ML"},
	}

	config := retriever.DefaultBM25Config()
	bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

	fmt.Printf("Initial document count: %d\n\n", bm25Retriever.GetDocumentCount())

	// Add new documents
	newDocs := []rag.Document{
		{ID: "doc3", Content: "New document about deep learning"},
		{ID: "doc4", Content: "New document about neural networks"},
	}
	bm25Retriever.AddDocuments(newDocs)

	fmt.Printf("After adding documents: %d\n\n", bm25Retriever.GetDocumentCount())

	// Query to see new documents
	ctx := context.Background()
	results, _ := bm25Retriever.Retrieve(ctx, "deep learning neural networks")

	fmt.Printf("Query: deep learning neural networks\n")
	fmt.Printf("Results:\n")
	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Content)
	}

	// Update a document
	fmt.Printf("\nUpdating doc1...\n")
	bm25Retriever.UpdateDocument(rag.Document{
		ID:      "doc1",
		Content: "Updated document about artificial intelligence and machine learning",
	})

	results, _ = bm25Retriever.Retrieve(ctx, "artificial intelligence")
	fmt.Printf("\nQuery: artificial intelligence\n")
	fmt.Printf("Results after update:\n")
	for i, result := range results {
		fmt.Printf("%d. %s\n", i+1, result.Content)
	}

	// Delete a document
	fmt.Printf("\nDeleting doc2...\n")
	bm25Retriever.DeleteDocument("doc2")
	fmt.Printf("Document count after deletion: %d\n", bm25Retriever.GetDocumentCount())
}

// BM25ParameterTuningExample demonstrates tuning BM25 parameters
func BM25ParameterTuningExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "The quick brown fox jumps over the lazy dog"},
		{ID: "doc2", Content: "A quick movement of the enemy will jeopardize five gunboats"},
		{ID: "doc3", Content: "The five boxing wizards jump quickly"},
		{ID: "doc4", Content: "Pack my box with five dozen liquor jugs"},
		{ID: "doc5", Content: "The jumpy fox and the quick dog are friends"},
	}

	query := "quick fox jump"

	// Test different k1 values (controls term frequency saturation)
	fmt.Println("Testing different k1 values:")
	fmt.Println("==========================")

	k1Values := []float64{0.5, 1.0, 1.5, 2.0}
	ctx := context.Background()

	for _, k1 := range k1Values {
		config := retriever.DefaultBM25Config()
		config.K = 3
		config.K1 = k1

		bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)
		results, _ := bm25Retriever.RetrieveWithConfig(ctx, query, &rag.RetrievalConfig{K: 3})

		fmt.Printf("\nk1 = %.1f:\n", k1)
		for i, result := range results {
			fmt.Printf("  %d. [Score: %.4f] %s\n", i+1, result.Score, result.Document.Content)
		}
	}

	// Test different b values (controls document length normalization)
	fmt.Println("\n\nTesting different b values:")
	fmt.Println("==========================")

	bValues := []float64{0.0, 0.5, 0.75, 1.0}

	for _, b := range bValues {
		config := retriever.DefaultBM25Config()
		config.K = 3
		config.B = b

		bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)
		results, _ := bm25Retriever.RetrieveWithConfig(ctx, query, &rag.RetrievalConfig{K: 3})

		fmt.Printf("\nb = %.2f:\n", b)
		for i, result := range results {
			fmt.Printf("  %d. [Score: %.4f] %s\n", i+1, result.Score, result.Document.Content)
		}
	}
}

// BM25StatsExample demonstrates accessing index statistics
func BM25StatsExample() {
	docs := []rag.Document{
		{ID: "doc1", Content: "Machine learning is transforming artificial intelligence"},
		{ID: "doc2", Content: "Deep learning uses neural networks for pattern recognition"},
		{ID: "doc3", Content: "Natural language processing enables human-computer interaction"},
		{ID: "doc4", Content: "Computer vision allows machines to interpret visual information"},
		{ID: "doc5", Content: "Reinforcement learning trains agents through reward systems"},
	}

	config := retriever.DefaultBM25Config()
	bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

	// Get and display statistics
	stats := bm25Retriever.GetStats()

	fmt.Println("BM25 Index Statistics:")
	fmt.Println("======================")
	fmt.Printf("Number of documents: %v\n", stats["num_documents"])
	fmt.Printf("Number of unique terms: %v\n", stats["num_unique_terms"])
	fmt.Printf("Average document length: %.2f tokens\n", stats["avg_doc_length"])
	fmt.Printf("k1 parameter: %.2f\n", stats["k1"])
	fmt.Printf("b parameter: %.2f\n", stats["b"])
}
