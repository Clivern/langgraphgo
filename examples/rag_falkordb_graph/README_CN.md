# 使用 FalkorDB 图数据库的 RAG 应用

本示例演示了如何将 FalkorDB 用作 RAG（检索增强生成）系统的知识图谱后端。FalkorDB 是一个 Redis 模块，提供图数据库功能，允许您存储和查询实体及其关系。

## 概述

本示例展示：

1. **自动实体和关系提取**：使用 LLM 从文档中提取实体和关系
2. **GraphRAG 引擎**：使用 GraphRAG 引擎进行基于实体的检索
3. **知识图谱构建**：从文本文档构建完整的知识图谱
4. **实体探索**：查询和探索图中的实体和关系
5. **高级图查询**：执行复杂的基于图的查询

## 前置条件

1. **FalkorDB 服务器**：您需要运行 FalkorDB 实例
   ```bash
   # 使用 Docker
   docker run -p 6379:6379 falkordb/falkordb

   # 或者将 FalkorDB 安装为 Redis 模块
   # 参见：https://docs.falkordb.com/docs/quick-start/
   ```

2. **Go 依赖**：示例需要以下 Go 模块：
   ```bash
   go get github.com/redis/go-redis/v9
   go get github.com/tmc/langchaingo/llms/openai
   ```

3. **OpenAI API 密钥**：设置您的 OpenAI API 密钥
   ```bash
   export OPENAI_API_KEY=your-api-key-here
   ```

## 运行示例

```bash
cd examples/rag_falkordb_graph
go run main.go
```

## 架构

### GraphRAG 流水线

```
文档
    ↓
实体提取 (LLM)
    ↓
关系提取 (LLM)
    ↓
知识图谱 (FalkorDB)
    ↓
GraphRAG 引擎
    ↓
查询处理和答案生成
```

### 核心组件

#### 1. 实体提取

系统使用 LLM 从文档中提取实体：

```go
graphRAGConfig := rag.GraphRAGConfig{
    EntityTypes: []string{
        "PERSON",      // 人员
        "ORGANIZATION", // 组织
        "LOCATION",    // 地点
        "PRODUCT",     // 产品
        "TECHNOLOGY",  // 技术
        "CONCEPT",     // 概念
    },
}
```

#### 2. 关系提取

自动识别提取实体之间的关系：

- **WORKS_AT**：人员 → 组织
- **FOUNDED_BY**：组织 → 人员
- **PRODUCES**：组织 → 产品
- **BASED_IN**：组织 → 地点

#### 3. 知识图谱存储

所有实体和关系都存储在 FalkorDB 中：

```go
// 连接字符串格式：falkordb://host:port/graph_name
kg, err := store.NewFalkorDBGraph("falkordb://localhost:6379/rag_graph")

// 添加实体
kg.AddEntity(ctx, &rag.Entity{
    ID:   "apple_inc",
    Name: "Apple Inc.",
    Type: "ORGANIZATION",
    Properties: map[string]any{
        "industry": "科技",
        "founded": "1976",
    },
})

// 添加关系
kg.AddRelationship(ctx, &rag.Relationship{
    ID:     "jobs_founded_apple",
    Source: "steve_jobs",
    Target: "apple_inc",
    Type:   "FOUNDED_BY",
})
```

## 示例文档

示例处理关于主要科技公司的文档：

- **Apple Inc.**：Steve Jobs、iPhone、iPad、Mac、iOS
- **Microsoft**：Bill Gates、Windows、Office、Azure
- **Google**：Larry Page、Android、Chrome、云平台
- **Tesla**：Elon Musk、Model S、Model Y、超级充电器
- **Amazon**：Jeff Bezos、AWS、Kindle、电子商务

## 查询示例

### 1. 基于实体的查询

```go
queries := []string{
    "Apple 生产什么产品？",
    "谁创立了 Microsoft 以及他们的主要产品是什么？",
    "告诉我关于电动汽车公司和它们创始人的信息",
}
```

### 2. 图遍历

系统可以遍历关系以查找相关信息：

```go
// 查找与 Apple Inc. 相关的实体
relatedEntities, err := kg.GetRelatedEntities(ctx, "apple_inc", 2)
```

### 3. 复杂图查询

```go
graphQuery := &rag.GraphQuery{
    EntityTypes: []string{"PERSON", "ORGANIZATION"},
    Limit:       10,
}
result, err := kg.Query(ctx, graphQuery)
```

## 性能考虑

### 处理速度

- **实体提取**：每文档约 2-3 秒（取决于 LLM）
- **关系提取**：每文档约 1-2 秒
- **图查询**：简单查询毫秒级

### 优化策略

1. **批处理**：一次处理多个文档
2. **缓存**：缓存提取的实体和关系
3. **并行提取**：并发从多个文档中提取
4. **混合方法**：对已知实体使用手动定义，对新内容使用 LLM

## 展示的功能

### 1. 自动知识图谱构建

```go
// 每个文档都会被自动处理：
documents := []rag.Document{
    {
        Content: "Apple Inc. 是一家科技公司...",
        Metadata: map[string]any{
            "source": "apple_overview.txt",
            "topic":  "Apple Inc.",
        },
    },
}

// 添加到知识图谱
err := graphEngine.AddDocuments(ctx, documents)
```

### 2. 实体和关系管理

- **实体类型**：PERSON、ORGANIZATION、LOCATION、PRODUCT、TECHNOLOGY、CONCEPT
- **关系类型**：WORKS_AT、FOUNDED_BY、BASED_IN、PRODUCES、COMPETES_WITH
- **实体属性**：自定义属性以丰富实体描述

### 3. 基于图的检索

```go
// 使用图上下文查询
result, err := graphEngine.Query(ctx, query)
if err == nil {
    fmt.Printf("找到 %d 个实体和 %d 个关系\n",
        len(result.Entities), len(result.Relationships))
}
```

### 4. 实体探索

```go
// 查找相关实体
relatedEntities, err := kg.GetRelatedEntities(ctx, "apple_inc", 2)

// 查询特定实体类型
graphQuery := &rag.GraphQuery{
    EntityTypes: []string{"PERSON"},
    Limit:       10,
}
```

## 使用场景

本示例适用于：

1. **文档分析**：从大型文档集合中自动提取知识
2. **知识管理**：将非结构化文本构建为可搜索的知识图谱
3. **研究应用**：在研究论文中发现隐藏的关系
4. **企业知识库**：将内部文档转换为可查询的图谱
5. **问答系统**：基于图关系提供上下文感知的答案

## 高级功能

### 1. 自定义实体提取提示

```go
graphRAGConfig.ExtractionPrompt = `
从以下文本中提取实体。重点关注这些实体类型：%s。
返回具有此结构的 JSON 响应：
{
  "entities": [
    {
      "name": "实体名称",
      "type": "实体类型",
      "description": "简要描述",
      "properties": {}
    }
  ]
}

文本：%s`
```

### 2. 关系检测

系统自动检测各种关系类型：
- **雇佣关系**：WORKS_AT、CEO_OF
- **创建关系**：FOUNDED_BY、CREATED_BY
- **位置关系**：BASED_IN、LOCATED_IN
- **产品关系**：PRODUCES、MANUFACTURES
- **竞争关系**：COMPETES_WITH、RIVAL_OF

### 3. 图可视化

示例包含 Mermaid 图表可视化：

```go
exporter := graph.NewExporter(pipeline.GetGraph())
fmt.Println(exporter.DrawMermaid())
```

## 故障排除

### 常见问题

1. **连接失败**：确保 FalkorDB 正在运行且可访问
2. **实体提取失败**：检查 OpenAI API 密钥和网络连接
3. **性能慢**：考虑批处理或手动实体定义
4. **内存使用**：监控大型图的 Redis 内存使用

### 调试模式

启用调试输出以查看内部处理：

```go
// 检查提取了哪些实体
fmt.Printf("提取了 %d 个实体\n", len(entities))
for _, entity := range entities {
    fmt.Printf("- %s (%s)\n", entity.Name, entity.Type)
}
```

## 扩展

### 1. 混合方法

结合自动提取与手动定义：

```go
// 手动添加已知实体
knownEntities := preloadKnownEntities()

// 从文档中提取新实体
extractedEntities := extractFromDocuments(ctx, documents)

// 合并并添加到知识图谱
allEntities := append(knownEntities, extractedEntities...)
```

### 2. 自定义关系类型

为特定领域定义您自己的关系类型：

```go
relationshipTypes := []string{
    "PARTNERS_WITH",  // 合作关系
    "ACQUIRES",        // 收购关系
    "INVESTS_IN",      // 投资关系
    "COLLABORATES_ON", // 合作项目
}
```

### 3. 丰富管道

添加外部数据源以丰富实体：

```go
// 使用外部 API 丰富实体
for _, entity := range entities {
    if entity.Type == "ORGANIZATION" {
        externalInfo := fetchCompanyData(entity.Name)
        mergeProperties(entity.Properties, externalInfo)
    }
}
```

## 最佳实践

1. **从小开始**：从少数文档开始并逐步扩展
2. **验证提取**：检查提取的实体是否准确
3. **优化提示**：为您领域自定义提取提示
4. **监控性能**：跟踪处理时间并优化瓶颈
5. **定期更新**：定期用新文档刷新知识图谱

## 集成

### 与现有 RAG 系统集成

此 FalkorDB 集成可以与传统的基于向量的 RAG 结合：

```go
// 结合向量和图搜索的混合检索器
vectorRetriever := retriever.NewVectorStoreRetriever(vectorStore, embedder, k)
graphRetriever := retriever.NewKnowledgeGraphRetriever(kg, k)
hybridRetriever := retriever.NewHybridRetriever(vectorRetriever, graphRetriever, 0.5)
```

### 与 LangChain 集成

```go
// 使用 LangChain 组件
embedder := rag.NewLangChainEmbedder(openaiEmbedder)
vectorStore := rag.NewLangChainVectorStore(chromaStore)
```

## 下一步

1. **尝试不同数据**：替换为您自己领域的特定文档
2. **自定义实体类型**：添加与您领域相关的实体类型
3. **批处理**：实现自动化文档处理流水线
4. **集成**：连接到您现有的 RAG 应用
5. **监控**：为生产使用添加指标和监控

## 贡献

欢迎贡献以改进 FalkorDB 集成！请：

1. Fork 仓库
2. 创建功能分支
3. 添加您的改进
4. 提交 Pull Request

## 许可证

此示例是 LangGraphGo 项目的一部分。有关许可证信息，请参阅主仓库。