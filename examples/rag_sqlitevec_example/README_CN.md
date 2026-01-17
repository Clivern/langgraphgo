# sqlite-vec 向量存储示例

本示例演示如何使用 [sqlite-vec](https://github.com/asg017/sqlite-vec) 作为 LangGraphGo RAG 管道的向量存储。

## 什么是 sqlite-vec？

sqlite-vec 是一个用纯 C 编写的超轻量级向量搜索 SQLite 扩展，具有以下特点：

- **零外部依赖** - 纯 C 实现
- **嵌入式存储** - 使用标准 SQLite 文件
- **跨平台** - 可在任何运行 SQLite 的地方运行（Linux、macOS、Windows、通过 WASM 在浏览器中运行）
- **多种向量类型** - 支持 float32、int8 和二进制向量
- **KNN 搜索** - 高效的 K 近邻向量搜索

## 功能演示

本示例展示了：

1. **创建 sqlite-vec 向量存储** 并使用持久化存储
2. **添加文档** 包含嵌入向量和元数据
3. **构建 RAG 管道** 使用向量存储
4. **相似度搜索** 进行文档检索
5. **元数据过滤** 缩小搜索结果范围
6. **持久化存储验证**（数据在存储重启后仍然存在）
7. **更新操作**（删除并重新插入模式）

## 运行示例

```bash
cd examples/rag_sqlitevec_example
go run main.go
```

## 代码概览

### 初始化向量存储

```go
store, err := store.NewSQLiteVecVectorStore(store.SQLiteVecConfig{
    DBPath:         "./vectors.db",  // SQLite 数据库文件路径
    CollectionName: "my_collection", // 集合/表名
    Embedder:       embedder,        // 嵌入函数
})
```

### 添加文档

```go
documents := []rag.Document{
    {
        ID:      "doc1",
        Content: "你的文档内容",
        Metadata: map[string]any{"category": "tech"},
    },
}

err := store.Add(ctx, documents)
```

### 创建 RAG 管道

```go
vectorRetriever := retriever.NewVectorStoreRetriever(store, embedder, 2)

config := rag.DefaultPipelineConfig()
config.Retriever = vectorRetriever
config.LLM = llm

pipeline := rag.NewRAGPipeline(config)
pipeline.BuildBasicRAG()

runnable, _ := pipeline.Compile()
```

### 查询管道

```go
result, err := runnable.Invoke(ctx, map[string]any{
    "query": "你的问题",
})

answer := result["answer"].(string)
documents := result["documents"].([]rag.RAGDocument)
```

## 存储选项

### 内存存储

用于临时数据或测试：

```go
store, err := store.NewSQLiteVecVectorStoreSimple("", embedder)
```

### 持久化存储

用于长期存储：

```go
store, err := store.NewSQLiteVecVectorStoreSimple("./vectors.db", embedder)
```

## 元数据过滤

搜索具有特定元数据的文档：

```go
queryEmbedding, _ := embedder.EmbedDocument(ctx, "搜索查询")

results, err := store.SearchWithFilter(ctx, queryEmbedding, 10, map[string]any{
    "category": "tech",
})
```

## sqlite-vec 的优势

1. **无外部服务** - 所有操作都在你的进程中运行
2. **部署简单** - 只需复制 SQLite 文件
3. **ACID 事务** - 完整的 SQLite 事务支持
4. **SQL 集成** - 可将向量搜索与关系查询结合
5. **占用空间小** - 扩展只有几百 KB

## 使用场景

- **边缘应用** - 在无互联网的设备上运行
- **桌面应用** - GUI 应用程序中的本地向量搜索
- **无服务器** - 无需单独的向量数据库服务
- **开发环境** - 简化本地开发和测试
- **多租户** - 每个租户使用单独的 SQLite 文件

## 与其他向量存储的对比

| 特性 | sqlite-vec | Chroma | Pinecone |
|---------|-----------|--------|----------|
| 嵌入式 | ✅ 是 | ❌ 否 | ❌ 否 |
| 外部服务 | ❌ 否 | ✅ 是 | ✅ 是 |
| SQL 查询 | ✅ 是 | ❌ 否 | ❌ 否 |
| ACID 事务 | ✅ 是 | ✅ 是 | ✅ 是 |
| 部署复杂度 | ⭐ 低 | ⭐⭐ 中 | ⭐⭐⭐ 高 |

## 生产环境建议

1. **使用真实嵌入** - 用 OpenAI 或类似服务替换 mock embedder
2. **调整维度** - 匹配嵌入模型的输出维度
3. **批量操作** - 批量添加文档以获得更好性能
4. **定期备份** - 复制 SQLite 文件进行备份
5. **索引优化** - 对于大数据集，考虑分区

## 技术细节

### 向量存储结构

sqlite-vec 使用 `vec0` 虚拟表存储向量，支持辅助列（auxiliary columns）存储元数据：

```sql
CREATE VIRTUAL TABLE table_name USING vec0(
    embedding float[128],  -- 向量列
    id TEXT PRIMARY KEY,    -- 文档 ID
    content TEXT,           -- 文档内容
    metadata TEXT,          -- 元数据（JSON）
    created_at INTEGER,     -- 创建时间
    updated_at INTEGER      -- 更新时间
)
```

### 向量搜索

使用 `MATCH` 操作符进行 KNN 搜索：

```sql
SELECT id, content, metadata, distance
FROM table_name
WHERE embedding MATCH ?
ORDER BY distance
LIMIT ?
```

### 元数据处理

元数据以 JSON 格式存储在 TEXT 列中，支持：
- 结构化元数据
- 灵活的查询和过滤
- 与 SQL JSON 函数的互操作性

## 限制与注意事项

### 更新操作

由于 `vec0` 虚拟表的限制，更新操作需要先删除再插入：

```go
// 先删除现有文档
store.Delete(ctx, []string{"doc_id"})

// 再添加更新后的文档
store.Add(ctx, []rag.Document{updatedDoc})
```

### 表名限制

- 表名会被清理（非字母数字字符替换为下划线）
- SQL 关键字会被自动用双引号括起来
- 建议使用简单的字母数字表名

### 元数据过滤

当前实现对辅助列的 SQL 过滤有限制，采用内存中过滤的方式：
- 获取更大的结果集（k * 10）
- 在应用层应用元数据过滤
- 返回过滤后的前 k 个结果

## 参考资源

- [sqlite-vec GitHub](https://github.com/asg017/sqlite-vec)
- [sqlite-vec 文档](https://alexgarcia.xyz/sqlite-vec/)
- [RAG 管道文档](../../rag/README.md)
- [其他向量存储示例](../rag_chromem_example/)

## 许可证

本示例遵循 LangGraphGo 项目的许可证。
