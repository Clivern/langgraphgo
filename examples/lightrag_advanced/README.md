# LightRAG Advanced Example

An advanced example demonstrating LightRAG's features with custom configurations and performance comparisons.

## Overview

This comprehensive example showcases:
- **Custom prompt templates** for entity and relationship extraction
- **Community detection** for global retrieval
- **Fusion method comparison** (RRF vs Weighted)
- **Performance benchmarking** across retrieval modes
- **Knowledge graph traversal**
- **Document operations** (add, update, delete)

## Prerequisites

```bash
go mod tidy
```

## Running the Example

### Without OpenAI API Key (Mock LLM)

```bash
go run main.go
```

### With OpenAI API Key

For better entity extraction:

```bash
export OPENAI_API_KEY="your-api-key-here"
go run main.go
```

## Features Demonstrated

### 1. Custom Configuration

```go
config := rag.LightRAGConfig{
    Mode:                 "hybrid",
    Temperature:          0.7,
    ChunkSize:            512,
    MaxEntitiesPerChunk:  20,
    EnableCommunityDetection: true,
    PromptTemplates: map[string]string{
        "entity_extraction": "...",
        "relationship_extraction": "...",
    },
}
```

### 2. Retrieval Mode Comparison

The example runs the same query across all four modes:
- **Naive**: Fast, basic retrieval
- **Local**: Entity-centric with multi-hop reasoning
- **Global**: Community-level summaries
- **Hybrid**: Best of both worlds

### 3. Fusion Method Comparison

Compares two fusion strategies for hybrid mode:
- **RRF (Reciprocal Rank Fusion)**: Rank-based fusion
- **Weighted**: Score-based weighted fusion

### 4. Knowledge Graph Operations

```go
// Query the knowledge graph
result, err := graphKg.Query(ctx, &rag.GraphQuery{
    EntityTypes: []string{"CONCEPT", "TECHNOLOGY"},
    Limit:       5,
})
```

### 5. Document Operations

```go
// Add new document
lightrag.AddDocuments(ctx, []rag.Document{newDoc})

// Update document
lightrag.UpdateDocument(ctx, updatedDoc)
```

### 6. Performance Benchmarking

Runs multiple queries to measure:
- Average response time per mode
- Success rate
- Latency statistics

## Sample Documents

The example uses documents about AI and Machine Learning:
- Transformer architecture
- Neural networks
- Large Language Models (LLMs)
- Machine learning fundamentals
- Attention mechanisms
- RAG (Retrieval-Augmented Generation)
- Fine-tuning
- Embeddings

## Expected Output

```
=== LightRAG Advanced Example ===
This example demonstrates advanced features of LightRAG including:
- Custom prompt templates
- Community detection
- Different fusion methods
- Performance comparison between modes

Indexing documents...
Indexed 8 documents in 15ms

=== Retrieval Mode Comparison ===

--- Naive Mode ---
Response Time: 250µs
Sources Retrieved: 3
Confidence: 0.15

--- Local Mode ---
Response Time: 450µs
Sources Retrieved: 5
Query Entities: 2

--- Global Mode ---
Response Time: 380µs
Sources Retrieved: 4
Communities: 2

--- Hybrid Mode ---
Response Time: 520µs
Sources: 5
Local Confidence: 0.25
Global Confidence: 0.18
...
```

## Advanced Configuration Options

### Local Retrieval

```go
LocalConfig: rag.LocalRetrievalConfig{
    TopK:               15,        // Number of entities
    MaxHops:            3,         // Graph traversal depth
    IncludeDescriptions: true,
    EntityWeight:       0.8,       // Entity relevance weight
}
```

### Global Retrieval

```go
GlobalConfig: rag.GlobalRetrievalConfig{
    MaxCommunities:     10,        // Communities to retrieve
    IncludeHierarchy:   true,      // Include hierarchy
    MaxHierarchyDepth:  5,         // Hierarchy depth
    CommunityWeight:    0.7,       // Community relevance
}
```

### Hybrid Fusion

```go
HybridConfig: rag.HybridRetrievalConfig{
    LocalWeight:  0.6,             // 60% local
    GlobalWeight: 0.4,             // 40% global
    FusionMethod: "rrf",           // or "weighted"
    RFFK:         60,              // RRF parameter
}
```

## Performance Tips

1. **Chunk Size**: Larger chunks = fewer API calls, less precise retrieval
2. **MaxHops**: Limit to 2-3 for performance
3. **TopK**: Start with 10-20, adjust based on results
4. **Community Detection**: Disable for small datasets (< 100 docs)

## Code Structure

- `main()`: Orchestrates all demonstrations
- `OpenAILLMAdapter`: Wraps OpenAI LLM
- `MockLLM`: Mock implementation for demo
- `createSampleDocuments()`: Creates test documents

## See Also

- [LightRAG Simple Example](../lightrag_simple/) - Basic usage
- [LightRAG Documentation](../../docs/lightrag.md) - Full documentation
