# LightRAG 简单示例

这是一个展示 LightRAG 四种检索模式的简单示例。

## 概述

本示例展示如何使用 LightRAG 的不同检索模式：
- **Naive（朴素）**：简单的向量相似度搜索
- **Local（本地）**：基于实体的检索，带图谱遍历
- **Global（全局）**：社区级别的检索
- **Hybrid（混合）**：结合本地和全局检索

## 前置要求

```bash
go mod tidy
```

## 运行示例

### 不使用 OpenAI API Key（使用 Mock LLM）

本示例包含 Mock LLM 用于演示：

```bash
go run main.go
```

### 使用 OpenAI API Key

生产环境使用时，设置您的 OpenAI API key：

```bash
export OPENAI_API_KEY="your-api-key-here"
go run main.go
```

## 示例功能

1. **创建 LightRAG 引擎**，使用混合模式配置
2. **添加示例文档**，包含 LangGraph、LightRAG、知识图谱、向量数据库和 RAG 相关内容
3. **测试每种检索模式**，使用三个不同查询：
   - "LightRAG 是什么，它是如何工作的？"
   - "解释 RAG 和知识图谱之间的关系"
   - "使用向量数据库有什么好处？"
4. **显示结果**，包括检索到的源、置信度分数和响应时间

## 预期输出

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

## 配置说明

示例使用以下配置：

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

## 代码结构

- `main()`: 设置 LightRAG 引擎并运行查询
- `OpenAILLMAdapter`: 包装 OpenAI LLM 以实现 `rag.LLMInterface`
- `MockLLM`: Mock 实现，用于无 API key 时的演示

## 相关文档

- [LightRAG 高级示例](../lightrag_advanced/) - 更全面的示例，包含性能对比
- [LightRAG 文档](../../docs/lightrag.md) - 完整文档
