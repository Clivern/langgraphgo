# Milvus 向量存储示例

本示例演示如何使用 [Milvus](https://milvus.io/) 作为 LangGraphGo RAG 管道的向量存储。

## 什么是 Milvus？

Milvus 是一个开源向量数据库，专为支持嵌入相似度搜索和 AI 应用而构建。它具有以下特点：

- **十亿级向量索引** - 处理大规模向量数据集
- **实时搜索性能** - 亚毫秒级延迟
- **多种索引类型** - HNSW、IVF、Flat 等
- **灵活部署** - 单机、分布式或云原生
- **丰富功能** - 分区、副本、标量过滤

## 功能演示

本示例展示了：

1. **创建 Milvus 向量存储** 并使用自定义配置
2. **添加文档** 包含嵌入向量和元数据
3. **构建 RAG 管道** 使用 Milvus 向量存储
4. **相似度搜索** 进行文档检索
5. **直接使用 Milvus API** 进行高级操作
6. **多语言支持** 通过 Milvus Go SDK v2

## 前置要求

### 启动 Milvus 服务器

本地开发环境可使用 Docker 启动 Milvus：

```bash
docker run -d \
  --name milvus-standalone \
  -p 19530:19530 \
  -v milvus:/var/lib/milvus \
  milvusdb/milvus:latest
```

或使用 Docker Compose 进行完整部署：

```bash
# 下载 docker-compose.yml
wget https://github.com/milvus-io/milvus/releases/download/v2.4.0/milvus-standalone-docker-compose.yml -O docker-compose.yml

# 启动 Milvus
docker-compose up -d
```

测试的时候，使用下面的命令启动 Milvus-lite：

```
pip install milvus
milvus-server
```

### 安装依赖

```bash
go get github.com/tmc/langchaingo/vectorstores/milvus/v2
```

## 运行示例

```bash
cd examples/rag_milvus_example
go run main.go
```

连接到远程 Milvus 实例：

```bash
MILVUS_ADDRESS=your-milvus-server:19530 go run main.go
```

## 代码概览

### 初始化 Milvus 客户端

```go
milvusConfig := client.Config{
    Address: "localhost:19530",
}

store, err := milvusv2.New(
    ctx,
    milvusConfig,
    milvusv2.WithEmbedder(embedder),
    milvusv2.WithCollectionName("my_documents"),
    milvusv2.WithIndex(entity.NewFlatIndex(entity.COSINE)),
    milvusv2.WithMetricType(entity.COSINE),
)
```

### 配置选项

```go
// 集合选项
milvusv2.WithCollectionName("name")     // 集合名称
milvusv2.WithPartitionName("partition") // 多租户分区
milvusv2.WithDropOld()                  // 删除现有集合

// 索引选项
milvusv2.WithIndex(entity.NewHNSWIndex(entity.COSINE, 16, 200))
milvusv2.WithMetricType(entity.COSINE)  // L2, IP, COSINE, HAMMING, JACCARD

// 性能选项
milvusv2.WithShards(2)                  // 分片数量
milvusv2.WithMaxTextLength(1000)        // 文本字段最大长度
milvusv2.WithSkipFlushOnWrite()         // 跳过立即刷新
```

### 索引类型

```go
// 自动索引（推荐）
entity.NewAutoIndex(entity.COSINE)

// 平面索引（精确搜索）
entity.NewFlatIndex(entity.COSINE)

// IVF 索引（平衡）
entity.NewIvfFlatIndex(entity.COSINE, 128)

// HNSW 索引（高召回率，快速）
entity.NewHNSWIndex(entity.COSINE, 16, 200)
```

### 添加文档

```go
docs := []schema.Document{
    {
        PageContent: "你的文档内容",
        Metadata: map[string]any{
            "category": "tech",
            "source": "doc1",
        },
    },
}

ids, err := store.AddDocuments(ctx, docs)
```

### 搜索文档

```go
// 相似度搜索
results, err := store.SimilaritySearch(ctx, "查询文本", 5)

// 带选项的搜索
results, err := store.SimilaritySearch(
    ctx,
    "查询文本",
    5,
    []vectorstores.Option{
        vectorstores.WithScoreThreshold(0.8),
    },
)
```

### 创建 RAG 管道

```go
// 用 LangGraphGo 适配器包装
langGraphStore := rag.NewLangChainVectorStore(store)

// 创建检索器
retriever := retriever.NewVectorStoreRetriever(langGraphStore, embedder, 2)

// 构建管道
config := rag.DefaultPipelineConfig()
config.Retriever = retriever
config.LLM = llm

pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()

runnable, _ := pipeline.Compile()
```

## 索引选择指南

### HNSW 索引
- **使用场景**：高召回率，快速搜索
- **参数**：M（最大连接数），efConstruction（构建时间）
- **内存占用**：较高
- **最适用**：有严格延迟要求的实时应用

```go
entity.NewHNSWIndex(entity.COSINE, 16, 200)
```

### IVF 索引
- **使用场景**：性能和内存平衡
- **参数**：nlist（聚类数量）
- **内存占用**：中等
- **最适用**：有内存限制的大规模数据集

```go
entity.NewIvfFlatIndex(entity.COSINE, 128)
```

### 平面索引
- **使用场景**：精确搜索，小数据集
- **参数**：无
- **内存占用**：低
- **最适用**：小数据集，精确匹配需求

```go
entity.NewFlatIndex(entity.COSINE)
```

### 自动索引
- **使用场景**：让 Milvus 自动选择
- **参数**：无
- **内存占用**：自动优化
- **最适用**：快速原型开发，动态工作负载

```go
entity.NewAutoIndex(entity.COSINE)
```

## 距离度量

```go
// L2 距离（欧几里得）
entity.L2

// 内积
entity.IP

// 余弦相似度
entity.COSINE

// 汉明距离（用于二进制向量）
entity.HAMMING

// Jaccard 距离（用于二进制向量）
entity.JACCARD
```

## 高级功能

### 分区实现多租户

```go
// 使用分区创建存储
store, err := milvusv2.New(
    ctx,
    milvusConfig,
    milvusv2.WithPartitionName("tenant_123"),
)

// 每个租户拥有隔离的数据
// 仅在租户的分区内搜索
```

### 标量过滤

```go
// 结合向量搜索和元数据过滤
results, err := store.SimilaritySearch(
    ctx,
    "查询",
    5,
    []vectorstores.Option{
        vectorstores.WithFilters(map[string]interface{}{
            "category": "tech",
            "year":    2024,
        }),
    },
)
```

### 副本实现扩展

```bash
# 通过 Milvus API 或 UI 创建副本
# 允许扩展读取操作
```

## 生产部署

### 单机部署

适用于：
- 开发和测试
- 小到中型数据集（< 1000 万向量）
- 单服务器部署

```bash
docker run -d --name milvus-standalone \
  -p 19530:19530 \
  -v milvus:/var/lib/milvus \
  milvusdb/milvus:latest
```

### 集群部署

适用于：
- 生产工作负载
- 大型数据集（> 1000 万向量）
- 高可用性要求

```bash
# 在 Kubernetes 上使用 Milvus Operator
kubectl apply -f https://github.com/milvus-io/milvus/releases/download/v2.4.0/milvus-operator.yaml
```

### 云服务

- [Zilliz Cloud](https://zilliz.com/) - 全托管的 Milvus
- [AWS Marketplace](https://aws.amazon.com/marketplace)
- [Google Cloud Marketplace](https://cloud.google.com/marketplace)

## 性能调优

### 索引调优

```go
// 高召回率（95%+）
entity.NewHNSWIndex(entity.COSINE, 32, 200)

// 平衡性能
entity.NewIvfFlatIndex(entity.COSINE, 256)

// 快速插入
entity.NewFlatIndex(entity.COSINE)
```

### 搜索参数

```go
// 调整搜索时间与准确度的平衡
options := []vectorstores.Option{
    vectorstores.WithScoreThreshold(0.7),    // 过滤低分结果
    vectorstores.WithTopK(10),                // 获取更多结果
}
```

### 分片策略

```go
// 跨分片分布数据
milvusv2.WithShards(4)  // 4 个分片并行处理
```

## 与其他向量存储的对比

| 特性 | Milvus | Pinecone | Weaviate | sqlite-vec |
|---------|--------|----------|----------|------------|
| 自托管 | ✅ 是 | ❌ 否 | ✅ 是 | ✅ 是 |
| 云托管 | ✅ 是 | ✅ 是 | ✅ 是 | ❌ 否 |
| 十亿级规模 | ✅ 是 | ✅ 是 | ✅ 是 | ❌ 否 |
| 分区 | ✅ 是 | ✅ 是 | ✅ 是 | ❌ 否 |
| 副本 | ✅ 是 | ✅ 是 | ✅ 是 | ❌ 否 |
| 复杂度 | ⭐⭐⭐ 高 | ⭐ 低 | ⭐⭐ 中 | ⭐ 低 |

## 故障排查

### 连接问题

```bash
# 检查 Milvus 是否运行
docker ps | grep milvus

# 查看 Milvus 日志
docker logs milvus-standalone

# 测试连接
telnet localhost 19530
```

### 集合已存在

```go
// 使用 WithDropOld() 删除现有集合
milvusv2.WithDropOld()
```

### 内存问题

```go
// 减少分片数
milvusv2.WithShards(1)

// 使用更节省内存的索引
entity.NewIvfFlatIndex(entity.COSINE, 64)
```

## 最佳实践

1. **选择正确的索引** - 从 AutoIndex 开始，然后优化
2. **使用分区** - 实现多租户并提高性能
3. **设置适当的度量** - 对归一化嵌入使用 COSINE
4. **监控性能** - 使用 Milvus 指标和监控
5. **定期备份** - 备份集合数据和元数据
6. **架构设计** - 提前规划架构和索引策略

## 技术细节

### Milvus 集合架构

Milvus 集合包含：
- **主键**：唯一标识每个向量
- **向量字段**：存储嵌入向量
- **标量字段**：存储元数据（文本、数字等）
- **索引**：加速向量搜索

### 数据插入流程

```
文档 → 嵌入生成 → 向量插入 → 段创建 → 索引构建 → 刷新
```

### 搜索流程

```
查询 → 嵌入生成 → 向量搜索 → 粗略召回 → 精确排序 → 结果返回
```

## 参考资源

- [Milvus 文档](https://milvus.io/docs)
- [Milvus Go SDK](https://github.com/milvus-io/milvus-sdk-go)
- [RAG 管道文档](../../rag/README.md)
- [LangChain Milvus 集成](https://github.com/tmc/langchaingo/tree/main/vectorstores/milvus/v2)
- [其他向量存储示例](../rag_sqlitevec_example/)

## 许可证

本示例遵循 LangGraphGo 项目的许可证。
