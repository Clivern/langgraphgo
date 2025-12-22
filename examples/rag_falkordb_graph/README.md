# RAG with FalkorDB Graph Database

This example demonstrates how to use FalkorDB as a knowledge graph backend for RAG (Retrieval-Augmented Generation) systems. FalkorDB is a Redis module that provides graph database capabilities, allowing you to store and query entities and their relationships.

## Overview

This example showcases:

1. **Automatic Entity and Relationship Extraction**: Using LLM to extract entities and relationships from documents
2. **GraphRAG Engine**: Using the GraphRAG engine for entity-based retrieval
3. **Knowledge Graph Construction**: Building a complete knowledge graph from text documents
4. **Entity Exploration**: Querying and exploring entities and relationships in the graph
5. **Advanced Graph Queries**: Performing complex graph-based queries

## Prerequisites

1. **FalkorDB Server**: You need a running FalkorDB instance
   ```bash
   # Using Docker
   docker run -p 6379:6379 falkordb/falkordb

   # Or install FalkorDB as a Redis module
   # See: https://docs.falkordb.com/docs/quick-start/
   ```

2. **Go Dependencies**: The example requires the following Go modules:
   ```bash
   go get github.com/redis/go-redis/v9
   go get github.com/tmc/langchaingo/llms/openai
   ```

3. **OpenAI API Key**: Set your OpenAI API key
   ```bash
   export OPENAI_API_KEY=your-api-key-here
   ```

## Running the Example

```bash
cd examples/rag_falkordb_graph
go run main.go
```

## Architecture

### GraphRAG Pipeline

```
Documents
    ↓
Entity Extraction (LLM)
    ↓
Relationship Extraction (LLM)
    ↓
Knowledge Graph (FalkorDB)
    ↓
GraphRAG Engine
    ↓
Query Processing & Answer Generation
```

### Key Components

#### 1. Entity Extraction

The system uses LLM to extract entities from documents:

```go
graphRAGConfig := rag.GraphRAGConfig{
    EntityTypes: []string{
        "PERSON",
        "ORGANIZATION",
        "LOCATION",
        "PRODUCT",
        "TECHNOLOGY",
        "CONCEPT",
    },
}
```

#### 2. Relationship Extraction

Automatically identifies relationships between extracted entities:

- **WORKS_AT**: Person → Organization
- **FOUNDED_BY**: Organization → Person
- **PRODUCES**: Organization → Product
- **BASED_IN**: Organization → Location

#### 3. Knowledge Graph Storage

All entities and relationships are stored in FalkorDB:

```go
// Connection string format: falkordb://host:port/graph_name
kg, err := store.NewFalkorDBGraph("falkordb://localhost:6379/rag_graph")

// Add entities
kg.AddEntity(ctx, &rag.Entity{
    ID:   "apple_inc",
    Name: "Apple Inc.",
    Type: "ORGANIZATION",
    Properties: map[string]any{
        "industry": "Technology",
        "founded": "1976",
    },
})

// Add relationships
kg.AddRelationship(ctx, &rag.Relationship{
    ID:     "jobs_founded_apple",
    Source: "steve_jobs",
    Target: "apple_inc",
    Type:   "FOUNDED_BY",
})
```

## Sample Documents

The example processes documents about major technology companies:

- **Apple Inc.**: Steve Jobs, iPhone, iPad, Mac, iOS
- **Microsoft**: Bill Gates, Windows, Office, Azure
- **Google**: Larry Page, Android, Chrome, Cloud Platform
- **Tesla**: Elon Musk, Model S, Model Y, Supercharger
- **Amazon**: Jeff Bezos, AWS, Kindle, e-commerce

## Query Examples

### 1. Entity-based Queries

```go
queries := []string{
    "What products does Apple make?",
    "Who founded Microsoft and what are their main products?",
    "Tell me about electric vehicle companies and their founders",
}
```

### 2. Graph Traversal

The system can traverse relationships to find connected information:

```go
// Find entities related to Apple Inc.
relatedEntities, err := kg.GetRelatedEntities(ctx, "apple_inc", 2)
```

### 3. Complex Graph Queries

```go
graphQuery := &rag.GraphQuery{
    EntityTypes: []string{"PERSON", "ORGANIZATION"},
    Limit:       10,
}
result, err := kg.Query(ctx, graphQuery)
```

## Performance Considerations

### Processing Speed

- **Entity Extraction**: ~2-3 seconds per document (LLM-dependent)
- **Relationship Extraction**: ~1-2 seconds per document
- **Graph Queries**: Milliseconds for simple queries

### Optimization Strategies

1. **Batch Processing**: Process multiple documents together
2. **Caching**: Cache extracted entities and relationships
3. **Parallel Extraction**: Extract from multiple documents concurrently
4. **Hybrid Approach**: Use manual definitions for known entities, LLM for new content

## Features Demonstrated

### 1. Automatic Knowledge Graph Construction

```go
// Each document is automatically processed:
documents := []rag.Document{
    {
        Content: "Apple Inc. is a technology company...",
        Metadata: map[string]any{
            "source": "apple_overview.txt",
            "topic":  "Apple Inc.",
        },
    },
}

// Add to knowledge graph
err := graphEngine.AddDocuments(ctx, documents)
```

### 2. Entity and Relationship Management

- **Entity Types**: PERSON, ORGANIZATION, LOCATION, PRODUCT, TECHNOLOGY, CONCEPT
- **Relationship Types**: WORKS_AT, FOUNDED_BY, BASED_IN, PRODUCES, COMPETES_WITH
- **Entity Properties**: Custom properties for rich entity descriptions

### 3. Graph-based Retrieval

```go
// Query with graph context
result, err := graphEngine.Query(ctx, query)
if err == nil {
    fmt.Printf("Found %d entities and %d relationships\n",
        len(result.Entities), len(result.Relationships))
}
```

### 4. Entity Exploration

```go
// Find related entities
relatedEntities, err := kg.GetRelatedEntities(ctx, "apple_inc", 2)

// Query specific entity types
graphQuery := &rag.GraphQuery{
    EntityTypes: []string{"PERSON"},
    Limit:       10,
}
```

## Use Cases

This example is ideal for:

1. **Document Analysis**: Automatically extract knowledge from large document collections
2. **Knowledge Management**: Build searchable knowledge graphs from unstructured text
3. **Research Applications**: Discover hidden relationships in research papers
4. **Enterprise Knowledge Bases**: Transform internal documents into queryable graphs
5. **Question Answering**: Provide context-aware answers based on graph relationships

## Advanced Features

### 1. Custom Entity Extraction Prompts

```go
graphRAGConfig.ExtractionPrompt = `
Extract entities from the following text. Focus on these entity types: %s.
Return a JSON response with this structure:
{
  "entities": [
    {
      "name": "entity_name",
      "type": "entity_type",
      "description": "brief description",
      "properties": {}
    }
  ]
}

Text: %s`
```

### 2. Relationship Detection

The system automatically detects various relationship types:
- **Employment**: WORKS_AT, CEO_OF
- **Creation**: FOUNDED_BY, CREATED_BY
- **Location**: BASED_IN, LOCATED_IN
- **Products**: PRODUCES, MANUFACTURES
- **Competition**: COMPETES_WITH, RIVAL_OF

### 3. Graph Visualization

The example includes Mermaid diagram visualization:

```go
exporter := graph.NewExporter(pipeline.GetGraph())
fmt.Println(exporter.DrawMermaid())
```

## Troubleshooting

### Common Issues

1. **Connection Failed**: Ensure FalkorDB is running and accessible
2. **Entity Extraction Fails**: Check OpenAI API key and network connectivity
3. **Slow Performance**: Consider batch processing or manual entity definitions
4. **Memory Usage**: Monitor Redis memory usage with large graphs

### Debug Mode

Enable debug output to see the internal processing:

```go
// Check what entities were extracted
fmt.Printf("Extracted %d entities\n", len(entities))
for _, entity := range entities {
    fmt.Printf("- %s (%s)\n", entity.Name, entity.Type)
}
```

## Extensions

### 1. Hybrid Approach

Combine automatic extraction with manual definitions:

```go
// Add known entities manually
knownEntities := preloadKnownEntities()

// Extract new entities from documents
extractedEntities := extractFromDocuments(ctx, documents)

// Merge and add to knowledge graph
allEntities := append(knownEntities, extractedEntities...)
```

### 2. Custom Relationship Types

Define your own relationship types for specific domains:

```go
relationshipTypes := []string{
    "PARTNERS_WITH",
    "ACQUIRES",
    "INVESTS_IN",
    "COLLABORATES_ON",
}
```

### 3. Enrichment Pipeline

Add external data sources to enrich entities:

```go
// Enrich entities with external APIs
for _, entity := range entities {
    if entity.Type == "ORGANIZATION" {
        externalInfo := fetchCompanyData(entity.Name)
        mergeProperties(entity.Properties, externalInfo)
    }
}
```

## Best Practices

1. **Start Small**: Begin with a few documents and scale up
2. **Validate Extractions**: Review extracted entities for accuracy
3. **Optimize Prompts**: Customize extraction prompts for your domain
4. **Monitor Performance**: Track processing times and optimize bottlenecks
5. **Regular Updates**: Periodically refresh the knowledge graph with new documents

## Integration

### With Existing RAG Systems

This FalkorDB integration can be combined with traditional vector-based RAG:

```go
// Hybrid retriever combining vector and graph search
vectorRetriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, k)
graphRetriever := retriever.NewKnowledgeGraphRetriever(kg, k)
hybridRetriever := retriever.NewHybridRetriever(vectorRetriever, graphRetriever, 0.5)
```

### With LangChain

```go
// Use LangChain components
embedder := rag.NewLangChainEmbedder(openaiEmbedder)
vectorStore := rag.NewLangChainVectorStore(chromaStore)
```

## Next Steps

1. **Try Different Data**: Replace with your own domain-specific documents
2. **Custom Entity Types**: Add entity types relevant to your domain
3. **Batch Processing**: Implement automated document processing pipelines
4. **Integration**: Connect to your existing RAG applications
5. **Monitoring**: Add metrics and monitoring for production use

## Contributing

Contributions to improve the FalkorDB integration are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add your improvements
4. Submit a pull request

## License

This example is part of the LangGraphGo project. See the main repository for license information.