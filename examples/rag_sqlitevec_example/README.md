# RAG with sqlite-vec Example

This example demonstrates how to use [sqlite-vec](https://github.com/asg017/sqlite-vec) as a vector store with LangGraphGo's RAG pipeline.

## What is sqlite-vec?

sqlite-vec is an extremely small vector search SQLite extension written in pure C. It provides:

- **Zero external dependencies** - Pure C implementation
- **Embedded storage** - Uses standard SQLite files
- **Cross-platform** - Runs anywhere SQLite runs (Linux, macOS, Windows, browsers via WASM)
- **Multiple vector types** - Supports float32, int8, and binary vectors
- **KNN search** - Efficient K-nearest neighbors vector search

## Features Demonstrated

This example shows:

1. **Creating a sqlite-vec vector store** with persistent storage
2. **Adding documents** with embeddings and metadata
3. **Building a RAG pipeline** using the vector store
4. **Similarity search** for document retrieval
5. **Metadata filtering** to narrow search results
6. **Persistent storage** verification (data survives store restart)
7. **Update operations** (delete and re-insert pattern)

## Running the Example

```bash
cd examples/rag_sqlitevec_example
go run main.go
```

## Code Overview

### Initialize the Vector Store

```go
store, err := store.NewSQLiteVecVectorStore(store.SQLiteVecConfig{
    DBPath:         "./vectors.db",  // SQLite database file path
    CollectionName: "my_collection", // Collection/table name
    Embedder:       embedder,        // Embedding function
})
```

### Add Documents

```go
documents := []rag.Document{
    {
        ID:      "doc1",
        Content: "Your document content here",
        Metadata: map[string]any{"category": "tech"},
    },
}

err := store.Add(ctx, documents)
```

### Create RAG Pipeline

```go
vectorRetriever := retriever.NewVectorStoreRetriever(store, embedder, 2)

config := rag.DefaultPipelineConfig()
config.Retriever = vectorRetriever
config.LLM = llm

pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()

runnable, _ := pipeline.Compile()
```

### Query the Pipeline

```go
result, err := runnable.Invoke(ctx, map[string]any{
    "query": "Your question here",
})

answer := result["answer"].(string)
documents := result["documents"].([]rag.RAGDocument)
```

## Storage Options

### In-Memory Storage

For temporary data or testing:

```go
store, err := store.NewSQLiteVecVectorStoreSimple("", embedder)
```

### Persistent Storage

For long-term storage:

```go
store, err := store.NewSQLiteVecVectorStoreSimple("./vectors.db", embedder)
```

## Metadata Filtering

Search documents with specific metadata:

```go
queryEmbedding, _ := embedder.EmbedDocument(ctx, "search query")

results, err := store.SearchWithFilter(ctx, queryEmbedding, 10, map[string]any{
    "category": "tech",
})
```

## Advantages of sqlite-vec

1. **No external services** - Everything runs in your process
2. **Simple deployment** - Just copy a SQLite file
3. **ACID transactions** - Full SQLite transaction support
4. **SQL integration** - Combine vector search with relational queries
5. **Small footprint** - The extension is only a few hundred KB

## Use Cases

- **Edge applications** - Run on devices without internet
- **Desktop apps** - Local vector search in GUI applications
- **Serverless** - No need for separate vector database service
- **Development** - Easy local development and testing
- **Multi-tenant** - Separate SQLite files per tenant

## Comparison with Other Vector Stores

| Feature | sqlite-vec | Chroma | Pinecone |
|---------|-----------|--------|----------|
| Embedded | ✅ Yes | ❌ No | ❌ No |
| External Service | ❌ No | ✅ Yes | ✅ Yes |
| SQL Queries | ✅ Yes | ❌ No | ❌ No |
| ACID Transactions | ✅ Yes | ✅ Yes | ✅ Yes |
| Setup Complexity | ⭐ Low | ⭐⭐ Medium | ⭐⭐⭐ High |

## Production Tips

1. **Use real embeddings** - Replace the mock embedder with OpenAI or similar
2. **Tune dimension** - Match your embedding model's output dimension
3. **Batch operations** - Add documents in batches for better performance
4. **Regular backups** - Copy the SQLite file for backups
5. **Index optimization** - Consider partitioning for large datasets

## See Also

- [sqlite-vec GitHub](https://github.com/asg017/sqlite-vec)
- [sqlite-vec Documentation](https://alexgarcia.xyz/sqlite-vec/)
- [RAG Pipeline Documentation](../../rag/README.md)
- [Other Vector Store Examples](../rag_chromem_example/)
