# RAG with LangChain Embeddings Example

This example demonstrates how to integrate LangChain Go's embeddings with LangGraphGo's RAG system.

## Overview

LangChain Go provides excellent embedding models from various providers (OpenAI, Cohere, HuggingFace, etc.). This example shows how to use them seamlessly with our RAG pipeline through a simple adapter.

## Key Features

- **Direct Integration**: Use LangChain embeddings with minimal wrapper
- **Multiple Providers**: Support for OpenAI, Cohere, and other providers
- **Type Conversion**: Automatic float32 ↔ float64 conversion
- **Complete RAG Pipeline**: End-to-end example with real embeddings
- **Similarity Comparison**: Demonstrate embedding quality

## Architecture

### Adapter Class

The `LangChainEmbedder` adapter in `rag/adapters.go`:

```go
type LangChainEmbedder struct {
    embedder embeddings.Embedder
}
```

**Key Features**:
- Wraps any LangChain embedder
- Implements our `Embedder` interface
- Converts types automatically
- Zero overhead, simple pass-through

## Usage

### Basic Embedding

```go
import (
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/smallnest/langgraphgo/rag"
)

// Create LangChain embedder
lcEmbedder, _ := embeddings.NewEmbedder(openai.New())

// Wrap with adapter
embedder := rag.NewLangChainEmbedder(lcEmbedder)

// Use with our interface
queryEmb, _ := embedder.EmbedDocument(ctx, "What is AI?")
docsEmb, _ := embedder.EmbedDocuments(ctx, texts)
```

### In RAG Pipeline

```go
import (
    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/store"
    "github.com/smallnest/langgraphgo/rag/retriever"
)

// Create embedder
lcEmbedder, _ := embeddings.NewEmbedder(openai.New())
embedder := rag.NewLangChainEmbedder(lcEmbedder)

// Create vector store with LangChain embeddings
vectorStore := store.NewInMemoryVectorStore(embedder)

// Generate embeddings
embeds, _ := embedder.EmbedDocuments(ctx, texts)
vectorStore.AddBatch(ctx, documents, embeds)

// Build RAG pipeline
retriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, 3)
config := rag.DefaultPipelineConfig()
config.Retriever = retriever
config.LLM = llm

pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()
```

## Running the Example

### Prerequisites

For OpenAI embeddings:
```bash
export OPENAI_API_KEY=your_api_key_here
```

For DeepSeek LLM:
```bash
export DEEPSEEK_API_KEY=your_api_key_here
```

### Run

```bash
cd examples/rag_with_embeddings
go run main.go
```

## Examples Included

### 1. OpenAI Embeddings
Test OpenAI's text-embedding-ada-002 model:
```go
openaiEmbedder, _ := embeddings.NewEmbedder(openai.New())
embedder := rag.NewLangChainEmbedder(openaiEmbedder)

queryEmb, _ := embedder.EmbedDocument(ctx, "What is machine learning?")
// Returns 1536-dimensional embedding
```

### 2. Complete RAG Pipeline
Build a full RAG system with real embeddings:
```go
// Use OpenAI embeddings if available, otherwise mock
vectorStore := store.NewInMemoryVectorStore(embedder)
embeds, _ := embedder.EmbedDocuments(ctx, texts)
vectorStore.AddBatch(ctx, documents, embeds)

// Query with semantic search
result, _ := runnable.Invoke(ctx, rag.RAGState{
    Query: "What is LangGraph?",
})
```

### 3. Similarity Comparison
Compare embeddings to understand semantic similarity:
```go
testTexts := []string{
    "Machine learning and AI",
    "Deep learning neural networks",
    "The weather is sunny",
}

embeds, _ := embedder.EmbedDocuments(ctx, testTexts)
similarity := cosineSimilarity(embeds[0], embeds[1])
// High similarity for related texts, low for unrelated
```

## Supported Embedding Providers

The adapter works with all LangChain embedding providers:

### OpenAI
```go
import "github.com/tmc/langchaingo/llms/openai"

lcEmbedder, _ := embeddings.NewEmbedder(openai.New())
embedder := rag.NewLangChainEmbedder(lcEmbedder)
```

**Models**:
- `text-embedding-ada-002` (1536 dimensions) - Default
- `text-embedding-3-small` (1536 dimensions)
- `text-embedding-3-large` (3072 dimensions)

### Cohere
```go
import "github.com/tmc/langchaingo/llms/cohere"

lcEmbedder, _ := embeddings.NewEmbedder(cohere.New())
embedder := rag.NewLangChainEmbedder(lcEmbedder)
```

### HuggingFace
```go
import "github.com/tmc/langchaingo/llms/huggingface"

lcEmbedder, _ := embeddings.NewEmbedder(huggingface.New())
embedder := rag.NewLangChainEmbedder(lcEmbedder)
```

### Vertex AI
```go
import "github.com/tmc/langchaingo/llms/vertexai"

lcEmbedder, _ := embeddings.NewEmbedder(vertexai.New())
embedder := rag.NewLangChainEmbedder(lcEmbedder)
```

## Type Conversion

The adapter handles automatic type conversion:

### LangChain → Our Type
```go
// LangChain returns [][]float32
lcEmbeds := [][]float32{{0.1, 0.2, 0.3}}

// Adapter converts to [][]float64 (actually internally uses float32 now in new RAG package)
ourEmbeds := [][]float32{{0.1, 0.2, 0.3}}
```

### Performance
- Conversion is O(n) where n is total embedding values
- Minimal overhead for typical embedding sizes
- No memory allocation optimization needed for most use cases

## Embedding Dimensions

Different models have different dimensions:

| Model                   | Dimension | Provider |
| ----------------------- | --------- | -------- |
| text-embedding-ada-002  | 1536      | OpenAI   |
| text-embedding-3-small  | 1536      | OpenAI   |
| text-embedding-3-large  | 3072      | OpenAI   |
| embed-english-v3.0      | 1024      | Cohere   |
| embed-multilingual-v3.0 | 1024      | Cohere   |

Ensure your vector store is configured for the correct dimension.

## Best Practices

### 1. Choose the Right Model
- **OpenAI ada-002**: Good balance of quality and cost
- **OpenAI 3-large**: Highest quality, higher cost
- **Cohere**: Good for multilingual content

### 2. Batch Processing
```go
// Process documents in batches for efficiency
texts := []string{...} // Many documents
embeds, _ := embedder.EmbedDocuments(ctx, texts)
// LangChain handles batching internally
```

### 3. Caching
```go
// Cache embeddings for frequently used texts
cache := make(map[string][]float32)

func getEmbedding(text string) []float32 {
    if emb, ok := cache[text]; ok {
        return emb
    }
    emb, _ := embedder.EmbedDocument(ctx, text)
    cache[text] = emb
    return emb
}
```

### 4. Error Handling
```go
embeds, err := embedder.EmbedDocuments(ctx, texts)
if err != nil {
    // Handle rate limits, network errors, etc.
    log.Printf("Embedding failed: %v", err)
    // Consider retry logic
}
```

## Comparison: Mock vs Real Embeddings

### Mock Embeddings (Development)
```go
embedder := store.NewMockEmbedder(1536)
```
- ✅ Fast, no API calls
- ✅ Deterministic
- ✅ Free
- ❌ Not semantically meaningful

### Real Embeddings (Production)
```go
lcEmbedder, _ := embeddings.NewEmbedder(openai.New())
embedder := rag.NewLangChainEmbedder(lcEmbedder)
```
- ✅ Semantically meaningful
- ✅ High quality retrieval
- ✅ Production-ready
- ❌ Requires API key and costs money

## Troubleshooting

### API Key Not Set
```
Error: missing API key
```
**Solution**: Set the appropriate environment variable:
```bash
export OPENAI_API_KEY=your_key
```

### Dimension Mismatch
```
Error: embedding dimension mismatch
```
**Solution**: Ensure vector store dimension matches model:
```go
// For OpenAI ada-002
vectorStore := store.NewInMemoryVectorStore(embedder)
// Embedder will return 1536-dimensional vectors
```

### Rate Limits
```
Error: rate limit exceeded
```
**Solution**: Implement batching and retry logic:
```go
// LangChain embedders support batching
lcEmbedder, _ := embeddings.NewEmbedder(
    openai.New(),
    embeddings.WithBatchSize(100),
)
```

## Performance Tips

1. **Batch Documents**: Process multiple documents at once
2. **Cache Results**: Store embeddings for reuse
3. **Use Appropriate Model**: Balance quality vs cost
4. **Monitor Usage**: Track API calls and costs
5. **Implement Retry**: Handle transient failures

## Next Steps

1. Try different embedding providers (Cohere, HuggingFace)
2. Experiment with different models and dimensions
3. Build a production RAG system with real embeddings
4. Implement caching and optimization strategies
5. Compare embedding quality across providers

## See Also

- [LangChain Embeddings Documentation](https://github.com/tmc/langchaingo)
- [OpenAI Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)
- [RAG Documentation](../../docs/RAG/RAG.md)
- [LangChain Integration Example](../rag_with_langchain/)
