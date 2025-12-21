# 使用 Chroma VectorStore 的 RAG 示例

本示例演示如何通过 langchaingo 适配器将 **Chroma 向量数据库**与 **LangGraphGo 的 RAG 管道**结合使用。

## 什么是 Chroma？

Chroma 是一个开源的嵌入数据库，旨在简化 LLM 应用的构建。它提供：
- 简单的 API 用于存储和查询嵌入
- 多种距离度量（余弦、L2、内积）
- 过滤和元数据支持
- 持久化存储
- 客户端-服务器架构

## 前置要求

1. **Chroma 服务器**：使用 Docker 运行 Chroma
   ```bash
   docker run -p 8000:8000 chromadb/chroma
   ```

2. **API 密钥**：DeepSeek 或 OpenAI
   ```bash
   export DEEPSEEK_API_KEY="your-api-key"
   # 或
   export OPENAI_API_KEY="your-api-key"
   ```

## 运行示例

```bash
cd examples/rag_chroma_example
go run main.go
```

## 功能演示

- 使用自定义配置创建 Chroma 向量存储
- 将带有嵌入的文档添加到 Chroma
- 使用 Chroma 作为向量存储构建 RAG 管道
- 使用自然语言问题查询管道
- 带相关性分数的相似度搜索

## 代码要点

### 创建 Chroma 存储

```go
chromaStore, err := chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),
    chroma.WithEmbedder(embedder),
    chroma.WithDistanceFunction("cosine"),
    chroma.WithNameSpace("langgraphgo_example"),
)

// 使用适配器封装
vectorStore := rag.NewLangChainVectorStore(chromaStore)
```

### 添加文档

```go
// 适配器/存储内部处理嵌入
err = vectorStore.Add(ctx, chunks)
```

### 相似度搜索

```go
// 使用检索器进行搜索
retriever := rag.NewLangChainRetriever(chromaStore, 3)
results, err := retriever.RetrieveWithConfig(ctx, query, &rag.RetrievalConfig{K: 5})

for _, result := range results {
    fmt.Printf("分数: %.4f - %s\n", result.Score, result.Document.Content)
}
```

## 预期输出

```
=== 使用 Chroma VectorStore 的 RAG 示例 ===

加载和分割文档...
分割为 6 个块

连接到 Chroma 向量数据库...
添加文档到 Chroma...
文档成功添加到 Chroma

构建 RAG 管道...
管道就绪！

================================================================================
查询 1：什么是 Go 编程语言？
--------------------------------------------------------------------------------

从 Chroma 检索到 3 个文档：
  [1] Go 是一种静态类型的编译型编程语言，由 Google 设计...
  [2] Go 的主要特性包括：- 快速编译时间 - 内置并发支持...
  [3] Go 特别适合：- Web 服务器和 API - 云和网络服务...

回答：
Go 是一种由 Google 设计的静态类型编译型编程语言...
```

## Chroma 配置选项

```go
chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),     // Chroma 服务器 URL
    chroma.WithEmbedder(embedder),                      // 嵌入函数
    chroma.WithDistanceFunction("cosine"),              // 距离度量：cosine, l2, ip
    chroma.WithNameSpace("my_collection"),              // 集合名称
)
```

## 故障排除

### Chroma 服务器未运行
```
错误：Failed to create Chroma store: connection refused
```
**解决方案**：使用 Docker 启动 Chroma 服务器：
```bash
docker run -p 8000:8000 chromadb/chroma
```

### 端口已被占用
```
错误：bind: address already in use
```
**解决方案**：使用不同的端口或停止冲突的服务：
```bash
docker run -p 8001:8000 chromadb/chroma
# 更新代码：chroma.WithChromaURL("http://localhost:8001")
```

## 下一步

- 尝试其他向量存储：Weaviate、Pinecone、Qdrant
- 实验不同的距离函数
- 为搜索添加元数据过滤
- 实现混合搜索（关键词 + 向量）

## 参考资料

- [Chroma 文档](https://docs.trychroma.com/)
- [LangChain Go Chroma 集成](https://github.com/tmc/langchaingo/tree/main/vectorstores/chroma)
- [LangGraphGo RAG 文档](../../docs/RAG/RAG_CN.md)
