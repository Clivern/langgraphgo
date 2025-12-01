# LangChain VectorStores Integration

## Overview

This document summarizes the integration of `github.com/tmc/langchaingo/vectorstores` into LangGraphGo, providing seamless access to multiple vector database backends for RAG applications.

## What Was Added

### 1. LangChainVectorStore Adapter (`prebuilt/rag_langchain_adapter.go`)

A new adapter that wraps any `langchaingo` vectorstore implementation to work with LangGraphGo's RAG pipeline:

```go
type LangChainVectorStore struct {
    store vectorstores.VectorStore
}

func NewLangChainVectorStore(store vectorstores.VectorStore) *LangChainVectorStore
```

**Methods**:
- `AddDocuments(ctx, documents, embeddings)` - Add documents to the vector store
- `SimilaritySearch(ctx, query, k)` - Search for similar documents
- `SimilaritySearchWithScore(ctx, query, k)` - Search with relevance scores

### 2. Test Suite (`prebuilt/rag_langchain_vectorstore_test.go`)

Comprehensive tests ensuring the adapter works correctly:
- Document addition tests
- Similarity search tests
- Score-based search tests
- Integration tests

### 3. Examples

#### Example 1: General VectorStore Integration
**Location**: `examples/rag_langchain_vectorstore_example/`

Demonstrates:
- Using in-memory vector store
- Optional Weaviate integration
- Complete RAG pipeline with LangChain components
- Similarity search with scores

**Files**:
- `main.go` - Complete working example
- `README.md` - English documentation
- `README_CN.md` - Chinese documentation

#### Example 2: Chroma Integration
**Location**: `examples/rag_chroma_example/`

Demonstrates:
- Chroma vector database setup
- Document indexing with Chroma
- RAG pipeline with Chroma backend
- Production-ready configuration

**Files**:
- `main.go` - Chroma-specific example
- `README.md` - English documentation with setup guide
- `README_CN.md` - Chinese documentation

### 4. Documentation Updates

Updated `docs/RAG/RAG.md` with:
- Comprehensive LangChain integration section
- Adapter usage examples
- Supported vector store backends
- Setup guides for Chroma, Weaviate, Pinecone
- Complete integration example
- Benefits and best practices

## Supported Vector Stores

The adapter supports **any** langchaingo vectorstore implementation, including:

| Vector Store | Type                 | Best For                            |
| ------------ | -------------------- | ----------------------------------- |
| **Chroma**   | Open-source          | Development, small-medium scale     |
| **Weaviate** | Open-source/Cloud    | Production, scalability             |
| **Pinecone** | Managed Service      | Ease of use, managed infrastructure |
| **Qdrant**   | Open-source/Cloud    | High performance, filtering         |
| **Milvus**   | Open-source/Cloud    | Large scale, distributed            |
| **PGVector** | PostgreSQL Extension | Existing PostgreSQL infrastructure  |

## Usage Pattern

### Basic Usage

```go
// 1. Create langchaingo vector store
chromaStore, err := chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),
    chroma.WithEmbedder(embedder),
)

// 2. Wrap with adapter
vectorStore := prebuilt.NewLangChainVectorStore(chromaStore)

// 3. Use in RAG pipeline
retriever := prebuilt.NewVectorStoreRetriever(vectorStore, 3)

config := prebuilt.DefaultRAGConfig()
config.Retriever = retriever
config.LLM = llm

pipeline := prebuilt.NewRAGPipeline(config)
pipeline.BuildBasicRAG()
```

### Complete Integration

```go
// All LangChain components together
loader := prebuilt.NewLangChainDocumentLoader(documentloaders.NewText(reader))
splitter := prebuilt.NewLangChainTextSplitter(textsplitter.NewRecursiveCharacter(...))
embedder := prebuilt.NewLangChainEmbedder(embeddings.NewEmbedder(llm))
vectorStore := prebuilt.NewLangChainVectorStore(chroma.New(...))

// Build RAG pipeline
chunks, _ := loader.LoadAndSplit(ctx, splitter)
embeddings, _ := embedder.EmbedDocuments(ctx, texts)
vectorStore.AddDocuments(ctx, chunks, embeddings)

retriever := prebuilt.NewVectorStoreRetriever(vectorStore, 3)
// ... continue with pipeline
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                  LangGraphGo RAG Pipeline                    │
│                                                              │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐             │
│  │ Retrieve │───▶│  Rerank  │───▶│ Generate │             │
│  └──────────┘    └──────────┘    └──────────┘             │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────┐           │
│  │   LangChainVectorStore Adapter              │           │
│  │   (prebuilt.NewLangChainVectorStore)        │           │
│  └─────────────────────────────────────────────┘           │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────┐           │
│  │   langchaingo vectorstores.VectorStore      │           │
│  │   - Chroma                                   │           │
│  │   - Weaviate                                 │           │
│  │   - Pinecone                                 │           │
│  │   - Qdrant                                   │           │
│  │   - Milvus                                   │           │
│  │   - PGVector                                 │           │
│  └─────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## Benefits

1. **Ecosystem Access**: Direct access to the entire langchaingo vectorstore ecosystem
2. **Production Ready**: Use battle-tested vector databases in production
3. **Flexibility**: Easy to switch between different vector stores
4. **No Vendor Lock-in**: Standard interface works with any backend
5. **Future Proof**: Automatically compatible with new langchaingo vectorstore implementations

## Testing

Run the tests:

```bash
# Test the adapter
go test ./prebuilt -run TestLangChainVectorStore

# Run all RAG tests
go test ./prebuilt -run RAG
```

## Examples

### Run the general example (in-memory):
```bash
cd examples/rag_langchain_vectorstore_example
export DEEPSEEK_API_KEY="your-key"
go run main.go
```

### Run the Chroma example:
```bash
# Start Chroma
docker run -p 8000:8000 chromadb/chroma

# Run example
cd examples/rag_chroma_example
export DEEPSEEK_API_KEY="your-key"
go run main.go
```

## Migration Guide

### From InMemoryVectorStore to Chroma

**Before**:
```go
embedder := prebuilt.NewMockEmbedder(128)
vectorStore := prebuilt.NewInMemoryVectorStore(embedder)
```

**After**:
```go
// Create LangChain embedder
lcEmbedder, _ := embeddings.NewEmbedder(llm)
embedder := prebuilt.NewLangChainEmbedder(lcEmbedder)

// Create Chroma store
chromaStore, _ := chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),
    chroma.WithEmbedder(lcEmbedder),
)
vectorStore := prebuilt.NewLangChainVectorStore(chromaStore)
```

The rest of your RAG pipeline code remains unchanged!

## Next Steps

1. **Try Different Backends**: Experiment with Weaviate, Pinecone, Qdrant
2. **Production Deployment**: Set up managed vector databases
3. **Performance Tuning**: Optimize chunk size, top-k, and distance metrics
4. **Hybrid Search**: Combine vector search with keyword search
5. **Monitoring**: Add metrics and logging for production use

## References

- [LangChain Go Documentation](https://github.com/tmc/langchaingo)
- [LangGraphGo RAG Documentation](../docs/RAG/RAG.md)
- [Chroma Documentation](https://docs.trychroma.com/)
- [Weaviate Documentation](https://weaviate.io/developers/weaviate)
- [Pinecone Documentation](https://docs.pinecone.io/)

## Contributing

To add support for a new vector store:

1. Ensure it implements `vectorstores.VectorStore` from langchaingo
2. Create an example in `examples/rag_<vectorstore>_example/`
3. Add setup instructions to the documentation
4. Submit a PR with tests

The adapter automatically supports any langchaingo vectorstore, so no code changes are needed!
