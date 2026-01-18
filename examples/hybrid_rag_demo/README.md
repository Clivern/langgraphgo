# Hybrid RAG Demo - BM25 + VectorDB

This example demonstrates a complete **Hybrid RAG (Retrieval-Augmented Generation)** system that combines:

- **BM25**: Sparse keyword-based retrieval (exact matching)
- **VectorDB**: Dense semantic-based retrieval (similarity matching)
- **Hybrid Search**: Combines both approaches for optimal results

## 🎯 What This Demo Shows

1. **Side-by-side comparison** of BM25, Vector, and Hybrid retrieval
2. **Complete RAG pipeline** with context building and response generation
3. **Weight tuning** to find the optimal balance between retrieval methods
4. **Statistics and insights** from both retrieval systems

## 🚀 Quick Start

### Run the Demo

```bash
cd examples/hybrid_rag_demo
go run main.go
```

Or build and run:

```bash
go build -o hybrid_rag main.go
./hybrid_rag
```

## 📊 Output Example

```
=== Hybrid RAG Demo (BM25 + VectorDB) ===

This demo combines:
  📊 BM25: Sparse keyword-based retrieval
  🔍 Vector: Dense semantic-based retrieval
  🔀 Hybrid: Best of both worlds!

1. Loading sample documents...
   ✓ Loaded 10 documents

...

Query: "LangGraph framework agents"
----------------------------------------------------------------------
📊 BM25 (Keyword):    [doc1] [doc9] [doc2]
🔍 Vector (Semantic): [doc3] [doc1] [doc5]
🔀 Hybrid (Combined): [doc1] [doc9] [doc3]
```

## 🔧 How It Works

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Hybrid RAG System                        │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  User Query ──┬──> 📊 BM25 Retriever ──┐                   │
│               │                         │                   │
│               ├──> 🔍 Vector Retriever ──┼──> 🔀 Merge      │
│               │                         │     Scores       │
│               │                         │                   │
│               └──> 🎯 Weights (40/60) ──┘                   │
│                                         │                   │
│                                         ▼                   │
│  📚 Top-K Relevant Documents ──> 📝 Context Building        │
│                                         │                   │
│                                         ▼                   │
│  🤖 LLM Generation ──> 💬 Final Response                    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Code Structure

```go
// 1. Load documents
docs := loadSampleDocuments()

// 2. Create embedder
embedder := store.NewMockEmbedder(1536)

// 3. Create vector store
vectorStore := store.NewInMemoryVectorStore(embedder)
vectorStore.Add(ctx, docs)

// 4. Create BM25 retriever (sparse)
bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

// 5. Create vector retriever (dense)
vectorRetriever := retriever.NewVectorRetriever(vectorStore, embedder, config)

// 6. Create hybrid retriever
hybridRetriever := retriever.NewHybridRetriever(
    []rag.Retriever{bm25Retriever, vectorRetriever},
    []float64{0.4, 0.6},  // BM25: 40%, Vector: 60%
    rag.RetrievalConfig{K: 5},
)

// 7. Retrieve and generate
results, _ := hybridRetriever.Retrieve(ctx, query)
context := buildContext(results)
response := generateResponse(ctx, llm, query, context)
```

## 📈 When to Use Each Method

### BM25 (Sparse Retrieval)
- ✅ **Best for**: Exact keyword matching, technical terms, product names
- ✅ **Use cases**:
  - Searching for specific technical documentation
  - Product/part number lookup
  - Exact phrase matching
  - Low-resource environments

### Vector (Dense Retrieval)
- ✅ **Best for**: Semantic understanding, concepts, synonyms
- ✅ **Use cases**:
  - Conceptual queries ("How does X work?")
  - Synonym handling ("car" vs "automobile")
  - Multi-lingual search
  - Recommendation systems

### Hybrid (Combined)
- ✅ **Best for**: Combining keyword precision with semantic understanding
- ✅ **Use cases**:
  - Enterprise search
  - Knowledge management
  - Customer support systems
  - RAG applications

## ⚖️ Tuning Weights

The hybrid retriever uses weights to balance between BM25 and Vector results:

```go
// Technical documentation - emphasize BM25
hybridRetriever := retriever.NewHybridRetriever(
    []rag.Retriever{bm25Retriever, vectorRetriever},
    []float64{0.6, 0.4},  // BM25: 60%, Vector: 40%
    config,
)

// General knowledge - emphasize Vector
hybridRetriever := retriever.NewHybridRetriever(
    []rag.Retriever{bm25Retriever, vectorRetriever},
    []float64{0.3, 0.7},  // BM25: 30%, Vector: 70%
    config,
)

// Balanced approach
hybridRetriever := retriever.NewHybridRetriever(
    []rag.Retriever{bm25Retriever, vectorRetriever},
    []float64{0.5, 0.5},  // BM25: 50%, Vector: 50%
    config,
)
```

### Weight Tuning Guidelines

| Scenario | BM25 Weight | Vector Weight |
|----------|-------------|---------------|
| Technical documentation | 0.6-0.7 | 0.3-0.4 |
| Legal/medical text | 0.5 | 0.5 |
| General knowledge | 0.3-0.4 | 0.6-0.7 |
| E-commerce product search | 0.7 | 0.3 |
| Customer support | 0.4-0.5 | 0.5-0.6 |

## 🔍 Comparison Results

The demo runs three types of queries to show the strengths of each method:

1. **Technical Terms** ("LangGraph framework agents")
   - BM25 excels at finding exact keyword matches

2. **Common Concepts** ("machine learning algorithms")
   - Vector retrieval captures semantic relationships

3. **Mixed Queries** ("programming languages comparison")
   - Hybrid combines both approaches for best results

## 📦 Customization

### Using Your Own Documents

```go
func loadMyDocuments() []rag.Document {
    return []rag.Document{
        {
            ID:      "doc1",
            Content: "Your document content here...",
            Metadata: map[string]any{
                "title": "Document Title",
                "source": "file.pdf",
            },
        },
        // ... more documents
    }
}
```

### Using a Real Vector Store

Replace the in-memory store with a persistent one:

```go
// ChromaDB
import "github.com/smallnest/langgraphgo/rag/store"

vectorStore, err := store.NewChromaVectorStore(
    store.ChromaConfig{
        Host: "localhost",
        Port: 8000,
    },
    embedder,
)

// Or use PostgreSQL, Redis, etc.
```

### Using a Real Embedder

Replace the mock embedder with OpenAI:

```go
import "github.com/tmc/langchaingo/embeddings"
import "github.com/tmc/langchaingo/llms/openai"

llm, _ := openai.New()
embedder, _ := embeddings.NewEmbedder(llm)
```

### Using a Real LLM for Generation

```go
import "github.com/tmc/langchaingo/llms/openai"

func generateResponse(ctx context.Context, llm llms.Model, query, context string) (string, error) {
    prompt := fmt.Sprintf(`
Context:
%s

Question: %s

Answer the question based on the context above.
`, context, query)

    return llms.GenerateFromSinglePrompt(ctx, llm, prompt)
}
```

## 🎓 Key Concepts

### Sparse vs Dense Retrieval

**Sparse Retrieval (BM25)**:
- Represents documents as sparse vectors of term frequencies
- Only considers terms that appear in the query
- Efficient for exact keyword matching
- Lower memory footprint

**Dense Retrieval (Vector)**:
- Represents documents as dense embedding vectors
- Captures semantic relationships
- Handles synonyms and related concepts
- Higher dimensional representations

### Hybrid Score Fusion

The hybrid retriever combines scores from both methods:

```
final_score = w1 * bm25_score + w2 * vector_score

Where:
- w1, w2 are weights that sum to 1.0
- bm25_score is normalized BM25 relevance
- vector_score is cosine similarity
```

### Multi-Source Boost

Documents retrieved by multiple methods receive a boost:

```go
if retrievedByBM25 && retrievedByVector {
    finalScore *= 1.1  // 10% boost
}
```

## 📊 Performance Considerations

### Memory Usage

| Component | Memory (10k docs) | Memory (100k docs) |
|-----------|-------------------|---------------------|
| BM25 Index | ~10-50 MB | ~100-500 MB |
| Vector Store (1536d) | ~60 MB | ~600 MB |
| Hybrid (both) | ~70-100 MB | ~700-1100 MB |

### Query Latency

| Method | Latency (10k docs) | Latency (100k docs) |
|--------|-------------------|---------------------|
| BM25 | ~10-50ms | ~50-200ms |
| Vector | ~50-100ms | ~200-500ms |
| Hybrid | ~60-150ms | ~250-700ms |

*Note: Actual performance depends on hardware and implementation*

## 🔗 Related Examples

- [BM25 Demo](../bm25_demo/) - BM25-only retrieval
- [Vector RAG Demo](../rag_demo/) - Vector-only RAG

## 📚 Further Reading

- [BM25 Integration Documentation](../../docs/bm25_integration.md)
- [RAG Pipeline Guide](../../docs/rag_pipeline.md)
- [Hybrid Search Best Practices](../../docs/hybrid_search.md)

## 🤝 Contributing

Contributions are welcome! Areas for improvement:

- Add reranking step after hybrid retrieval
- Implement query expansion
- Add caching for repeated queries
- Support for more vector stores
- Benchmarking suite

## 📄 License

Part of LangGraphGo project. See main LICENSE file for details.

## 🎯 Summary

This hybrid RAG demo shows how combining sparse (BM25) and dense (vector) retrieval provides better results than either method alone. The key is finding the right weight balance for your use case through experimentation and validation.
