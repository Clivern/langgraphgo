# 使用 LangChain VectorStores 的 RAG 示例

本示例演示如何将 **langchaingo vectorstores** 与 **LangGraphGo 的 RAG 管道**集成。展示了如何通过我们的适配器层使用 LangChain 的向量存储实现。

## 功能演示

1. **LangChain VectorStore 集成**：使用 langchaingo 的向量存储实现
2. **文档加载与分割**：使用 LangChain 加载器和分割器加载文档
3. **嵌入生成**：使用 LangChain 嵌入器生成向量
4. **RAG 管道**：构建完整的向量检索 RAG 工作流
5. **多种 VectorStore 后端**：支持内存和外部存储（Weaviate 等）
6. **相似度搜索**：基本搜索和带相关性分数的搜索

## 架构

```
┌─────────────────────────────────────────────────────────────┐
│                    RAG 管道                                  │
│                                                              │
│  ┌──────────┐    ┌──────────┐    ┌──────────┐             │
│  │   检索   │───▶│  重排序  │───▶│   生成   │             │
│  └──────────┘    └──────────┘    └──────────┘             │
│       │                                  │                  │
│       ▼                                  ▼                  │
│  ┌─────────────────────────────────────────────┐           │
│  │     LangChain VectorStore 适配器            │           │
│  │  (封装 langchaingo vectorstores)            │           │
│  └─────────────────────────────────────────────┘           │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────┐           │
│  │   LangChain VectorStore 实现                │           │
│  │   - 内存存储                                 │           │
│  │   - Weaviate                                 │           │
│  │   - Pinecone                                 │           │
│  │   - Chroma                                   │           │
│  │   - Qdrant                                   │           │
│  └─────────────────────────────────────────────┘           │
└─────────────────────────────────────────────────────────────┘
```

## 前置要求

1. **DeepSeek API Key**（或 OpenAI API Key）：
   ```bash
   export DEEPSEEK_API_KEY="your-api-key"
   # 或
   export OPENAI_API_KEY="your-api-key"
   ```

2. **（可选）Weaviate 实例**用于外部向量存储：
   ```bash
   # 使用 Docker 运行 Weaviate
   docker run -d \
     -p 8080:8080 \
     -e AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true \
     -e PERSISTENCE_DATA_PATH=/var/lib/weaviate \
     semitechnologies/weaviate:latest

   # 设置环境变量
   export WEAVIATE_URL="localhost:8080"
   ```

## 运行示例

```bash
cd examples/rag_langchain_vectorstore_example
go run main.go
```

## 代码详解

### 1. 初始化组件

```go
// 创建 LLM
llm, err := openai.New(
    openai.WithModel("deepseek-v3"),
    openai.WithBaseURL("https://api.deepseek.com"),
)

// 创建嵌入器
embedder, err := embeddings.NewEmbedder(llm)
```

### 2. 加载和分割文档

```go
// 使用 LangChain 加载器加载文档
textLoader := documentloaders.NewText(textReader)
loader := prebuilt.NewLangChainDocumentLoader(textLoader)

// 使用 LangChain 分割器分割
splitter := textsplitter.NewRecursiveCharacter(
    textsplitter.WithChunkSize(200),
    textsplitter.WithChunkOverlap(50),
)

chunks, err := loader.LoadAndSplit(ctx, splitter)
```

### 3. 创建向量存储

```go
// 方式 1：内存存储
inMemStore := prebuilt.NewInMemoryVectorStore(
    prebuilt.NewLangChainEmbedder(embedder),
)

// 方式 2：外部存储（Weaviate）
weaviateStore, err := weaviate.New(
    weaviate.WithScheme("http"),
    weaviate.WithHost(weaviateURL),
    weaviate.WithEmbedder(embedder),
)

// 使用适配器封装
wrappedStore := prebuilt.NewLangChainVectorStore(weaviateStore)
```

### 4. 添加文档到向量存储

```go
// 生成嵌入
embeddings, err := embedder.EmbedDocuments(ctx, texts)

// 添加到存储
err = vectorStore.AddDocuments(ctx, chunks, embeddings)
```

### 5. 构建 RAG 管道

```go
// 创建检索器
retriever := prebuilt.NewVectorStoreRetriever(vectorStore, 3)

// 配置管道
config := prebuilt.DefaultRAGConfig()
config.Retriever = retriever
config.LLM = llm
config.IncludeCitations = true

// 构建和编译
pipeline := prebuilt.NewRAGPipeline(config)
pipeline.BuildAdvancedRAG()
runnable, err := pipeline.Compile()
```

### 6. 查询管道

```go
result, err := runnable.Invoke(ctx, prebuilt.RAGState{
    Query: "什么是 LangGraph？",
})

finalState := result.(prebuilt.RAGState)
fmt.Println(finalState.Answer)
```

## 支持的 VectorStore 后端

适配器支持任何 langchaingo vectorstore 实现：

### 内置存储
- **内存存储**：用于测试和开发
- **Weaviate**：开源向量数据库
- **Pinecone**：托管向量数据库
- **Chroma**：嵌入数据库
- **Qdrant**：向量相似度搜索引擎
- **Milvus**：云原生向量数据库

### 使用模式

```go
// 1. 创建 langchaingo vectorstore
store, err := <vectorstore>.New(
    // vectorstore 特定选项
    <vectorstore>.WithEmbedder(embedder),
)

// 2. 使用适配器封装
adaptedStore := prebuilt.NewLangChainVectorStore(store)

// 3. 在 RAG 管道中使用
retriever := prebuilt.NewVectorStoreRetriever(adaptedStore, topK)
```

## 示例输出

```
=== 使用 LangChain VectorStores 的 RAG 示例 ===

示例 1：使用 LangChain 集成的内存 VectorStore
--------------------------------------------------------------------------------
分割为 8 个块
文档成功添加到向量存储

示例 2：使用 LangChain VectorStore 的 RAG 管道
--------------------------------------------------------------------------------
管道可视化：
graph TD
    retrieve --> generate
    generate --> format_citations
    format_citations --> __end__

查询 1：什么是 LangGraph？
检索到 3 个文档：
  [1] LangGraph 是一个用于构建有状态、多角色应用的库，使用 LLM。它扩展了...
  [2] LangGraph 的主要特性包括：- 有状态的图形工作流 - 支持循环...
  [3] LangGraph 支持多种检查点后端，包括：- PostgreSQL 用于生产...

回答：LangGraph 是一个专为使用大型语言模型（LLM）构建有状态、多角色应用而设计的库。
它通过以循环方式协调多个步骤中的多个链来扩展 LangChain 表达式语言...

引用：
  [1] Unknown
  [2] Unknown
  [3] Unknown
```

## 高级功能

### 带分数的相似度搜索

```go
results, err := vectorStore.SimilaritySearchWithScore(ctx, query, k)
for _, result := range results {
    fmt.Printf("分数: %.4f - %s\n", result.Score, result.Document.PageContent)
}
```

### 自定义检索器

```go
type CustomRetriever struct {
    store VectorStore
    // 自定义字段
}

func (r *CustomRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]Document, error) {
    // 自定义检索逻辑
    return r.store.SimilaritySearch(ctx, query, r.topK)
}
```

## 与其他 LangChain 组件集成

本示例展示了 LangGraphGo 如何与 langchaingo 生态系统无缝集成：

- **文档加载器**：Text、CSV、PDF、HTML 等
- **文本分割器**：递归、基于 Token、语义
- **嵌入**：OpenAI、Cohere、HuggingFace
- **向量存储**：Weaviate、Pinecone、Chroma、Qdrant
- **LLM**：OpenAI、Anthropic、Cohere、本地模型

## 下一步

1. 探索其他示例：
   - `rag_with_langchain/` - 基本 LangChain 集成
   - `rag_example/` - 自定义 RAG 实现
   - `rag_advanced_example/` - 高级 RAG 模式

2. 尝试不同的向量存储：
   - 设置 Pinecone、Chroma 或 Qdrant
   - 比较性能和功能

3. 自定义管道：
   - 添加重排序
   - 实现混合搜索
   - 添加查询扩展

## 参考资料

- [LangGraphGo 文档](../../docs/RAG/RAG_CN.md)
- [LangChain Go 文档](https://github.com/tmc/langchaingo)
- [Weaviate 文档](https://weaviate.io/developers/weaviate)
