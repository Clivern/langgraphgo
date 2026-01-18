# BM25 Integration for LangGraphGo

## Overview

BM25 (Best Matching 25) is a probabilistic information retrieval function that ranks documents based on query terms appearing in the document. It's a sparse retrieval method that works well for keyword-based search and can be combined with dense vector retrieval for hybrid search systems.

## Features

- **Sparse Retrieval**: Fast keyword-based document retrieval
- **Configurable Parameters**: Tune k1 and b parameters for optimal performance
- **Multiple Tokenizers**: Support for English, Chinese, regex, and n-gram tokenization
- **Dynamic Index**: Add, update, and delete documents at runtime
- **Hybrid Search**: Combine with vector retrievers for best results
- **Score Thresholding**: Filter results by minimum relevance score

## Installation

The BM25 retriever is included in the `rag/retriever` package:

```go
import "github.com/smallnest/langgraphgo/rag/retriever"
```

## Basic Usage

### Creating a BM25 Retriever

```go
import (
    "context"
    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/retriever"
)

// Prepare documents
docs := []rag.Document{
    {
        ID:      "doc1",
        Content: "Machine learning is a subset of artificial intelligence",
        Metadata: map[string]any{
            "title": "ML Overview",
        },
    },
    {
        ID:      "doc2",
        Content: "Deep learning uses neural networks",
        Metadata: map[string]any{
            "title": "DL Overview",
        },
    },
}

// Create BM25 retriever with default config
config := retriever.DefaultBM25Config()
config.K = 3 // Retrieve top 3 documents

bm25Retriever, err := retriever.NewBM25Retriever(docs, config)
if err != nil {
    log.Fatal(err)
}

// Query documents
ctx := context.Background()
results, err := bm25Retriever.Retrieve(ctx, "neural networks")
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("[%s] %s\n", result.ID, result.Content)
}
```

### Getting Scores with RetrieveWithConfig

```go
// Use RetrieveWithConfig to get relevance scores
retrievalConfig := &rag.RetrievalConfig{
    K:              5,
    ScoreThreshold: 0.5, // Only return docs with score >= 0.5
}

results, err := bm25Retriever.RetrieveWithConfig(ctx, query, retrievalConfig)
if err != nil {
    log.Fatal(err)
}

for _, result := range results {
    fmt.Printf("Score: %.4f | %s\n", result.Score, result.Document.Content)
}
```

## Configuration

### BM25 Parameters

BM25 has two key parameters that affect ranking:

```go
type BM25Config struct {
    // k1 controls term frequency saturation (1.2 - 2.0)
    // Higher k1 = more weight to term frequency
    K1 float64

    // b controls document length normalization (0 - 1)
    // 0 = no normalization, 1 = full normalization
    B float64

    // K is the number of documents to retrieve
    K int

    // ScoreThreshold filters results by minimum score
    ScoreThreshold float64
}
```

### Tuning Parameters

```go
// For documents with varying lengths, increase b
config.B = 0.75 // Default, good for most cases

// For longer queries, decrease k1
config.K1 = 1.2 // Less term frequency emphasis

// For short queries, increase k1
config.K1 = 2.0 // More term frequency emphasis
```

## Tokenizers

### Default Word Tokenizer

```go
// Default regex-based tokenizer (matches words)
bm25Retriever, err := retriever.NewBM25Retriever(docs, config)
```

### Chinese Text Tokenizer

```go
import "github.com/smallnest/langgraphgo/rag/tokenizer"

chineseTokenizer := tokenizer.NewChineseTokenizer()
bm25Retriever, err := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    chineseTokenizer,
)
```

### Custom Regex Tokenizer

```go
// Create tokenizer with custom pattern
regexTokenizer, err := tokenizer.NewRegexTokenizer(`\b\w+\b`)
if err != nil {
    log.Fatal(err)
}

bm25Retriever, err := retriever.NewBM25RetrieverWithTokenizer(
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

bm25Retriever, err := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    bigramTokenizer,
)
```

## Dynamic Document Management

### Adding Documents

```go
newDocs := []rag.Document{
    {ID: "doc3", Content: "New document content"},
    {ID: "doc4", Content: "Another new document"},
}

bm25Retriever.AddDocuments(newDocs)
fmt.Printf("Total documents: %d\n", bm25Retriever.GetDocumentCount())
```

### Updating Documents

```go
bm25Retriever.UpdateDocument(rag.Document{
    ID:      "doc1",
    Content: "Updated content",
})
```

### Deleting Documents

```go
bm25Retriever.DeleteDocument("doc2")
```

## Hybrid Retrieval

BM25 works great in combination with vector retrieval for hybrid search:

```go
import (
    "github.com/smallnest/langgraphgo/rag/retriever"
)

// Create BM25 retriever (sparse)
bm25Config := retriever.DefaultBM25Config()
bm25Retriever, _ := retriever.NewBM25Retriever(docs, bm25Config)

// Create vector retriever (dense)
vectorConfig := rag.RetrievalConfig{K: 5}
vectorRetriever := retriever.NewVectorRetriever(vectorStore, embedder, vectorConfig)

// Combine both with weights
hybridRetriever := retriever.NewHybridRetriever(
    []rag.Retriever{bm25Retriever, vectorRetriever},
    []float64{0.5, 0.5}, // Equal weights
    rag.RetrievalConfig{K: 5},
)

// Query using hybrid retrieval
results, err := hybridRetriever.Retrieve(ctx, "your query here")
```

## Performance Tips

1. **For exact keyword matching**: Use BM25 alone
2. **For semantic understanding**: Use vector retrieval alone
3. **For best results**: Use hybrid retrieval with both BM25 and vector
4. **Index size**: BM25 index is typically smaller than vector index
5. **Query speed**: BM25 is faster than vector retrieval for large datasets

## Index Statistics

Get information about your BM25 index:

```go
stats := bm25Retriever.GetStats()

fmt.Printf("Documents: %v\n", stats["num_documents"])
fmt.Printf("Unique terms: %v\n", stats["num_unique_terms"])
fmt.Printf("Avg doc length: %.1f\n", stats["avg_doc_length"])
fmt.Printf("k1: %.2f\n", stats["k1"])
fmt.Printf("b: %.2f\n", stats["b"])
```

## Example: Complete RAG Pipeline

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
    // 1. Load documents
    docs := loadDocuments()

    // 2. Create BM25 retriever
    config := retriever.DefaultBM25Config()
    config.K = 5

    bm25Retriever, err := retriever.NewBM25Retriever(docs, config)
    if err != nil {
        log.Fatal(err)
    }

    // 3. Query for relevant documents
    ctx := context.Background()
    query := "machine learning algorithms"

    results, err := bm25Retriever.Retrieve(ctx, query)
    if err != nil {
        log.Fatal(err)
    }

    // 4. Use retrieved documents for generation
    context := buildContext(results)
    response := generateWithLLM(query, context)

    fmt.Printf("Query: %s\n", query)
    fmt.Printf("Response: %s\n", response)
}

func loadDocuments() []rag.Document {
    // Load your documents here
    return []rag.Document{
        {ID: "1", Content: "Document 1 content..."},
        {ID: "2", Content: "Document 2 content..."},
    }
}

func buildContext(docs []rag.Document) string {
    var context string
    for _, doc := range docs {
        context += doc.Content + "\n"
    }
    return context
}

func generateWithLLM(query, context string) string {
    // Use your LLM to generate response
    return "Generated response based on retrieved context"
}
```

## API Reference

### Types

- `BM25Retriever`: Main BM25 retriever implementation
- `BM25Config`: Configuration for BM25 parameters
- `Tokenizer`: Interface for text tokenization

### Functions

- `NewBM25Retriever(docs, config)`: Create BM25 retriever with default tokenizer
- `NewBM25RetrieverWithTokenizer(docs, config, tokenizer)`: Create with custom tokenizer
- `DefaultBM25Config()`: Get default configuration

### Methods

- `Retrieve(ctx, query)`: Retrieve documents (returns []Document)
- `RetrieveWithK(ctx, query, k)`: Retrieve exactly k documents
- `RetrieveWithConfig(ctx, query, config)`: Retrieve with custom config (returns scores)
- `AddDocuments(docs)`: Add documents to index
- `UpdateDocument(doc)`: Update existing document
- `DeleteDocument(id)`: Remove document from index
- `GetDocumentCount()`: Get number of indexed documents
- `GetStats()`: Get index statistics

## Comparison with LangChain Python

| Feature | LangGraphGo | LangChain Python |
|---------|-------------|------------------|
| Basic BM25 | ✅ | ✅ |
| Custom Tokenizers | ✅ | ✅ |
| Parameter Tuning | ✅ | ✅ |
| Dynamic Updates | ✅ | ✅ |
| Chinese Support | ✅ Built-in | Via external libs |
| Hybrid Search | ✅ Built-in | Via custom code |
| Thread-safe | ✅ | ❌ GIL limited |

## License

Part of LangGraphGo project. See main LICENSE file for details.
