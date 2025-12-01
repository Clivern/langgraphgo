# RAG with LangChain VectorStores Example

This example demonstrates how to integrate **langchaingo vectorstores** with **LangGraphGo's RAG pipeline**. It shows how to use LangChain's vector store implementations through our adapter layer.

## Features Demonstrated

1. **LangChain VectorStore Integration**: Using langchaingo's vectorstore implementations
2. **Document Loading & Splitting**: Loading documents with LangChain loaders and splitters
3. **Embeddings Generation**: Using LangChain embedders for vector generation
4. **RAG Pipeline**: Building complete RAG workflows with vector retrieval
5. **Multiple VectorStore Backends**: Support for in-memory and external stores (Weaviate, etc.)
6. **Similarity Search**: Both basic search and search with relevance scores

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    RAG Pipeline                              │
│                                                              │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐             │
│  │ Retrieve │───▶│  Rerank  │───▶│ Generate │             │
│  └──────────┘    └──────────┘    └──────────┘             │
│       │                                  │                  │
│       ▼                                  ▼                  │
│  ┌─────────────────────────────────────────────┐           │
│  │     LangChain VectorStore Adapter           │           │
│  │  (wraps langchaingo vectorstores)           │           │
│  └─────────────────────────────────────────────┘           │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────┐           │
│  │   LangChain VectorStore Implementations     │           │
│  │   - In-Memory                                │           │
│  │   - Weaviate                                 │           │
│  │   - Pinecone                                 │           │
│  │   - Chroma                                   │           │
│  │   - Qdrant                                   │           │
│  └─────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## Prerequisites

1. **DeepSeek API Key** (or OpenAI API Key):
   ```bash
   export DEEPSEEK_API_KEY="your-api-key"
   # or
   export OPENAI_API_KEY="your-api-key"
   ```

2. **(Optional) Weaviate Instance** for external vector store:
   ```bash
   # Run Weaviate with Docker
   docker run -d \
     -p 8080:8080 \
     -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true \
     -e PERSISTENCE_DATA_PATH=/var/lib/weaviate \
     semitechnologies/weaviate:latest

   # Set environment variable
   export WEAVIATE_URL="localhost:8080"
   ```

## Running the Example

```bash
cd examples/rag_langchain_vectorstore_example
go run main.go
```

## Code Walkthrough

### 1. Initialize Components

```go
// Create LLM
llm, err := openai.New(
    openai.WithModel("deepseek-v3"),
    openai.WithBaseURL("https://api.deepseek.com"),
)

// Create embedder
embedder, err := embeddings.NewEmbedder(llm)
```

### 2. Load and Split Documents

```go
// Load documents using LangChain loader
textLoader := documentloaders.NewText(textReader)
loader := prebuilt.NewLangChainDocumentLoader(textLoader)

// Split with LangChain splitter
splitter := textsplitter.NewRecursiveCharacter(
    textsplitter.WithChunkSize(200),
    textsplitter.WithChunkOverlap(50),
)

chunks, err := loader.LoadAndSplit(ctx, splitter)
```

### 3. Create Vector Store

```go
// Option 1: In-memory store
inMemStore := prebuilt.NewInMemoryVectorStore(
    prebuilt.NewLangChainEmbedder(embedder),
)

// Option 2: External store (Weaviate)
weaviateStore, err := weaviate.New(
    weaviate.WithScheme("http"),
    weaviate.WithHost(weaviateURL),
    weaviate.WithEmbedder(embedder),
)

// Wrap with adapter
wrappedStore := prebuilt.NewLangChainVectorStore(weaviateStore)
```

### 4. Add Documents to Vector Store

```go
// Generate embeddings
embeddings, err := embedder.EmbedDocuments(ctx, texts)

// Add to store
err = vectorStore.AddDocuments(ctx, chunks, embeddings)
```

### 5. Build RAG Pipeline

```go
// Create retriever
retriever := prebuilt.NewVectorStoreRetriever(vectorStore, 3)

// Configure pipeline
config := prebuilt.DefaultRAGConfig()
config.Retriever = retriever
config.LLM = llm
config.IncludeCitations = true

// Build and compile
pipeline := prebuilt.NewRAGPipeline(config)
pipeline.BuildAdvancedRAG()
runnable, err := pipeline.Compile()
```

### 6. Query the Pipeline

```go
result, err := runnable.Invoke(ctx, prebuilt.RAGState{
    Query: "What is LangGraph?",
})

finalState := result.(prebuilt.RAGState)
fmt.Println(finalState.Answer)
```

## Supported VectorStore Backends

The adapter supports any langchaingo vectorstore implementation:

### Built-in Stores
- **In-Memory**: For testing and development
- **Weaviate**: Open-source vector database
- **Pinecone**: Managed vector database
- **Chroma**: Embedding database
- **Qdrant**: Vector similarity search engine
- **Milvus**: Cloud-native vector database

### Usage Pattern

```go
// 1. Create langchaingo vectorstore
store, err := <vectorstore>.New(
    // vectorstore-specific options
    <vectorstore>.WithEmbedder(embedder),
)

// 2. Wrap with adapter
adaptedStore := prebuilt.NewLangChainVectorStore(store)

// 3. Use in RAG pipeline
retriever := prebuilt.NewVectorStoreRetriever(adaptedStore, topK)
```

## Example Output

```
=== RAG with LangChain VectorStores Example ===

Example 1: In-Memory VectorStore with LangChain Integration
--------------------------------------------------------------------------------
Split into 8 chunks
Documents added to vector store successfully

Example 2: RAG Pipeline with LangChain VectorStore
--------------------------------------------------------------------------------
Pipeline Visualization:
graph TD
    retrieve --> generate
    generate --> format_citations
    format_citations --> __end__

Query 1: What is LangGraph?
Retrieved 3 documents:
  [1] LangGraph is a library for building stateful, multi-actor applications with LLMs. It extends...
  [2] Key features of LangGraph include: - Stateful graph-based workflows - Support for cycles...
  [3] LangGraph supports multiple checkpoint backends including: - PostgreSQL for production...

Answer: LangGraph is a library designed for building stateful, multi-actor applications with 
Large Language Models (LLMs). It extends the LangChain Expression Language by enabling the 
coordination of multiple chains across multiple steps of computation in a cyclic manner...

Citations:
  [1] Unknown
  [2] Unknown
  [3] Unknown
```

## Advanced Features

### Similarity Search with Scores

```go
results, err := vectorStore.SimilaritySearchWithScore(ctx, query, k)
for _, result := range results {
    fmt.Printf("Score: %.4f - %s\n", result.Score, result.Document.PageContent)
}
```

### Custom Retriever

```go
type CustomRetriever struct {
    store VectorStore
    // custom fields
}

func (r *CustomRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]Document, error) {
    // Custom retrieval logic
    return r.store.SimilaritySearch(ctx, query, r.topK)
}
```

## Integration with Other LangChain Components

This example shows how LangGraphGo seamlessly integrates with the langchaingo ecosystem:

- **Document Loaders**: Text, CSV, PDF, HTML, etc.
- **Text Splitters**: Recursive, Token-based, Semantic
- **Embeddings**: OpenAI, Cohere, HuggingFace
- **Vector Stores**: Weaviate, Pinecone, Chroma, Qdrant
- **LLMs**: OpenAI, Anthropic, Cohere, local models

## Next Steps

1. Explore other examples:
   - `rag_with_langchain/` - Basic LangChain integration
   - `rag_example/` - Custom RAG implementation
   - `rag_advanced_example/` - Advanced RAG patterns

2. Try different vector stores:
   - Set up Pinecone, Chroma, or Qdrant
   - Compare performance and features

3. Customize the pipeline:
   - Add reranking
   - Implement hybrid search
   - Add query expansion

## References

- [LangGraphGo Documentation](../../docs/RAG/RAG.md)
- [LangChain Go Documentation](https://github.com/tmc/langchaingo)
- [Weaviate Documentation](https://weaviate.io/developers/weaviate)
