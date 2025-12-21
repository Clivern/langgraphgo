# RAG with Query Rewriting

This example demonstrates how to implement a Retrieval-Augmented Generation (RAG) pipeline that includes a **Query Rewriting** step.

## Overview

In many RAG applications, the user's initial query might be too vague or not optimized for semantic search. By using an LLM to rewrite the query *before* retrieval, we can significantly improve the relevance of the retrieved documents.

This pipeline consists of the following steps:
1.  **Rewrite Query**: An LLM transforms the user's natural language query into a more specific, search-optimized query.
2.  **Retrieve**: The system retrieves documents using the *rewritten* query.
3.  **Generate**: The LLM generates a final answer using the original query (for context) and the retrieved documents.

## How to Run

```bash
export OPENAI_API_KEY=your_key_here
go run main.go
```

## Implementation Details

- Uses `github.com/smallnest/langgraphgo/graph` for the state machine.
- Uses `github.com/smallnest/langgraphgo/rag` for data structures.
- Implements a custom graph with `rewrite_query` -> `retrieve` -> `generate` nodes.
