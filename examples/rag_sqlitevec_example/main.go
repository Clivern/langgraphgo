package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/retriever"
	"github.com/smallnest/langgraphgo/rag/store"
	"github.com/tmc/langchaingo/llms/openai"
)

func main() {
	ctx := context.Background()

	// Initialize LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
	}

	// Create sample documents about sqlite-vec
	documents := []rag.Document{
		{
			ID: "doc1",
			Content: "sqlite-vec is an extremely small vector search SQLite extension written in pure C. " +
				"It runs anywhere SQLite runs, including in the browser with WebAssembly.",
			Metadata: map[string]any{"source": "sqlite_vec_docs", "category": "introduction"},
		},
		{
			ID: "doc2",
			Content: "Key features of sqlite-vec include: no dependencies, pure C implementation, " +
				"support for float32, int8, and binary vectors, and KNN-style vector search queries.",
			Metadata: map[string]any{"source": "sqlite_vec_docs", "category": "features"},
		},
		{
			ID: "doc3",
			Content: "sqlite-vec provides persistent storage using standard SQLite files. " +
				"Vectors are stored in vec0 virtual tables with optional auxiliary columns for metadata.",
			Metadata: map[string]any{"source": "sqlite_vec_docs", "category": "storage"},
		},
		{
			ID: "doc4",
			Content: "Vector search in sqlite-vec uses the MATCH operator with distance metrics. " +
				"Supports cosine similarity and L2 distance for nearest neighbor queries.",
			Metadata: map[string]any{"source": "sqlite_vec_docs", "category": "search"},
		},
		{
			ID: "doc5",
			Content: "LangGraphGo integrates sqlite-vec through the CGO bindings, providing a lightweight " +
				"embedded vector database option for RAG applications without external dependencies.",
			Metadata: map[string]any{"source": "langgraphgo_docs", "category": "integration"},
		},
		{
			ID: "doc6",
			Content: "The sqlite-vec extension supports multiple programming languages including Python, Node.js, " +
				"Ruby, Rust, and Go through native bindings and WASM compilation.",
			Metadata: map[string]any{"source": "sqlite_vec_docs", "category": "languages"},
		},
	}

	fmt.Println("=== RAG with sqlite-vec VectorStore Example ===")
	fmt.Println()
	fmt.Println("Initializing sqlite-vec vector store...")

	// Create a temporary directory for persistent storage
	tempDir := filepath.Join(os.TempDir(), "sqlitevec_example")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// For production use, you would use a real embedder like OpenAI:
	// llmForEmbeddings, err := openai.New(openai.WithEmbeddingModel("text-embedding-3-small"))
	// if err != nil {
	// 	log.Fatalf("Failed to create LLM for embeddings: %v", err)
	// }
	// openaiEmbedder, err := embeddings.NewEmbedder(llmForEmbeddings)
	// if err != nil {
	// 	log.Fatalf("Failed to create OpenAI embedder: %v", err)
	// }
	// embedder := rag.NewLangChainEmbedder(openaiEmbedder)

	// For demo purposes, use mock embedder
	embedder := store.NewMockEmbedder(128)

	// Create sqlite-vec vector store with persistent storage
	sqliteVecStore, err := store.NewSQLiteVecVectorStore(store.SQLiteVecConfig{
		DBPath:         filepath.Join(tempDir, "vectors.db"),
		CollectionName: "langgraphgo_example",
		Embedder:       embedder,
	})
	if err != nil {
		log.Fatalf("Failed to create sqlite-vec store: %v", err)
	}
	defer sqliteVecStore.Close()

	fmt.Printf("Store created with collection: %s\n", sqliteVecStore.GetCollectionName())
	fmt.Printf("Table name: %s\n", sqliteVecStore.GetTableName())
	fmt.Printf("Storage location: %s\n\n", filepath.Join(tempDir, "vectors.db"))

	fmt.Println("Adding documents to sqlite-vec...")
	err = sqliteVecStore.Add(ctx, documents)
	if err != nil {
		log.Fatalf("Failed to add documents: %v", err)
	}
	fmt.Printf("Successfully added %d documents\n\n", len(documents))

	// Display store statistics
	stats, err := sqliteVecStore.GetStats(ctx)
	if err != nil {
		log.Fatalf("Failed to get stats: %v", err)
	}
	fmt.Printf("Store Statistics:\n")
	fmt.Printf("  Total Documents: %d\n", stats.TotalDocuments)
	fmt.Printf("  Vector Dimension: %d\n\n", stats.Dimension)

	// Create retriever
	vectorRetriever := retriever.NewVectorStoreRetriever(sqliteVecStore, embedder, 2)

	// Configure RAG pipeline
	config := rag.DefaultPipelineConfig()
	config.Retriever = vectorRetriever
	config.LLM = llm

	// Build basic RAG pipeline
	fmt.Println("Building RAG pipeline...")
	pipeline := rag.NewRAGPipeline(config)
	err = pipeline.BuildBasicRAG()
	if err != nil {
		log.Fatalf("Failed to build RAG pipeline: %v", err)
	}

	// Compile the pipeline
	runnable, err := pipeline.Compile()
	if err != nil {
		log.Fatalf("Failed to compile pipeline: %v", err)
	}

	// Visualize the graph
	fmt.Println("\nPipeline Graph:")
	exporter := graph.GetGraphForRunnable(runnable)
	fmt.Println(exporter.DrawASCII())
	fmt.Println()

	// Run queries
	queries := []string{
		"What is sqlite-vec?",
		"What are the key features?",
		"Tell me about the storage options",
		"Which programming languages are supported?",
	}

	for i, query := range queries {
		fmt.Println("================================================================================")
		fmt.Printf("Query %d: %s\n", i+1, query)
		fmt.Println("--------------------------------------------------------------------------------")

		result, err := runnable.Invoke(ctx, map[string]any{
			"query": query,
		})
		if err != nil {
			log.Printf("Failed to process query: %v", err)
			continue
		}

		if answer, ok := result["answer"].(string); ok {
			fmt.Printf("\nAnswer:\n%s\n", answer)
		}

		if docs, ok := result["documents"].([]rag.RAGDocument); ok {
			fmt.Printf("\nRetrieved %d documents:\n", len(docs))
			for j, doc := range docs {
				fmt.Printf("  [%d] %s\n", j+1, truncate(doc.Content, 100))
				fmt.Printf("      Metadata: %v\n", doc.Metadata)
			}
		}
		fmt.Println()
	}

	// Demonstrate metadata filtering
	fmt.Println("================================================================================")
	fmt.Println("Metadata Filtering Example")
	fmt.Println("--------------------------------------------------------------------------------")

	// Generate query embedding for filtering
	queryEmbedding, err := embedder.EmbedDocument(ctx, "features and capabilities")
	if err != nil {
		log.Printf("Failed to generate query embedding: %v", err)
	} else {
		// Search with metadata filter
		filteredResults, err := sqliteVecStore.SearchWithFilter(ctx, queryEmbedding, 10, map[string]any{
			"category": "features",
		})
		if err != nil {
			log.Printf("Failed to search with filter: %v", err)
		} else {
			fmt.Printf("\nFound %d documents with category='features':\n", len(filteredResults))
			for i, result := range filteredResults {
				fmt.Printf("  [%d] %s\n", i+1, truncate(result.Document.Content, 80))
				fmt.Printf("      Score: %.4f\n", result.Score)
				fmt.Printf("      Category: %v\n", result.Document.Metadata["category"])
			}
		}
	}
	fmt.Println()

	// Demonstrate persistent storage (reopen store)
	fmt.Println("================================================================================")
	fmt.Println("Persistent Storage Verification")
	fmt.Println("--------------------------------------------------------------------------------")

	// Close and reopen the store to verify persistence
	sqliteVecStore.Close()
	sqliteVecStore2, err := store.NewSQLiteVecVectorStore(store.SQLiteVecConfig{
		DBPath:         filepath.Join(tempDir, "vectors.db"),
		CollectionName: "langgraphgo_example",
		Embedder:       embedder,
	})
	if err != nil {
		log.Printf("Failed to reopen store: %v", err)
	} else {
		defer sqliteVecStore2.Close()

		stats2, _ := sqliteVecStore2.GetStats(ctx)
		fmt.Printf("\nReopened store - Documents persist: %d documents\n", stats2.TotalDocuments)
		if stats2.TotalDocuments > 0 {
			fmt.Println("✓ Data successfully persisted across store instances!")
		}
	}

	// Demonstrate update operations
	fmt.Println("\n================================================================================")
	fmt.Println("Update Operations Example")
	fmt.Println("--------------------------------------------------------------------------------")

	updateDoc := rag.Document{
		ID: "doc2",
		Content: "Updated: sqlite-vec features include zero dependencies, pure C implementation, " +
			"support for float/int8/binary vectors, KNN search, and multi-language bindings.",
		Metadata: map[string]any{"source": "sqlite_vec_docs", "category": "features", "updated": true},
	}

	// Delete existing doc2 first (vec0 has limitations on UPDATE)
	err = sqliteVecStore2.Delete(ctx, []string{"doc2"})
	if err != nil {
		log.Printf("Failed to delete document: %v", err)
	} else {
		// Add updated document
		err = sqliteVecStore2.Add(ctx, []rag.Document{updateDoc})
		if err != nil {
			log.Printf("Failed to add updated document: %v", err)
		} else {
			fmt.Println("\nSuccessfully updated document 'doc2'")

			// Verify the update
			queryEmbedding, _ := embedder.EmbedDocument(ctx, "features")
			results, _ := sqliteVecStore2.Search(ctx, queryEmbedding, 5)
			for _, result := range results {
				if result.Document.ID == "doc2" {
					fmt.Printf("Updated content: %s\n", truncate(result.Document.Content, 100))
					break
				}
			}
		}
	}

	fmt.Println("\n=== Example completed successfully! ===")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("  - sqlite-vec provides a lightweight, embedded vector database")
	fmt.Println("  - Uses standard SQLite files for persistent storage")
	fmt.Println("  - Supports KNN vector search with the MATCH operator")
	fmt.Println("  - Integrates seamlessly with LangGraphGo's RAG pipeline")
	fmt.Println("  - Perfect for applications requiring embedded vector search")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
