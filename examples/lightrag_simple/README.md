# LightRAG Simple Example

A simple example demonstrating the four retrieval modes of LightRAG.

## Overview

This example shows how to use LightRAG with different retrieval modes:
- **Naive**: Simple vector similarity search
- **Local**: Entity-based retrieval with graph traversal
- **Global**: Community-level retrieval
- **Hybrid**: Combines local and global retrieval

## Prerequisites

```bash
go mod tidy
```

## Running the Example

### Without OpenAI API Key (Mock LLM)

The example includes a Mock LLM for demonstration purposes:

```bash
go run main.go
```

### With OpenAI API Key

For production use, set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
go run main.go
```

## What the Example Does

1. **Creates a LightRAG engine** with hybrid mode configuration
2. **Adds sample documents** about LangGraph, LightRAG, Knowledge Graphs, Vector Databases, and RAG
3. **Tests each retrieval mode** with three different queries:
   - "What is LightRAG and how does it work?"
   - "Explain the relationship between RAG and knowledge graphs"
   - "What are the benefits of using vector databases?"
4. **Displays results** including retrieved sources, confidence scores, and response times

## Expected Output

```
=== LightRAG Simple Example ===

Adding documents to LightRAG...
Successfully indexed 5 documents

=== LightRAG Configuration ===
Mode: hybrid
Chunk Size: 512
Chunk Overlap: 50
...

=== Testing NAIVE Mode ===
--- Query 1: What is LightRAG and how does it work? ---
Retrieved 3 sources
Confidence: 0.03
Response Time: 15.209µs
...

=== Testing LOCAL Mode ===
...

=== Testing GLOBAL Mode ===
...

=== Testing HYBRID Mode ===
...
```

## Configuration

The example uses the following configuration:

```go
config := rag.LightRAGConfig{
    Mode:                 "hybrid",
    ChunkSize:            512,
    ChunkOverlap:         50,
    MaxEntitiesPerChunk:  20,
    LocalConfig: rag.LocalRetrievalConfig{
        TopK:               10,
        MaxHops:            2,
        IncludeDescriptions: true,
    },
    GlobalConfig: rag.GlobalRetrievalConfig{
        MaxCommunities:     5,
        IncludeHierarchy:   false,
    },
    HybridConfig: rag.HybridRetrievalConfig{
        LocalWeight:  0.5,
        GlobalWeight: 0.5,
        FusionMethod: "rrf",
    },
}
```

## Code Structure

- `main()`: Sets up the LightRAG engine and runs queries
- `OpenAILLMAdapter`: Wraps OpenAI LLM to implement `rag.LLMInterface`
- `MockLLM`: Mock implementation for demonstration without API key

## See Also

- [LightRAG Advanced Example](../lightrag_advanced/) - More comprehensive example with performance comparison
- [LightRAG Documentation](../../docs/lightrag.md) - Full documentation
