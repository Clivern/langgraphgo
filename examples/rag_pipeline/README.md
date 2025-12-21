# RAG Pipeline Example

This example demonstrates how to build a Retrieval-Augmented Generation (RAG) pipeline using LangGraphGo.

## Overview

This example demonstrates how to use the `RAGPipeline` component to:
- **Build Knowledge Base**: Batch process and ingest local text documents into a vector store.
- **Intelligent Q&A**: Use a compiled RAG graph to answer questions based on the ingested documents.

The example uses a modular architecture:
1. **Load**: `TextLoader` for local files.
2. **Split**: `RecursiveCharacterTextSplitter` for chunking.
3. **Embed**: LangChain OpenAI embeddings.
4. **Store**: In-memory vector store.
5. **Pipeline**: `RAGPipeline` for orchestrating retrieval and generation.

## Running the Example

Make sure you have `OPENAI_API_KEY` set in your environment.

```bash
cd examples/rag_pipeline
go run main.go
```

## Key Features

- **Document Retrieval**: Vector-based document search
- **Context Integration**: Combine retrieved docs with queries
- **Generation Pipeline**: LLM-based response generation
- **Scalable Architecture**: Handle large document collections