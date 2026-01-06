# LightRAG

LightRAG is a lightweight Retrieval-Augmented Generation framework that combines low-level semantic chunks with high-level graph structures. It provides four retrieval modes to support different use cases.

## Overview

LightRAG implements a dual-layer retrieval system:

- **Semantic Layer**: Traditional vector-based retrieval for finding similar content
- **Graph Layer**: Knowledge graph-based retrieval for finding related entities and relationships

This combination allows LightRAG to provide more comprehensive and contextually relevant results than traditional RAG systems.

## Retrieval Modes

### 1. Naive Mode

Simple retrieval using vector similarity search without graph structure.

**Use cases**:
- Quick prototyping
- Simple document search
- When relationship information is not important

**Example**:
```go
config := rag.LightRAGConfig{
    Mode: "naive",
    // ... other config
}
```

### 2. Local Mode

Retrieves relevant entities and their relationships within a localized context through knowledge graph traversal.

**Use cases**:
- Finding related concepts
- Multi-hop reasoning
- Entity-centric queries

**Configuration**:
```go
config := rag.LightRAGConfig{
    Mode: "local",
    LocalConfig: rag.LocalRetrievalConfig{
        TopK:               10,    // Number of entities to retrieve
        MaxHops:            2,     // Maximum hops in the graph
        IncludeDescriptions: true,  // Include entity descriptions
        EntityWeight:       0.8,   // Weight for entity relevance
    },
}
```

### 3. Global Mode

Retrieves information from community-level summaries, providing a high-level view of the knowledge graph.

**Use cases**:
- Understanding broad topics
- Domain overviews
- Exploring community structures

**Configuration**:
```go
config := rag.LightRAGConfig{
    Mode: "global",
    GlobalConfig: rag.GlobalRetrievalConfig{
        MaxCommunities:     5,   // Number of communities to retrieve
        IncludeHierarchy:   true, // Include community hierarchy
        MaxHierarchyDepth:  3,   // Maximum hierarchy depth
        CommunityWeight:    0.7, // Weight for community relevance
    },
    EnableCommunityDetection: true,
}
```

### 4. Hybrid Mode

Combines local and global retrieval results for comprehensive answers.

**Use cases**:
- Complex queries requiring multiple perspectives
- When both detail and overview are needed
- General-purpose RAG applications

**Configuration**:
```go
config := rag.LightRAGConfig{
    Mode: "hybrid",
    HybridConfig: rag.HybridRetrievalConfig{
        LocalWeight:  0.5,  // Weight for local results
        GlobalWeight: 0.5,  // Weight for global results
        FusionMethod: "rrf", // "rrf" or "weighted"
        RFFK:         60,   // RRF parameter
    },
}
```

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/engine"
    "github.com/smallnest/langgraphgo/rag/store"
    "github.com/tmc/langchaingo/llms/openai"
)

func main() {
    ctx := context.Background()

    // Initialize LLM and embedder
    llm, _ := openai.New()
    embedder := store.NewMockEmbedder(128)

    // Create knowledge graph and vector store
    kg, _ := store.NewKnowledgeGraph("memory://")
    vectorStore := store.NewInMemoryVectorStore(embedder)

    // Configure LightRAG
    config := rag.LightRAGConfig{
        Mode:         "hybrid",
        ChunkSize:    512,
        ChunkOverlap: 50,
    }

    // Create LightRAG engine
    lightrag, _ := engine.NewLightRAGEngine(config, llm, embedder, kg, vectorStore)

    // Add documents
    documents := []rag.Document{
        {
            ID:        "doc1",
            Content:   "Your document content here...",
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }
    lightrag.AddDocuments(ctx, documents)

    // Query
    result, _ := lightrag.Query(ctx, "Your question here?")
    log.Printf("Found %d sources", len(result.Sources))
}
```

## Configuration Options

### Chunking

```go
config := rag.LightRAGConfig{
    ChunkSize:    512,  // Size of each chunk in characters
    ChunkOverlap: 50,   // Overlap between chunks
}
```

### Entity Extraction

```go
config := rag.LightRAGConfig{
    MaxEntitiesPerChunk:        20,   // Maximum entities per chunk
    EntityExtractionThreshold:  0.5,  // Threshold for entity extraction
    Temperature:                0.7,  // Temperature for LLM operations
}
```

### Community Detection

```go
config := rag.LightRAGConfig{
    EnableCommunityDetection:      true,
    CommunityDetectionAlgorithm:  "louvain",  // "louvain", "leiden", or "label_propagation"
    MaxCommunities:               10,
}
```

### Custom Prompt Templates

```go
config := rag.LightRAGConfig{
    PromptTemplates: map[string]string{
        "entity_extraction": "Your custom entity extraction prompt...",
        "relationship_extraction": "Your custom relationship extraction prompt...",
    },
}
```

## Fusion Methods

### Reciprocal Rank Fusion (RRF)

RRF combines results from multiple sources using reciprocal rank scoring:

```go
config := rag.LightRAGConfig{
    HybridConfig: rag.HybridRetrievalConfig{
        FusionMethod: "rrf",
        RFFK:         60,  // RRF parameter (higher = more damping)
    },
}
```

### Weighted Fusion

Weighted fusion combines scores using configured weights:

```go
config := rag.LightRAGConfig{
    HybridConfig: rag.HybridRetrievalConfig{
        FusionMethod: "weighted",
        LocalWeight:  0.6,
        GlobalWeight: 0.4,
    },
}
```

## Advanced Features

### Knowledge Graph Access

```go
// Get the underlying knowledge graph
kg := lightrag.GetKnowledgeGraph()

// Query the graph
result, _ := kg.Query(ctx, &rag.GraphQuery{
    EntityTypes: []string{"PERSON", "ORGANIZATION"},
    Limit:       10,
})

// Get related entities
entities, _ := kg.GetRelatedEntities(ctx, entityID, maxDepth)
```

### Metrics

```go
metrics := lightrag.GetMetrics()
log.Printf("Total Queries: %d", metrics.TotalQueries)
log.Printf("Average Latency: %v", metrics.AverageLatency)
log.Printf("Total Documents: %d", metrics.TotalDocuments)
```

### Document Operations

```go
// Add documents
err := lightrag.AddDocuments(ctx, documents)

// Update document
err := lightrag.UpdateDocument(ctx, document)

// Delete document
err := lightrag.DeleteDocument(ctx, docID)
```

## Examples

See the examples directory for complete working examples:

- `examples/lightrag_simple/main.go` - Basic usage with all retrieval modes
- `examples/lightrag_advanced/main.go` - Advanced features and performance comparison

Run examples:

```bash
# Simple example
go run examples/lightrag_simple/main.go

# Advanced example
go run examples/lightrag_advanced/main.go
```

## Performance Considerations

### Choosing the Right Mode

- **Naive**: Fastest, lowest quality
- **Local**: Medium speed, good for entity-centric queries
- **Global**: Medium speed, good for broad topics
- **Hybrid**: Slowest but most comprehensive

### Optimization Tips

1. **Chunk Size**: Larger chunks = fewer API calls but less precise retrieval
2. **MaxHops**: Limit to 2-3 for performance
3. **TopK**: Start with 10-20, adjust based on results
4. **Community Detection**: Disable for small datasets (< 100 documents)

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        LightRAG Engine                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────┐         ┌──────────────┐                  │
│  │   Naive     │         │    Local     │                  │
│  │  Retrieval  │         │  Retrieval   │                  │
│  │  (Vector)   │         │   (Graph)    │                  │
│  └─────────────┘         └──────────────┘                  │
│         │                        │                          │
│         └────────────┬───────────┘                          │
│                      │                                      │
│              ┌───────▼────────┐                            │
│              │   Hybrid Mode  │                            │
│              │   (Fusion)     │                            │
│              └───────┬────────┘                            │
│                      │                                     │
│              ┌───────▼────────┐                            │
│              │    Global      │                            │
│              │  (Communities) │                            │
│              └────────────────┘                            │
│                                                              │
├─────────────────────────────────────────────────────────────┤
│                      Storage Layer                           │
│  ┌─────────────┐         ┌──────────────┐                  │
│  │Vector Store │         │Knowledge     │                  │
│  │             │         │Graph         │                  │
│  └─────────────┘         └──────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

## Comparison with Other RAG Implementations

### Vector RAG

- **Vector RAG**: Simple vector similarity search
- **LightRAG**: Adds knowledge graph for entity relationships

### GraphRAG

- **GraphRAG**: Focus on entity-relationship graphs
- **LightRAG**: Combines vector + graph with community detection

### Traditional RAG

- **Traditional RAG**: Single retrieval method
- **LightRAG**: Multiple retrieval modes with intelligent fusion

## References

- [LightRAG Paper](https://github.com/HKUDS/LightRAG)
- [LangGraphGo Documentation](../README.md)
- [RAG Examples](../examples/)

## License

Part of LangGraphGo project. See main project LICENSE file.
