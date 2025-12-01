# RAG 与 LangChain Embeddings 集成示例

本示例演示如何将 LangChain Go 的嵌入模型与 LangGraphGo 的 RAG 系统集成。

## 概述

LangChain Go 提供了来自各种提供商（OpenAI、Cohere、HuggingFace 等）的优秀嵌入模型。本示例展示如何通过简单的适配器无缝地将它们与我们的 RAG 流水线一起使用。

## 主要特性

- **直接集成**: 使用最小包装器使用 LangChain 嵌入
- **多个提供商**: 支持 OpenAI、Cohere 和其他提供商
- **类型转换**: 自动 float32 ↔ float64 转换
- **完整的 RAG 流水线**: 使用真实嵌入的端到端示例
- **相似度比较**: 演示嵌入质量

## 架构

### 适配器类

`prebuilt/rag_langchain_adapter.go` 中的 `LangChainEmbedder` 适配器：

```go
type LangChainEmbedder struct {
    embedder embeddings.Embedder
}
```

**主要特性**:
- 包装任何 LangChain 嵌入器
- 实现我们的 `Embedder` 接口
- 转换 `float32` (LangChain) ↔ `float64` (我们的类型)
- 零开销，简单的传递

## 使用方法

### 基本嵌入

```go
import (
    "github.com/tmc/langchaingo/embeddings"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/smallnest/langgraphgo/prebuilt"
)

// 创建 LangChain 嵌入器
lcEmbedder, _ := embeddings.NewEmbedder(openai.New())

// 使用适配器包装
embedder := prebuilt.NewLangChainEmbedder(lcEmbedder)

// 使用我们的接口
queryEmb, _ := embedder.EmbedQuery(ctx, "什么是 AI？")
docsEmb, _ := embedder.EmbedDocuments(ctx, texts)
```

### 在 RAG 流水线中使用

```go
// 创建嵌入器
lcEmbedder, _ := embeddings.NewEmbedder(openai.New())
embedder := prebuilt.NewLangChainEmbedder(lcEmbedder)

// 使用 LangChain 嵌入创建向量存储
vectorStore := prebuilt.NewInMemoryVectorStore(embedder)

// 生成嵌入
embeds, _ := embedder.EmbedDocuments(ctx, texts)
vectorStore.AddDocuments(ctx, documents, embeds)

// 构建 RAG 流水线
retriever := prebuilt.NewVectorStoreRetriever(vectorStore, 3)
config := prebuilt.DefaultRAGConfig()
config.Retriever = retriever
config.LLM = llm

pipeline := prebuilt.NewRAGPipeline(config)
pipeline.BuildBasicRAG()
```

## 运行示例

### 前提条件

对于 OpenAI 嵌入：
```bash
export OPENAI_API_KEY=your_api_key_here
```

对于 DeepSeek LLM：
```bash
export DEEPSEEK_API_KEY=your_api_key_here
```

### 运行

```bash
cd examples/rag_with_embeddings
go run main.go
```

## 包含的示例

### 1. OpenAI 嵌入
测试 OpenAI 的 text-embedding-ada-002 模型：
```go
openaiEmbedder, _ := embeddings.NewEmbedder(openai.New())
embedder := prebuilt.NewLangChainEmbedder(openaiEmbedder)

queryEmb, _ := embedder.EmbedQuery(ctx, "什么是机器学习？")
// 返回 1536 维嵌入
```

### 2. 完整的 RAG 流水线
使用真实嵌入构建完整的 RAG 系统：
```go
// 如果可用则使用 OpenAI 嵌入，否则使用模拟
vectorStore := prebuilt.NewInMemoryVectorStore(embedder)
embeds, _ := embedder.EmbedDocuments(ctx, texts)
vectorStore.AddDocuments(ctx, documents, embeds)

// 使用语义搜索查询
result, _ := runnable.Invoke(ctx, prebuilt.RAGState{
    Query: "什么是 LangGraph？",
})
```

### 3. 相似度比较
比较嵌入以理解语义相似性：
```go
testTexts := []string{
    "机器学习和人工智能",
    "深度学习神经网络",
    "今天天气晴朗",
}

embeds, _ := embedder.EmbedDocuments(ctx, testTexts)
similarity := cosineSimilarity(embeds[0], embeds[1])
// 相关文本的相似度高，不相关的低
```

## 支持的嵌入提供商

适配器适用于所有 LangChain 嵌入提供商：

### OpenAI
```go
import "github.com/tmc/langchaingo/llms/openai"

lcEmbedder, _ := embeddings.NewEmbedder(openai.New())
embedder := prebuilt.NewLangChainEmbedder(lcEmbedder)
```

**模型**:
- `text-embedding-ada-002` (1536 维) - 默认
- `text-embedding-3-small` (1536 维)
- `text-embedding-3-large` (3072 维)

### Cohere
```go
import "github.com/tmc/langchaingo/llms/cohere"

lcEmbedder, _ := embeddings.NewEmbedder(cohere.New())
embedder := prebuilt.NewLangChainEmbedder(lcEmbedder)
```

### HuggingFace
```go
import "github.com/tmc/langchaingo/llms/huggingface"

lcEmbedder, _ := embeddings.NewEmbedder(huggingface.New())
embedder := prebuilt.NewLangChainEmbedder(lcEmbedder)
```

### Vertex AI
```go
import "github.com/tmc/langchaingo/llms/vertexai"

lcEmbedder, _ := embeddings.NewEmbedder(vertexai.New())
embedder := prebuilt.NewLangChainEmbedder(lcEmbedder)
```

## 类型转换

适配器处理自动类型转换：

### LangChain → 我们的类型
```go
// LangChain 返回 [][]float32
lcEmbeds := [][]float32{{0.1, 0.2, 0.3}}

// 适配器转换为 [][]float64
ourEmbeds := [][]float64{{0.1, 0.2, 0.3}}
```

### 性能
- 转换是 O(n)，其中 n 是总嵌入值
- 对于典型的嵌入大小，开销最小
- 大多数用例不需要内存分配优化

## 嵌入维度

不同的模型有不同的维度：

| 模型                    | 维度 | 提供商 |
| ----------------------- | ---- | ------ |
| text-embedding-ada-002  | 1536 | OpenAI |
| text-embedding-3-small  | 1536 | OpenAI |
| text-embedding-3-large  | 3072 | OpenAI |
| embed-english-v3.0      | 1024 | Cohere |
| embed-multilingual-v3.0 | 1024 | Cohere |

确保您的向量存储配置了正确的维度。

## 最佳实践

### 1. 选择正确的模型
- **OpenAI ada-002**: 质量和成本的良好平衡
- **OpenAI 3-large**: 最高质量，成本更高
- **Cohere**: 适合多语言内容

### 2. 批处理
```go
// 批量处理文档以提高效率
texts := []string{...} // 许多文档
embeds, _ := embedder.EmbedDocuments(ctx, texts)
// LangChain 内部处理批处理
```

### 3. 缓存
```go
// 缓存常用文本的嵌入
cache := make(map[string][]float64)

func getEmbedding(text string) []float64 {
    if emb, ok := cache[text]; ok {
        return emb
    }
    emb, _ := embedder.EmbedQuery(ctx, text)
    cache[text] = emb
    return emb
}
```

### 4. 错误处理
```go
embeds, err := embedder.EmbedDocuments(ctx, texts)
if err != nil {
    // 处理速率限制、网络错误等
    log.Printf("嵌入失败: %v", err)
    // 考虑重试逻辑
}
```

## 对比：模拟 vs 真实嵌入

### 模拟嵌入（开发）
```go
embedder := prebuilt.NewMockEmbedder(1536)
```
- ✅ 快速，无 API 调用
- ✅ 确定性
- ✅ 免费
- ❌ 没有语义意义

### 真实嵌入（生产）
```go
lcEmbedder, _ := embeddings.NewEmbedder(openai.New())
embedder := prebuilt.NewLangChainEmbedder(lcEmbedder)
```
- ✅ 语义有意义
- ✅ 高质量检索
- ✅ 生产就绪
- ❌ 需要 API 密钥和费用

## 故障排除

### API 密钥未设置
```
错误: missing API key
```
**解决方案**: 设置适当的环境变量：
```bash
export OPENAI_API_KEY=your_key
```

### 维度不匹配
```
错误: embedding dimension mismatch
```
**解决方案**: 确保向量存储维度与模型匹配：
```go
// 对于 OpenAI ada-002
vectorStore := prebuilt.NewInMemoryVectorStore(embedder)
// 嵌入器将返回 1536 维向量
```

### 速率限制
```
错误: rate limit exceeded
```
**解决方案**: 实现批处理和重试逻辑：
```go
// LangChain 嵌入器支持批处理
lcEmbedder, _ := embeddings.NewEmbedder(
    openai.New(),
    embeddings.WithBatchSize(100),
)
```

## 性能提示

1. **批量处理文档**: 一次处理多个文档
2. **缓存结果**: 存储嵌入以供重用
3. **使用适当的模型**: 平衡质量与成本
4. **监控使用**: 跟踪 API 调用和成本
5. **实现重试**: 处理瞬时故障

## 下一步

1. 尝试不同的嵌入提供商（Cohere、HuggingFace）
2. 实验不同的模型和维度
3. 使用真实嵌入构建生产 RAG 系统
4. 实现缓存和优化策略
5. 比较不同提供商的嵌入质量

## 另请参阅

- [LangChain Embeddings 文档](https://github.com/tmc/langchaingo)
- [OpenAI Embeddings 指南](https://platform.openai.com/docs/guides/embeddings)
- [RAG 文档](../../docs/RAG/RAG_CN.md)
- [LangChain 集成示例](../rag_with_langchain/)
