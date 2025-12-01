# LangChain VectorStores 集成

## 概述

本文档总结了将 `github.com/tmc/langchaingo/vectorstores` 集成到 LangGraphGo 的工作，为 RAG 应用提供了对多个向量数据库后端的无缝访问。

## 新增内容

### 1. LangChainVectorStore 适配器 (`prebuilt/rag_langchain_adapter.go`)

一个新的适配器，封装任何 `langchaingo` vectorstore 实现以与 LangGraphGo 的 RAG 管道配合使用：

```go
type LangChainVectorStore struct {
    store vectorstores.VectorStore
}

func NewLangChainVectorStore(store vectorstores.VectorStore) *LangChainVectorStore
```

**方法**：
- `AddDocuments(ctx, documents, embeddings)` - 向向量存储添加文档
- `SimilaritySearch(ctx, query, k)` - 搜索相似文档
- `SimilaritySearchWithScore(ctx, query, k)` - 带相关性分数的搜索

### 2. 测试套件 (`prebuilt/rag_langchain_vectorstore_test.go`)

全面的测试确保适配器正常工作：
- 文档添加测试
- 相似度搜索测试
- 基于分数的搜索测试
- 集成测试

### 3. 示例

#### 示例 1：通用 VectorStore 集成
**位置**：`examples/rag_langchain_vectorstore_example/`

演示：
- 使用内存向量存储
- 可选的 Weaviate 集成
- 使用 LangChain 组件的完整 RAG 管道
- 带分数的相似度搜索

**文件**：
- `main.go` - 完整的工作示例
- `README.md` - 英文文档
- `README_CN.md` - 中文文档

#### 示例 2：Chroma 集成
**位置**：`examples/rag_chroma_example/`

演示：
- Chroma 向量数据库设置
- 使用 Chroma 进行文档索引
- 使用 Chroma 后端的 RAG 管道
- 生产就绪的配置

**文件**：
- `main.go` - Chroma 特定示例
- `README.md` - 英文文档及设置指南
- `README_CN.md` - 中文文档

### 4. 文档更新

更新了 `docs/RAG/RAG.md`，包含：
- 全面的 LangChain 集成部分
- 适配器使用示例
- 支持的向量存储后端
- Chroma、Weaviate、Pinecone 的设置指南
- 完整的集成示例
- 优势和最佳实践

## 支持的向量存储

适配器支持**任何** langchaingo vectorstore 实现，包括：

| 向量存储     | 类型            | 最适合                   |
| ------------ | --------------- | ------------------------ |
| **Chroma**   | 开源            | 开发、中小规模           |
| **Weaviate** | 开源/云         | 生产、可扩展性           |
| **Pinecone** | 托管服务        | 易用性、托管基础设施     |
| **Qdrant**   | 开源/云         | 高性能、过滤             |
| **Milvus**   | 开源/云         | 大规模、分布式           |
| **PGVector** | PostgreSQL 扩展 | 现有 PostgreSQL 基础设施 |

## 使用模式

### 基本使用

```go
// 1. 创建 langchaingo 向量存储
chromaStore, err := chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),
    chroma.WithEmbedder(embedder),
)

// 2. 使用适配器封装
vectorStore := prebuilt.NewLangChainVectorStore(chromaStore)

// 3. 在 RAG 管道中使用
retriever := prebuilt.NewVectorStoreRetriever(vectorStore, 3)

config := prebuilt.DefaultRAGConfig()
config.Retriever = retriever
config.LLM = llm

pipeline := prebuilt.NewRAGPipeline(config)
pipeline.BuildBasicRAG()
```

### 完整集成

```go
// 所有 LangChain 组件一起使用
loader := prebuilt.NewLangChainDocumentLoader(documentloaders.NewText(reader))
splitter := prebuilt.NewLangChainTextSplitter(textsplitter.NewRecursiveCharacter(...))
embedder := prebuilt.NewLangChainEmbedder(embeddings.NewEmbedder(llm))
vectorStore := prebuilt.NewLangChainVectorStore(chroma.New(...))

// 构建 RAG 管道
chunks, _ := loader.LoadAndSplit(ctx, splitter)
embeddings, _ := embedder.EmbedDocuments(ctx, texts)
vectorStore.AddDocuments(ctx, chunks, embeddings)

retriever := prebuilt.NewVectorStoreRetriever(vectorStore, 3)
// ... 继续构建管道
```

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                  LangGraphGo RAG 管道                        │
│                                                              │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐             │
│  │   检索   │───▶│  重排序  │───▶│   生成   │             │
│  └──────────┘    └──────────┘    └──────────┘             │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────┐           │
│  │   LangChainVectorStore 适配器               │           │
│  │   (prebuilt.NewLangChainVectorStore)        │           │
│  └─────────────────────────────────────────────┘           │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────┐           │
│  │   langchaingo vectorstores.VectorStore      │           │
│  │   - Chroma                                   │           │
│  │   - Weaviate                                 │           │
│  │   - Pinecone                                 │           │
│  │   - Qdrant                                   │           │
│  │   - Milvus                                   │           │
│  │   - PGVector                                 │           │
│  └─────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## 优势

1. **生态系统访问**：直接访问整个 langchaingo vectorstore 生态系统
2. **生产就绪**：在生产环境中使用经过实战检验的向量数据库
3. **灵活性**：轻松在不同向量存储之间切换
4. **无供应商锁定**：标准接口适用于任何后端
5. **面向未来**：自动兼容新的 langchaingo vectorstore 实现

## 测试

运行测试：

```bash
# 测试适配器
go test ./prebuilt -run TestLangChainVectorStore

# 运行所有 RAG 测试
go test ./prebuilt -run RAG
```

## 示例

### 运行通用示例（内存）：
```bash
cd examples/rag_langchain_vectorstore_example
export DEEPSEEK_API_KEY="your-key"
go run main.go
```

### 运行 Chroma 示例：
```bash
# 启动 Chroma
docker run -p 8000:8000 chromadb/chroma

# 运行示例
cd examples/rag_chroma_example
export DEEPSEEK_API_KEY="your-key"
go run main.go
```

## 迁移指南

### 从 InMemoryVectorStore 迁移到 Chroma

**之前**：
```go
embedder := prebuilt.NewMockEmbedder(128)
vectorStore := prebuilt.NewInMemoryVectorStore(embedder)
```

**之后**：
```go
// 创建 LangChain embedder
lcEmbedder, _ := embeddings.NewEmbedder(llm)
embedder := prebuilt.NewLangChainEmbedder(lcEmbedder)

// 创建 Chroma 存储
chromaStore, _ := chroma.New(
    chroma.WithChromaURL("http://localhost:8000"),
    chroma.WithEmbedder(lcEmbedder),
)
vectorStore := prebuilt.NewLangChainVectorStore(chromaStore)
```

其余的 RAG 管道代码保持不变！

## 下一步

1. **尝试不同后端**：实验 Weaviate、Pinecone、Qdrant
2. **生产部署**：设置托管向量数据库
3. **性能调优**：优化块大小、top-k 和距离度量
4. **混合搜索**：结合向量搜索和关键词搜索
5. **监控**：为生产使用添加指标和日志

## 参考资料

- [LangChain Go 文档](https://github.com/tmc/langchaingo)
- [LangGraphGo RAG 文档](../docs/RAG/RAG_CN.md)
- [Chroma 文档](https://docs.trychroma.com/)
- [Weaviate 文档](https://weaviate.io/developers/weaviate)
- [Pinecone 文档](https://docs.pinecone.io/)

## 贡献

要添加对新向量存储的支持：

1. 确保它实现了 langchaingo 的 `vectorstores.VectorStore` 接口
2. 在 `examples/rag_<vectorstore>_example/` 中创建示例
3. 在文档中添加设置说明
4. 提交带有测试的 PR

适配器自动支持任何 langchaingo vectorstore，因此不需要代码更改！
