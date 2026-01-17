# RAG with Milvus VectorStore Example

This example demonstrates how to use [Milvus](https://milvus.io/) as a vector store with LangGraphGo's RAG pipeline.

## What is Milvus?

Milvus is an open-source vector database built to power embedding similarity search and AI applications. It provides:

- **Billion-scale vector indexing** - Handle massive vector datasets
- **Real-time search performance** - Sub-millisecond latency
- **Multiple index types** - HNSW, IVF, Flat, and more
- **Flexible deployment** - Standalone, distributed, or cloud-native
- **Rich features** - Partitions, replicas, scalar filtering

## Features Demonstrated

This example shows:

1. **Creating a Milvus vector store** with custom configuration
2. **Adding documents** with embeddings and metadata
3. **Building a RAG pipeline** using the Milvus vector store
4. **Similarity search** for document retrieval
5. **Direct Milvus API usage** for advanced operations
6. **Multi-language support** through Milvus Go SDK v2

## Prerequisites

### Start Milvus Server

For local development, start Milvus using Docker:

```bash
docker run -d \
  --name milvus-standalone \
  -p 19530:19530 \
  -v milvus:/var/lib/milvus \
  milvusdb/milvus:latest
```

Or use Docker Compose for a more complete setup:

```bash
# Download docker-compose.yml
wget https://github.com/milvus-io/milvus/releases/download/v2.4.0/milvus-standalone-docker-compose.yml -O docker-compose.yml

# Start Milvus
docker-compose up -d
```

### Install Dependencies

```bash
go get github.com/tmc/langchaingo/vectorstores/milvus/v2
```

## Running the Example

```bash
cd examples/rag_milvus_example
go run main.go
```

To connect to a remote Milvus instance:

```bash
MILVUS_ADDRESS=your-milvus-server:19530 go run main.go
```

## Code Overview

### Initialize Milvus Client

```go
milvusConfig := client.Config{
    Address: "localhost:19530",
}

store, err := milvusv2.New(
    ctx,
    milvusConfig,
    milvusv2.WithEmbedder(embedder),
    milvusv2.WithCollectionName("my_documents"),
    milvusv2.WithIndex(entity.NewFlatIndex(entity.COSINE)),
    milvusv2.WithMetricType(entity.COSINE),
)
```

### Configuration Options

```go
// Collection options
milvusv2.WithCollectionName("name")     // Collection name
milvusv2.WithPartitionName("partition") // Partition for multi-tenancy
milvusv2.WithDropOld()                  // Drop existing collection

// Index options
milvusv2.WithIndex(entity.NewHNSWIndex(entity.COSINE, 16, 200))
milvusv2.WithMetricType(entity.COSINE)  // L2, IP, COSINE, HAMMING, JACCARD

// Performance options
milvusv2.WithShards(2)                  // Number of shards
milvusv2.WithMaxTextLength(1000)        // Max text field length
milvusv2.WithSkipFlushOnWrite()         // Skip immediate flush
```

### Index Types

```go
// Auto index (recommended)
entity.NewAutoIndex(entity.COSINE)

// Flat index (exact search)
entity.NewFlatIndex(entity.COSINE)

// IVF index (balanced)
entity.NewIvfFlatIndex(entity.COSINE, 128)

// HNSW index (high recall, fast)
entity.NewHNSWIndex(entity.COSINE, 16, 200)
```

### Add Documents

```go
docs := []schema.Document{
    {
        PageContent: "Your document content",
        Metadata: map[string]any{
            "category": "tech",
            "source": "doc1",
        },
    },
}

ids, err := store.AddDocuments(ctx, docs)
```

### Search Documents

```go
// Similarity search
results, err := store.SimilaritySearch(ctx, "query text", 5)

// With options
results, err := store.SimilaritySearch(
    ctx,
    "query text",
    5,
    []vectorstores.Option{
        vectorstores.WithScoreThreshold(0.8),
    },
)
```

### Create RAG Pipeline

```go
// Wrap with LangGraphGo adapter
langGraphStore := rag.NewLangChainVectorStore(store)

// Create retriever
retriever := retriever.NewVectorStoreRetriever(langGraphStore, embedder, 2)

// Build pipeline
config := rag.DefaultPipelineConfig()
config.Retriever = retriever
config.LLM = llm

pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()

runnable, _ := pipeline.Compile()
```

## Index Selection Guide

### HNSW Index
- **Use case**: High recall, fast search
- **Parameters**: M (max connections), efConstruction (build time)
- **Memory**: Higher memory usage
- **Best for**: Real-time applications with strict latency requirements

```go
entity.NewHNSWIndex(entity.COSINE, 16, 200)
```

### IVF Index
- **Use case**: Balanced performance and memory
- **Parameters**: nlist (number of clusters)
- **Memory**: Moderate memory usage
- **Best for**: Large-scale datasets with memory constraints

```go
entity.NewIvfFlatIndex(entity.COSINE, 128)
```

### Flat Index
- **Use case**: Exact search, small datasets
- **Parameters**: None
- **Memory**: Low memory usage
- **Best for**: Small datasets, exact match requirements

```go
entity.NewFlatIndex(entity.COSINE)
```

### Auto Index
- **Use case**: Let Milvus decide
- **Parameters**: None
- **Memory**: Optimized automatically
- **Best for**: Quick prototyping, dynamic workloads

```go
entity.NewAutoIndex(entity.COSINE)
```

## Distance Metrics

```go
// L2 Distance (Euclidean)
entity.L2

// Inner Product
entity.IP

// Cosine Similarity
entity.COSINE

// Hamming Distance (for binary vectors)
entity.HAMMING

// Jaccard Distance (for binary vectors)
entity.JACCARD
```

## Advanced Features

### Partitions for Multi-Tenancy

```go
// Create store with partition
store, err := milvusv2.New(
    ctx,
    milvusConfig,
    milvusv2.WithPartitionName("tenant_123"),
)

// Each tenant has isolated data
// Search only within the tenant's partition
```

### Scalar Filtering

```go
// Combine vector search with metadata filters
results, err := store.SimilaritySearch(
    ctx,
    "query",
    5,
    []vectorstores.Option{
        vectorstores.WithFilters(map[string]interface{}{
            "category": "tech",
            "year":    2024,
        }),
    },
)
```

### Replicas for Scaling

```bash
# Create replicas via Milvus API or UI
# Allows scaling read operations
```

## Production Deployment

### Standalone Deployment

Good for:
- Development and testing
- Small to medium datasets (< 10M vectors)
- Single-server deployments

```bash
docker run -d --name milvus-standalone \
  -p 19530:19530 \
  -v milvus:/var/lib/milvus \
  milvusdb/milvus:latest
```

### Cluster Deployment

Good for:
- Production workloads
- Large datasets (> 10M vectors)
- High availability requirements

```bash
# Use Milvus Operator on Kubernetes
kubectl apply -f https://github.com/milvus-io/milvus/releases/download/v2.4.0/milvus-operator.yaml
```

### Cloud Services

- [Zilliz Cloud](https://zilliz.com/) - Fully managed Milvus
- [AWS Marketplace](https://aws.amazon.com/marketplace)
- [Google Cloud Marketplace](https://cloud.google.com/marketplace)

## Performance Tuning

### Index Tuning

```go
// For high recall (95%+)
entity.NewHNSWIndex(entity.COSINE, 32, 200)

// For balanced performance
entity.NewIvfFlatIndex(entity.COSINE, 256)

// For fast insertion
entity.NewFlatIndex(entity.COSINE)
```

### Search Parameters

```go
// Adjust search time vs accuracy
options := []vectorstores.Option{
    vectorstores.WithScoreThreshold(0.7),    // Filter low scores
    vectorstores.WithTopK(10),                // Get more results
}
```

### Sharding Strategy

```go
// Distribute data across shards
milvusv2.WithShards(4)  // 4 shards for parallel processing
```

## Comparison with Other Vector Stores

| Feature | Milvus | Pinecone | Weaviate | sqlite-vec |
|---------|--------|----------|----------|------------|
| Self-hosted | ✅ Yes | ❌ No | ✅ Yes | ✅ Yes |
| Cloud Managed | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No |
| Billion Scale | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No |
| Partitions | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No |
| Replicas | ✅ Yes | ✅ Yes | ✅ Yes | ❌ No |
| Complexity | ⭐⭐⭐ High | ⭐ Low | ⭐⭐ Medium | ⭐ Low |

## Troubleshooting

### Connection Issues

```bash
# Check if Milvus is running
docker ps | grep milvus

# Check Milvus logs
docker logs milvus-standalone

# Test connection
telnet localhost 19530
```

### Collection Already Exists

```go
// Use WithDropOld() to drop existing collection
milvusv2.WithDropOld()
```

### Memory Issues

```go
// Reduce shard count
milvusv2.WithShards(1)

// Use more memory-efficient index
entity.NewIvfFlatIndex(entity.COSINE, 64)
```

## Best Practices

1. **Choose the right index** - Start with AutoIndex, then optimize
2. **Use partitions** - Enable multi-tenancy and improve performance
3. **Set appropriate metrics** - Use COSINE for normalized embeddings
4. **Monitor performance** - Use Milvus metrics and monitoring
5. **Regular backups** - Backup collection data and metadata
6. **Schema design** - Plan your schema and indexing strategy ahead

## See Also

- [Milvus Documentation](https://milvus.io/docs)
- [Milvus Go SDK](https://github.com/milvus-io/milvus-sdk-go)
- [RAG Pipeline Documentation](../../rag/README.md)
- [LangChain Milvus Integration](https://github.com/tmc/langchaingo/tree/main/vectorstores/milvus/v2)
- [Other Vector Store Examples](../rag_sqlitevec_example/)

## License

This example follows the LangGraphGo project license.
