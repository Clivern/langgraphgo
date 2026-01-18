package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/retriever"
	"github.com/smallnest/langgraphgo/rag/store"
)

func main() {
	fmt.Println("=== Hybrid RAG Demo (BM25 + VectorDB) ===\n")
	fmt.Println("This demo combines:")
	fmt.Println("  📊 BM25: Sparse keyword-based retrieval")
	fmt.Println("  🔍 Vector: Dense semantic-based retrieval")
	fmt.Println("  🔀 Hybrid: Best of both worlds!\n")

	// Run the demo
	runHybridRAGDemo()
}

func runHybridRAGDemo() {
	ctx := context.Background()

	// 1. Load sample documents
	fmt.Println("1. Loading sample documents...")
	docs := loadSampleDocuments()
	fmt.Printf("   ✓ Loaded %d documents\n\n", len(docs))

	// 2. Create embedder (using mock embedder for demo)
	fmt.Println("2. Initializing embedder...")
	embedder := store.NewMockEmbedder(1536) // OpenAI embedding dimension
	fmt.Printf("   ✓ Embedder initialized (dimension: %d)\n\n", embedder.GetDimension())

	// 3. Create and populate vector store
	fmt.Println("3. Creating vector store...")
	vectorStore := store.NewInMemoryVectorStore(embedder)

	// Add documents to vector store
	err := addDocumentsToVectorStore(ctx, vectorStore, docs)
	if err != nil {
		log.Fatalf("Failed to add documents to vector store: %v", err)
	}
	fmt.Printf("   ✓ Vector store populated with %d documents\n\n", len(docs))

	// 4. Create BM25 retriever (sparse retrieval)
	fmt.Println("4. Creating BM25 retriever (sparse)...")
	bm25Config := retriever.DefaultBM25Config()
	bm25Config.K = 10 // Get more candidates for hybrid
	bm25Retriever, err := retriever.NewBM25Retriever(docs, bm25Config)
	if err != nil {
		log.Fatalf("Failed to create BM25 retriever: %v", err)
	}
	fmt.Println("   ✓ BM25 retriever created")
	fmt.Println("   → Best for: exact keyword matching, technical terms\n")

	// 5. Create vector retriever (dense retrieval)
	fmt.Println("5. Creating vector retriever (dense)...")
	vectorConfig := rag.RetrievalConfig{
		K:              10, // Get more candidates for hybrid
		ScoreThreshold: 0.0,
	}
	vectorRetriever := retriever.NewVectorRetriever(vectorStore, embedder, vectorConfig)
	fmt.Println("   ✓ Vector retriever created")
	fmt.Println("   → Best for: semantic understanding, concepts, meaning\n")

	// 6. Create hybrid retriever
	fmt.Println("6. Creating hybrid retriever...")
	hybridConfig := rag.RetrievalConfig{
		K:              5, // Final top-k results
		ScoreThreshold: 0.0,
	}

	// You can adjust the weights based on your use case:
	// - More BM25 weight (0.6, 0.4): Better for technical queries, exact terms
	// - More Vector weight (0.4, 0.6): Better for conceptual queries, synonyms
	// - Equal weights (0.5, 0.5): Balanced approach
	hybridRetriever := retriever.NewHybridRetriever(
		[]rag.Retriever{bm25Retriever, vectorRetriever},
		[]float64{0.4, 0.6}, // BM25: 40%, Vector: 60%
		hybridConfig,
	)
	fmt.Println("   ✓ Hybrid retriever created")
	fmt.Println("   → Weights: BM25 40%, Vector 60%")
	fmt.Println("   → Best for: combining keyword precision with semantic understanding\n")

	// 7. Comparison Test - Side by side comparison
	fmt.Println("7. Comparison Test")
	fmt.Println("=====================================\n")

	queries := []struct {
		query string
		desc  string
	}{
		{"LangGraph framework agents", "Technical terms (BM25 should excel)"},
		{"machine learning algorithms", "Common concepts (Vector should excel)"},
		{"programming languages comparison", "Mixed terms (Hybrid should excel)"},
	}

	for _, q := range queries {
		fmt.Printf("Query: \"%s\"\n", q.query)
		fmt.Printf("Type: %s\n", q.desc)
		fmt.Println(strings.Repeat("-", 70))

		// BM25 Results
		fmt.Print("📊 BM25 (Keyword):    ")
		bm25Results, _ := bm25Retriever.Retrieve(ctx, q.query)
		printResultIDs(bm25Results)

		// Vector Results
		fmt.Print("🔍 Vector (Semantic): ")
		vectorResults, _ := vectorRetriever.Retrieve(ctx, q.query)
		printResultIDs(vectorResults)

		// Hybrid Results
		fmt.Print("🔀 Hybrid (Combined): ")
		hybridResults, _ := hybridRetriever.Retrieve(ctx, q.query)
		printResultIDs(hybridResults)

		// Show hybrid results with scores
		fmt.Println("\n   Hybrid details (top 3):")
		retrievalConfig := &rag.RetrievalConfig{K: 3, IncludeScores: true}
		detailedResults, _ := hybridRetriever.RetrieveWithConfig(ctx, q.query, retrievalConfig)
		for i, r := range detailedResults {
			fmt.Printf("   %d. [%s] Score: %.3f - %.50s...\n",
				i+1, r.Document.ID, r.Score, r.Document.Content)
		}
		fmt.Println()
	}

	// 8. RAG Pipeline Example
	fmt.Println("8. RAG Pipeline Example")
	fmt.Println("=====================================\n")

	query := "What is LangGraph and how does it work with agents?"

	fmt.Printf("User Query: \"%s\"\n\n", query)

	// Step 1: Retrieve relevant documents using hybrid search
	fmt.Println("📚 Step 1: Retrieving relevant documents...")
	retrievedDocs, err := hybridRetriever.Retrieve(ctx, query)
	if err != nil {
		log.Fatalf("Failed to retrieve: %v", err)
	}

	fmt.Printf("   Retrieved %d relevant documents\n", len(retrievedDocs))
	for i, doc := range retrievedDocs {
		fmt.Printf("   %d. [%s] %.60s...\n", i+1, doc.ID, doc.Content)
	}

	// Step 2: Build context from retrieved documents
	fmt.Println("\n📝 Step 2: Building context from retrieved documents...")
	contextStr := buildContext(retrievedDocs)
	fmt.Printf("   Context length: %d characters\n", len(contextStr))

	// Step 3: Generate response (simulated)
	fmt.Println("\n🤖 Step 3: Generating response...")
	fmt.Println("\n💬 Response:")
	fmt.Println("─────────────────────────────────────────────────────────────")

	response := generateMockResponse(query, retrievedDocs)
	fmt.Println(response)

	fmt.Println("─────────────────────────────────────────────────────────────")

	// 9. Weight Comparison
	fmt.Println("\n9. Weight Comparison")
	fmt.Println("=====================================\n")

	fmt.Println("Testing different weight combinations for query:")
	fmt.Printf("\"%s\"\n\n", queries[2].query)

	weights := []struct {
		name    string
		weights []float64
	}{
		{"BM25 Only", []float64{1.0, 0.0}},
		{"Vector Only", []float64{0.0, 1.0}},
		{"Balanced", []float64{0.5, 0.5}},
		{"BM25 Leaning", []float64{0.6, 0.4}},
		{"Vector Leaning", []float64{0.4, 0.6}},
	}

	for _, w := range weights {
		hr := retriever.NewHybridRetriever(
			[]rag.Retriever{bm25Retriever, vectorRetriever},
			w.weights,
			rag.RetrievalConfig{K: 3},
		)

		results, _ := hr.Retrieve(ctx, queries[2].query)
		fmt.Printf("%-15s: ", w.name)
		printResultIDs(results)
	}

	// 10. Statistics and Insights
	fmt.Println("\n10. Statistics and Insights")
	fmt.Println("=====================================\n")

	// BM25 Statistics
	bm25Stats := bm25Retriever.GetStats()
	fmt.Println("📊 BM25 Index Statistics:")
	fmt.Printf("   Total documents: %v\n", bm25Stats["num_documents"])
	fmt.Printf("   Unique terms: %v\n", bm25Stats["num_unique_terms"])
	fmt.Printf("   Average doc length: %.1f tokens\n", bm25Stats["avg_doc_length"])
	fmt.Printf("   K1 parameter: %.2f\n", bm25Stats["k1"])
	fmt.Printf("   B parameter: %.2f\n", bm25Stats["b"])

	// Vector Store Statistics
	vectorStats, _ := vectorStore.GetStats(ctx)
	fmt.Println("\n🔍 Vector Store Statistics:")
	fmt.Printf("   Total documents: %v\n", vectorStats.TotalDocuments)
	fmt.Printf("   Total vectors: %v\n", vectorStats.TotalVectors)
	fmt.Printf("   Dimension: %v\n", vectorStats.Dimension)

	fmt.Println("\n" + strings.Repeat("=", 70))
	fmt.Println("✅ Hybrid RAG Demo Complete!")
	fmt.Println("=")
	fmt.Println("\nKey Takeaways:")
	fmt.Println("  • BM25 excels at exact keyword matching")
	fmt.Println("  • Vector retrieval captures semantic meaning")
	fmt.Println("  • Hybrid search combines both for optimal results")
	fmt.Println("  • Weights can be tuned based on your use case")
	fmt.Println("\nNext Steps:")
	fmt.Println("  • Adjust weights based on your data and queries")
	fmt.Println("  • Add reranking for even better results")
	fmt.Println("  • Integrate with a real LLM for generation")
	fmt.Println("  • Use a persistent vector store for production")
}

func addDocumentsToVectorStore(ctx context.Context, vs *store.InMemoryVectorStore, docs []rag.Document) error {
	for _, doc := range docs {
		err := vs.Add(ctx, []rag.Document{doc})
		if err != nil {
			return fmt.Errorf("failed to add document %s: %w", doc.ID, err)
		}
	}
	return nil
}

func printResultIDs(docs []rag.Document) {
	if len(docs) == 0 {
		fmt.Println("[No results]")
		return
	}

	for i, doc := range docs {
		if i >= 3 {
			fmt.Printf("... (+%d more)", len(docs)-3)
			break
		}
		fmt.Printf("[%s] ", doc.ID)
	}
	fmt.Println()
}

func loadSampleDocuments() []rag.Document {
	return []rag.Document{
		{
			ID:      "doc1",
			Content: "LangGraph is a framework for building stateful, multi-actor applications with LLMs. It allows you to create agent workflows where LLMs can call tools, maintain state across interactions, and coordinate with other agents.",
			Metadata: map[string]any{
				"title":    "LangGraph Overview",
				"category": "framework",
				"tags":     []string{"LLM", "agents", "workflow"},
			},
		},
		{
			ID:      "doc2",
			Content: "BM25 is a ranking function used by search engines to estimate the relevance of documents to a given search query. It's based on the probabilistic retrieval framework and improves upon TF-IDF by incorporating document length normalization.",
			Metadata: map[string]any{
				"title":    "BM25 Overview",
				"category": "algorithm",
				"tags":     []string{"search", "ranking", "information-retrieval"},
			},
		},
		{
			ID:      "doc3",
			Content: "Machine learning is a subset of artificial intelligence that enables systems to learn and improve from experience without being explicitly programmed. Common algorithms include neural networks, decision trees, and support vector machines.",
			Metadata: map[string]any{
				"title":    "Machine Learning Introduction",
				"category": "AI",
				"tags":     []string{"ML", "AI", "algorithms"},
			},
		},
		{
			ID:      "doc4",
			Content: "Vector databases store data as high-dimensional vectors, enabling semantic search and similarity matching. They're essential for RAG (Retrieval-Augmented Generation) systems and power AI applications like recommendation systems and image search.",
			Metadata: map[string]any{
				"title":    "Vector Databases",
				"category": "database",
				"tags":     []string{"vectors", "embeddings", "search"},
			},
		},
		{
			ID:      "doc5",
			Content: "RAG (Retrieval-Augmented Generation) combines retrieval systems with generative AI. It retrieves relevant documents and uses them as context for LLM generation, producing more accurate and grounded responses.",
			Metadata: map[string]any{
				"title":    "RAG Explained",
				"category": "AI",
				"tags":     []string{"RAG", "LLM", "generation"},
			},
		},
		{
			ID:      "doc6",
			Content: "Go is a statically typed, compiled programming language designed at Google. It's known for simplicity, concurrency support, and efficient performance. Go is commonly used for cloud services, APIs, and distributed systems.",
			Metadata: map[string]any{
				"title":    "Go Programming Language",
				"category": "programming",
				"tags":     []string{"Go", "golang", "concurrency"},
			},
		},
		{
			ID:      "doc7",
			Content: "Python is a high-level, interpreted programming language known for its simplicity and readability. It's widely used in data science, machine learning, web development, and automation.",
			Metadata: map[string]any{
				"title":    "Python Programming",
				"category": "programming",
				"tags":     []string{"Python", "data-science", "ML"},
			},
		},
		{
			ID:      "doc8",
			Content: "Hybrid search combines multiple retrieval strategies, typically sparse (keyword-based) and dense (vector-based) methods. This approach provides both exact keyword matching and semantic understanding, improving overall search quality.",
			Metadata: map[string]any{
				"title":    "Hybrid Search",
				"category": "search",
				"tags":     []string{"hybrid", "RAG", "search-algorithms"},
			},
		},
		{
			ID:      "doc9",
			Content: "Agents are autonomous systems that use LLMs to perform tasks. They can reason about their environment, use tools, interact with other agents, and maintain memory across conversations to achieve complex goals.",
			Metadata: map[string]any{
				"title":    "AI Agents",
				"category": "AI",
				"tags":     []string{"agents", "LLM", "autonomy"},
			},
		},
		{
			ID:      "doc10",
			Content: "Embeddings are vector representations of text that capture semantic meaning. Similar concepts have similar embeddings, allowing machines to understand relationships between words and documents mathematically.",
			Metadata: map[string]any{
				"title":    "Text Embeddings",
				"category": "AI",
				"tags":     []string{"embeddings", "vectors", "NLP"},
			},
		},
	}
}

func buildContext(docs []rag.Document) string {
	var context strings.Builder
	context.WriteString("Relevant Context:\n")
	context.WriteString("═════════════════\n\n")

	for i, doc := range docs {
		context.WriteString(fmt.Sprintf("[%d] %s\n", i+1, doc.Content))
		context.WriteString(fmt.Sprintf("    Source: %s\n\n", doc.ID))
	}

	return context.String()
}

func generateMockResponse(query string, docs []rag.Document) string {
	// In a real application, you would call an LLM here
	// For demo purposes, we'll return a structured response

	var response strings.Builder

	response.WriteString(fmt.Sprintf("Based on the retrieved documents, here's what I found about \"%s\":\n\n", query))

	// Find most relevant doc
	var bestDoc rag.Document
	var bestScore float64 = 0

	for _, doc := range docs {
		// Simple relevance scoring
		score := calculateRelevance(query, doc.Content)
		if score > bestScore {
			bestScore = score
			bestDoc = doc
		}
	}

	if bestDoc.ID != "" {
		response.WriteString(fmt.Sprintf("According to \"%s\" (doc %s):\n", bestDoc.Metadata["title"], bestDoc.ID))
		response.WriteString(bestDoc.Content)
		response.WriteString("\n\n")

		// Add additional context from other docs
		if len(docs) > 1 {
			response.WriteString("Additional relevant information was also found in other documents,")
			response.WriteString(" providing more context about this topic.")
		}
	}

	response.WriteString(fmt.Sprintf("\n\n📊 Note: This is a simulated response. In production, you would connect this to an actual LLM (like GPT-4)"))
	response.WriteString(fmt.Sprintf("\n   using the retrieved %d documents as context.", len(docs)))

	return response.String()
}

func calculateRelevance(query, content string) float64 {
	// Simple relevance calculation
	queryTerms := strings.Fields(strings.ToLower(query))
	contentLower := strings.ToLower(content)

	score := 0.0
	for _, term := range queryTerms {
		if strings.Contains(contentLower, term) {
			score += 1.0
		}
	}

	return score
}

// Color codes for terminal output (optional enhancement)
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// Helper function to print colored output (optional)
func printColor(color, text string) {
	if os.Getenv("NO_COLOR") != "" {
		fmt.Print(text)
	} else {
		fmt.Print(color + text + ColorReset)
	}
}
