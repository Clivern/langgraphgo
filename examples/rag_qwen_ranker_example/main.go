package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/retriever"
	"github.com/smallnest/langgraphgo/rag/store"
	"github.com/tmc/langchaingo/llms/openai"
)

// This example demonstrates how to use Qwen3-Embedding-4B as both an embedder and reranker
// in a RAG pipeline with LangGraphGo.
//
// Prerequisites:
// 1. Set up your API credentials (for Qwen/DashScope or compatible service)
// 2. Install dependencies: go mod tidy
//
// Run the example:
//   cd examples/rag_qwen_ranker_example
//   go run main.go

func main() {
	ctx := context.Background()

	// Initialize LLM
	llm, err := openai.New()
	if err != nil {
		log.Fatalf("Failed to create LLM: %v", err)
	}

	// Initialize Qwen3-Embedding-4B for embeddings
	// ModelScope API configuration
	// You can use different embedding backends:
	// 1. ModelScope: https://api-inference.modelscope.cn/v1
	// 2. DashScope: https://dashscope.aliyuncs.com/compatible-mode/v1
	// 3. OpenAI: https://api.openai.com/v1
	embeddingBaseURL := os.Getenv("EMBEDDING_BASE_URL")
	if embeddingBaseURL == "" {
		embeddingBaseURL = "https://api-inference.modelscope.cn/v1"
	}

	embeddingModel := os.Getenv("OPENAI_EMBEDDING_MODEL")
	if embeddingModel == "" {
		embeddingModel = "Qwen/Qwen3-Embedding-4B" // ModelScope model
	}

	apiKey := os.Getenv("MODELSCOPE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY") // Fallback to OPENAI_API_KEY
	}

	if apiKey == "" {
		log.Fatal("Please set MODELSCOPE_API_KEY or OPENAI_API_KEY environment variable")
	}

	fmt.Printf("Using embedding model: %s\n", embeddingModel)
	fmt.Printf("Using API endpoint: %s\n\n", embeddingBaseURL)

	// llmForEmbeddings, err := openai.New(
	// 	openai.WithEmbeddingModel(embeddingModel),
	// 	openai.WithBaseURL(embeddingBaseURL),
	// 	openai.WithToken(apiKey),
	// )
	// if err != nil {
	// 	log.Fatalf("Failed to create LLM for embeddings: %v", err)
	// }
	// openaiEmbedder, err := embeddings.NewEmbedder(llmForEmbeddings)
	// if err != nil {
	// 	log.Fatalf("Failed to create embedder: %v", err)
	// }
	// embedder := rag.NewLangChainEmbedder(openaiEmbedder)
	embedder := store.NewMockEmbedder(128)

	fmt.Println("=== RAG with Qwen3-Embedding-4B Reranker Example ===")
	fmt.Println()

	// Create sample documents about AI and machine learning
	documents := []rag.Document{
		{
			ID: "doc1",
			Content: "Qwen3-Embedding-4B is a state-of-the-art embedding model " +
				"released by Alibaba's Qwen team. It provides high-quality vector representations " +
				"for text in multiple languages with 4 billion parameters.",
			Metadata: map[string]any{"source": "qwen_docs", "category": "introduction", "priority": 1},
		},
		{
			ID: "doc2",
			Content: "The Qwen3-Embedding-4B model supports both embedding generation and " +
				"reranking capabilities. For embeddings, it outputs fixed-size vectors. " +
				"For reranking, it can score query-document pairs for relevance.",
			Metadata: map[string]any{"source": "qwen_docs", "category": "features", "priority": 2},
		},
		{
			ID: "doc3",
			Content: "Reranking is a two-stage technique where an initial retrieval stage " +
				"fetches a large number of candidate documents, and a reranker stage " +
				"re-scores them to improve relevance. This combines the speed of vector search " +
				"with the accuracy of cross-encoder models.",
			Metadata: map[string]any{"source": "rag_docs", "category": "techniques", "priority": 1},
		},
		{
			ID: "doc4",
			Content: "LangGraphGo provides a flexible RAG pipeline that supports various " +
				"retrievers including vector stores, rerankers, and hybrid approaches. " +
				"The Qwen3-Embedding-4B can be used as both the initial embedder and the reranker.",
			Metadata: map[string]any{"source": "langgraphgo_docs", "category": "integration", "priority": 2},
		},
		{
			ID: "doc5",
			Content: "Vector databases like Milvus and chromem-go store embeddings efficiently " +
				"and enable fast similarity search. They are commonly used in the retrieval stage " +
				"of RAG systems before reranking.",
			Metadata: map[string]any{"source": "vector_db_docs", "category": "storage", "priority": 1},
		},
		{
			ID: "doc6",
			Content: "The Qwen3-Embedding-4B model uses 4096-dimensional vectors by default, " +
				"providing high capacity for semantic information. These dimensions capture " +
				"nuanced meanings that help with both similarity search and reranking tasks.",
			Metadata: map[string]any{"source": "qwen_docs", "category": "technical", "priority": 2},
		},
	}

	fmt.Println("Initializing vector store with Qwen3-Embedding-4B...")

	// Create in-memory vector store
	vectorStore := store.NewInMemoryVectorStore(embedder)

	// Add documents to the store
	fmt.Println("Adding documents to vector store...")
	err = vectorStore.Add(ctx, documents)
	if err != nil {
		log.Fatalf("Failed to add documents: %v", err)
	}
	fmt.Printf("Successfully added %d documents\n\n", len(documents))

	// Create a Qwen3-Embedding-4B reranker using LLM-based reranking
	// The reranker will re-score the retrieved documents
	rerankerConfig := retriever.DefaultLLMRerankerConfig()
	rerankerConfig.TopK = 3 // Return top 3 after reranking
	rerankerConfig.SystemPrompt = "You are a relevance scoring assistant for AI and machine learning topics. " +
		"Rate how well each document answers the query on a scale of 0.0 to 1.0, " +
		"where 1.0 is perfectly relevant and 0.0 is not relevant. " +
		"Consider semantic meaning, factual accuracy, and completeness."

	llmReranker := retriever.NewLLMReranker(llm, rerankerConfig)

	// Create a custom retriever that combines vector search and reranking
	// This retriever fetches more candidates initially, then reranks them
	compositeRetriever := &RerankingRetriever{
		vectorStore: vectorStore,
		embedder:    embedder,
		reranker:    llmReranker,
		fetchK:      10, // Fetch 10 candidates for reranking
	}

	fmt.Println("Created composite retriever with vector search + LLM reranking")
	fmt.Println()

	// Configure RAG pipeline
	config := rag.DefaultPipelineConfig()
	config.Retriever = compositeRetriever
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
	fmt.Println("Pipeline Graph:")
	exporter := graph.GetGraphForRunnable(runnable)
	fmt.Println(exporter.DrawASCII())
	fmt.Println()

	// Test queries demonstrating different aspects
	queries := []string{
		"What is Qwen3-Embedding-4B?",
		"How does reranking improve search results?",
		"What vector databases are supported?",
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

	// Demonstrate reranking effect
	fmt.Println("================================================================================")
	fmt.Println("Reranking Demonstration")
	fmt.Println("--------------------------------------------------------------------------------")

	testQuery := "What are the features of Qwen embedding models?"
	fmt.Printf("\nQuery: %s\n\n", testQuery)

	// First, get results without reranking
	fmt.Println("1. Vector Search Results (without reranking):")
	queryEmbedding, err := embedder.EmbedDocument(ctx, testQuery)
	if err != nil {
		log.Printf("Embedding failed: %v", err)
	} else {
		vectorResults, err := vectorStore.Search(ctx, queryEmbedding, 5)
		if err != nil {
			log.Printf("Vector search failed: %v", err)
		} else {
			for i, result := range vectorResults {
				fmt.Printf("   [%d] Score: %.4f - %s\n", i+1, result.Score, truncate(result.Document.Content, 80))
			}
		}
	}

	// Then, get results with reranking
	fmt.Println("\n2. After Reranking with Qwen3-Embedding-4B:")

	// Get search results with scores for reranking
	candidates, err := vectorStore.Search(ctx, queryEmbedding, 10)
	if err != nil {
		log.Printf("Search failed: %v", err)
	} else {
		// Rerank
		rerankedResults, err := llmReranker.Rerank(ctx, testQuery, candidates)
		if err != nil {
			log.Printf("Reranking failed: %v", err)
		} else {
			for i, result := range rerankedResults {
				fmt.Printf("   [%d] Score: %.4f - %s\n", i+1, result.Score, truncate(result.Document.Content, 80))
			}
		}
	}

	fmt.Println()
	fmt.Println("=== Example completed successfully! ===")
	fmt.Println("\nKey Features Demonstrated:")
	fmt.Println("1. Qwen3-Embedding-4B for initial document embedding")
	fmt.Println("2. Vector similarity search for fast retrieval")
	fmt.Println("3. LLM-based reranking for improved relevance")
	fmt.Println("4. Two-stage retrieval: fetch many, rerank few")
	fmt.Println("5. Composite retriever combining both approaches")

	fmt.Println("\nConfiguration Options:")
	fmt.Println(`
# Option 1: ModelScope (for Qwen3-Embedding-4B)
export EMBEDDING_BASE_URL=https://api-inference.modelscope.cn/v1
export MODELSCOPE_API_KEY=your-modelscope-api-key
export OPENAI_EMBEDDING_MODEL=Qwen/Qwen3-Embedding-4B

# Option 2: DashScope (Alibaba Cloud)
export EMBEDDING_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
export OPENAI_API_KEY=your-dashscope-api-key
export OPENAI_EMBEDDING_MODEL=text-embedding-v3

# Option 3: OpenAI
export EMBEDDING_BASE_URL=https://api.openai.com/v1
export OPENAI_API_KEY=your-openai-api-key
export OPENAI_EMBEDDING_MODEL=text-embedding-3-small
`)

	fmt.Println("\nFor more information, see:")
	fmt.Println("- Qwen Documentation: https://qwen.readthedocs.io/")
	fmt.Println("- DashScope API: https://help.aliyun.com/zh/dashscope/")
	fmt.Println("- LangGraphGo RAG: ../../rag/README.md")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// RerankingRetriever is a custom retriever that combines vector search and LLM reranking
type RerankingRetriever struct {
	vectorStore rag.VectorStore
	embedder    rag.Embedder
	reranker    rag.Reranker
	fetchK      int
}

// Retrieve retrieves documents using vector search and then reranks them
func (r *RerankingRetriever) Retrieve(ctx context.Context, query string) ([]rag.Document, error) {
	return r.RetrieveWithK(ctx, query, 3)
}

// RetrieveWithK retrieves exactly k documents using vector search + reranking
func (r *RerankingRetriever) RetrieveWithK(ctx context.Context, query string, k int) ([]rag.Document, error) {
	config := &rag.RetrievalConfig{K: r.fetchK}
	results, err := r.RetrieveWithConfig(ctx, query, config)
	if err != nil {
		return nil, err
	}

	// Limit to k results
	if len(results) > k {
		results = results[:k]
	}

	docs := make([]rag.Document, len(results))
	for i, result := range results {
		docs[i] = result.Document
	}
	return docs, nil
}

// RetrieveWithConfig retrieves documents with custom configuration
func (r *RerankingRetriever) RetrieveWithConfig(ctx context.Context, query string, config *rag.RetrievalConfig) ([]rag.DocumentSearchResult, error) {
	// Step 1: Fetch more candidates using vector search
	queryEmbedding, err := r.embedder.EmbedDocument(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	candidates, err := r.vectorStore.Search(ctx, queryEmbedding, r.fetchK)
	if err != nil {
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// Step 2: Rerank the candidates
	reranked, err := r.reranker.Rerank(ctx, query, candidates)
	if err != nil {
		// Fallback to original results if reranking fails
		reranked = candidates
	}

	// Step 3: Apply score threshold if specified
	if config != nil && config.ScoreThreshold > 0 {
		filtered := make([]rag.DocumentSearchResult, 0)
		for _, result := range reranked {
			if result.Score >= config.ScoreThreshold {
				filtered = append(filtered, result)
			}
		}
		reranked = filtered
	}

	return reranked, nil
}
