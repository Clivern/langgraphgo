# RAG with Chroma VectorStore Example

This example demonstrates how to use **Chroma vector database** with **LangGraphGo's RAG pipeline** through the langchaingo adapter.

## What is Chroma?

Chroma is an open-source embedding database designed to make it easy to build LLM applications. It provides:
- Simple API for storing and querying embeddings
- Multiple distance metrics (cosine, L2, IP)
- Filtering and metadata support
- Persistent storage
- Client-server architecture

## Prerequisites

1. **Chroma Server**: Run Chroma using Docker
   ```bash
   docker run -p 8000:8000 chromadb/chroma
   ```

2. **API Key**: DeepSeek or OpenAI
   ```bash
   export DEEPSEEK_API_KEY="your-api-key"
   # or
   export OPENAI_API_KEY="your-api-key"
   ```

## Running the Example

```bash
cd examples/rag_chroma_example
go run main.go
```

## Features Demonstrated

- Creating a Chroma vector store with custom configuration
- Adding documents with embeddings to Chroma
- Building a RAG pipeline with Chroma as the vector store
- Querying the pipeline with natural language questions
- Similarity search with relevance scores

## Code Highlights

### Creating Chroma Store

```go
chromaStore, err := chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),
    chroma.WithEmbedder(embedder),
    chroma.WithDistanceFunction("cosine"),
    chroma.WithNameSpace("langgraphgo_example"),
)

// Wrap with adapter
vectorStore := prebuilt.NewLangChainVectorStore(chromaStore)
```

### Adding Documents

```go
err = vectorStore.AddDocuments(ctx, chunks, embeddings)
```

### Similarity Search

```go
results, err := vectorStore.SimilaritySearchWithScore(ctx, query, k)
for _, result := range results {
    fmt.Printf("Score: %.4f - %s\n", result.Score, result.Document.PageContent)
}
```

## Expected Output

```
=== RAG with Chroma VectorStore Example ===

Loading and splitting documents...
Split into 6 chunks

Connecting to Chroma vector database...
Adding documents to Chroma...
Documents successfully added to Chroma

Building RAG pipeline...
Pipeline ready!

================================================================================
Query 1: What is Go programming language?
--------------------------------------------------------------------------------

Retrieved 3 documents from Chroma:
  [1] Go is a statically typed, compiled programming language designed at Google...
  [2] Key features of Go include: - Fast compilation times - Built-in concurrency...
  [3] Go is particularly well-suited for: - Web servers and APIs - Cloud and network...

Answer:
Go is a statically typed, compiled programming language designed at Google...
```

## Chroma Configuration Options

```go
chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),     // Chroma server URL
    chroma.WithEmbedder(embedder),                      // Embedding function
    chroma.WithDistanceFunction("cosine"),              // Distance metric: cosine, l2, ip
    chroma.WithNameSpace("my_collection"),              // Collection name
)
```

## Troubleshooting

### Chroma server not running
```
Error: Failed to create Chroma store: connection refused
```
**Solution**: Start Chroma server with Docker:
```bash
docker run -p 8000:8000 chromadb/chroma
```

### Port already in use
```
Error: bind: address already in use
```
**Solution**: Use a different port or stop the conflicting service:
```bash
docker run -p 8001:8000 chromadb/chroma
# Update code: chroma.WithChromaURL("http://localhost:8001")
```

## Next Steps

- Try other vector stores: Weaviate, Pinecone, Qdrant
- Experiment with different distance functions
- Add metadata filtering to searches
- Implement hybrid search (keyword + vector)

## References

- [Chroma Documentation](https://docs.trychroma.com/)
- [LangChain Go Chroma Integration](https://github.com/tmc/langchaingo/tree/main/vectorstores/chroma)
- [LangGraphGo RAG Documentation](../../docs/RAG/RAG.md)
