# BM25 Retriever Example

This example demonstrates how to use the BM25 (Best Matching 25) sparse retrieval system in LangGraphGo.

## Overview

BM25 is a probabilistic information retrieval function that ranks documents based on query term frequency. It's particularly effective for keyword-based search and can be combined with vector retrieval for hybrid search systems.

## Features Demonstrated

1. **Basic BM25 Retrieval** - Simple keyword-based document search
2. **Score Thresholding** - Filter results by minimum relevance score
3. **Custom Tokenizers** - Support for English, Chinese, and custom regex tokenization
4. **Dynamic Document Management** - Add, update, and delete documents at runtime
5. **Parameter Tuning** - Adjust k1 and b parameters for optimal performance
6. **Hybrid Retrieval Setup** - Combine BM25 with vector retrieval

## Running the Example

### Build and Run

```bash
go build -o bm25_demo main.go
./bm25_demo
```

Or run directly:

```bash
go run main.go
```

## Code Examples

### Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/retriever"
)

func main() {
    // Prepare documents
    docs := []rag.Document{
        {
            ID:      "doc1",
            Content: "LangGraph is a framework for building stateful applications with LLMs",
            Metadata: map[string]any{
                "title": "LangGraph Overview",
            },
        },
        {
            ID:      "doc2",
            Content: "BM25 is a ranking function for information retrieval",
            Metadata: map[string]any{
                "title": "BM25 Overview",
            },
        },
    }

    // Create BM25 retriever
    config := retriever.DefaultBM25Config()
    config.K = 2 // Retrieve top 2 documents

    bm25Retriever, err := retriever.NewBM25Retriever(docs, config)
    if err != nil {
        log.Fatal(err)
    }

    // Query documents
    ctx := context.Background()
    results, err := bm25Retriever.Retrieve(ctx, "framework LLM")
    if err != nil {
        log.Fatal(err)
    }

    // Display results
    fmt.Printf("Found %d results:\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. [%s] %s\n", i+1, result.Metadata["title"], result.Content)
    }
}
```

### Using Score Threshold

```go
// Only return documents with score >= 0.5
config := retriever.DefaultBM25Config()
config.K = 10
config.ScoreThreshold = 0.5

bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

ctx := context.Background()
retrievalConfig := &rag.RetrievalConfig{
    K:              10,
    ScoreThreshold: 0.5,
}

results, _ := bm25Retriever.RetrieveWithConfig(ctx, query, retrievalConfig)

for _, result := range results {
    fmt.Printf("Score: %.4f | %s\n", result.Score, result.Document.Content)
}
```

### Chinese Text Support

```go
import "github.com/smallnest/langgraphgo/rag/tokenizer"

// Create documents with Chinese text
docs := []rag.Document{
    {ID: "doc1", Content: "Go语言支持并发编程"},
    {ID: "doc2", Content: "Python是一种易学的编程语言"},
}

// Use Chinese tokenizer
chineseTokenizer := tokenizer.NewChineseTokenizer()

config := retriever.DefaultBM25Config()
bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    chineseTokenizer,
)

ctx := context.Background()
results, _ := bm25Retriever.Retrieve(ctx, "编程语言")
```

### Dynamic Document Management

```go
// Add new documents
newDocs := []rag.Document{
    {ID: "doc3", Content: "New document content"},
}
bm25Retriever.AddDocuments(newDocs)

// Update existing document
bm25Retriever.UpdateDocument(rag.Document{
    ID:      "doc1",
    Content: "Updated content",
})

// Delete document
bm25Retriever.DeleteDocument("doc2")

// Get statistics
stats := bm25Retriever.GetStats()
fmt.Printf("Total documents: %v\n", stats["num_documents"])
```

### Hybrid Retrieval (BM25 + Vector)

```go
// Create BM25 retriever (sparse retrieval)
bm25Config := retriever.DefaultBM25Config()
bm25Retriever, _ := retriever.NewBM25Retriever(docs, bm25Config)

// Create vector retriever (dense retrieval)
vectorConfig := rag.RetrievalConfig{K: 5}
vectorRetriever := retriever.NewVectorRetriever(
    vectorStore,
    embedder,
    vectorConfig,
)

// Combine both with equal weights
hybridRetriever := retriever.NewHybridRetriever(
    []rag.Retriever{bm25Retriever, vectorRetriever},
    []float64{0.5, 0.5},
    rag.RetrievalConfig{K: 5},
)

// Query using hybrid retrieval
ctx := context.Background()
results, _ := hybridRetriever.Retrieve(ctx, "your query")
```

## BM25 Parameters

### k1 Parameter (Term Frequency Saturation)

Controls how much term frequency affects the score:

- **Range**: 1.2 - 2.0
- **Lower values** (e.g., 1.2): Less emphasis on term frequency
- **Higher values** (e.g., 2.0): More emphasis on term frequency
- **Use cases**:
  - Short queries: Use higher k1
  - Long queries: Use lower k1

```go
config := retriever.DefaultBM25Config()
config.K1 = 1.5 // Default value
```

### b Parameter (Document Length Normalization)

Controls how document length affects the score:

- **Range**: 0.0 - 1.0
- **0.0**: No length normalization
- **1.0**: Full length normalization
- **Recommended**: 0.75 (default)

```go
config := retriever.DefaultBM25Config()
config.B = 0.75 // Default value, works well for most cases
```

## Tokenizers

### Default Word Tokenizer

```go
// Automatically uses regex-based word tokenization
bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)
```

### Chinese Tokenizer

```go
import "github.com/smallnest/langgraphgo/rag/tokenizer"

chineseTokenizer := tokenizer.NewChineseTokenizer()
bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    chineseTokenizer,
)
```

### Custom Regex Tokenizer

```go
// Create tokenizer with custom pattern
regexTokenizer, _ := tokenizer.NewRegexTokenizer(`\b[a-zA-Z]+\b`)
bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    regexTokenizer,
)
```

### N-gram Tokenizer

```go
// Create bigram tokenizer
baseTokenizer := tokenizer.DefaultRegexTokenizer()
bigramTokenizer := tokenizer.NewNgramTokenizer(2, baseTokenizer)

bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    bigramTokenizer,
)
```

## Expected Output

When you run the example, you should see output similar to:

```
=== BM25 Retriever Demo ===

1. Basic BM25 Retrieval
-----------------------
Query: framework for building LLM applications
Found 2 results:
  1. [LangGraph Overview] LangGraph is a framework for building stateful...
  2. [RAG Overview] RAG combines retrieval systems with generation...

2. BM25 with Score Threshold
----------------------------
Query: neural networks learning (threshold: 0.5)
Found 3 results above threshold:
  1. [Score: 3.1372] Deep learning uses neural networks...
  2. [Score: 0.9276] Machine learning algorithms learn...
  3. [Score: 0.9276] Neural networks are inspired by...

3. BM25 with Custom Tokenizer
------------------------------
Query: 编程语言
Found 2 results:
  1. Go语言支持并发编程
  2. Python是一种易学的编程语言

4. Hybrid Retrieval Setup
-------------------------
BM25 results for 'search similarity matching':
  1. [Score: 1.3852] Vector search uses embeddings...
  2. [Score: 0.9365] BM25 uses term frequency...

5. Dynamic Document Management
-------------------------------
Initial document count: 1
After adding: 3
Updated doc1
After deleting doc2: 2

6. Parameter Tuning
-------------------
Query: quick fast
Testing k1 parameter (term frequency saturation):
  k1=0.5: [doc1:0.98] [doc2:0.98]
  k1=1.5: [doc1:0.98] [doc2:0.98]
```

## Use Cases

1. **Keyword Search**: When users search for specific terms
2. **Hybrid Search**: Combine with vector retrieval for semantic + keyword matching
3. **Multi-language**: Support for English, Chinese, and other languages
4. **Dynamic Indexes**: When documents change frequently
5. **Fast Prototyping**: Quick retrieval without vector databases

## Performance Tips

- **For exact keyword matching**: Use BM25 alone
- **For semantic understanding**: Use vector retrieval alone
- **For best results**: Use hybrid retrieval with both BM25 and vector
- **Index size**: BM25 index is typically smaller than vector index
- **Query speed**: BM25 is faster than vector retrieval for large datasets

## Comparison: BM25 vs Vector Retrieval

| Feature | BM25 | Vector |
|---------|------|--------|
| Type | Sparse (keyword) | Dense (semantic) |
| Index Size | Small | Large |
| Query Speed | Fast | Slower |
| Exact Match | Excellent | Poor |
| Semantic Understanding | Poor | Excellent |
| Best For | Keywords, technical terms | Concepts, meaning |

## See Also

- [BM25 Integration Documentation](../../docs/bm25_integration.md)
- [BM25 Summary](../../docs/bm25_summary.md)
- [LangGraphGo Documentation](../../README.md)

## License

Part of LangGraphGo project. See main LICENSE file for details.
