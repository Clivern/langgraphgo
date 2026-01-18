# BM25 检索器示例

本示例演示如何在 LangGraphGo 中使用 BM25 (Best Matching 25) 稀疏检索系统。

## 概述

BM25 是一种基于查询词频的概率信息检索函数，用于对文档进行排序。它在基于关键字的搜索中特别有效，并可结合向量检索实现混合搜索系统。

## 演示功能

1. **基础 BM25 检索** - 简单的关键字文档搜索
2. **分数阈值过滤** - 根据最小相关性分数过滤结果
3. **自定义分词器** - 支持英文、中文和自定义正则分词
4. **动态文档管理** - 运行时添加、更新和删除文档
5. **参数调优** - 调整 k1 和 b 参数以获得最佳性能
6. **混合检索设置** - 结合 BM25 和向量检索

## 运行示例

### 构建并运行

```bash
go build -o bm25_demo main.go
./bm25_demo
```

或直接运行：

```bash
go run main.go
```

## 代码示例

### 基础用法

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/smallnest/langgraphgo/rag"
    "github.com/smallnest/langgraphgo/rag/retriever"
)

func main() {
    // 准备文档
    docs := []rag.Document{
        {
            ID:      "doc1",
            Content: "LangGraph 是一个用于构建有状态 LLM 应用的框架",
            Metadata: map[string]any{
                "title": "LangGraph 概述",
            },
        },
        {
            ID:      "doc2",
            Content: "BM25 是信息检索的排序函数",
            Metadata: map[string]any{
                "title": "BM25 概述",
            },
        },
    }

    // 创建 BM25 检索器
    config := retriever.DefaultBM25Config()
    config.K = 2 // 检索前 2 个文档

    bm25Retriever, err := retriever.NewBM25Retriever(docs, config)
    if err != nil {
        log.Fatal(err)
    }

    // 查询文档
    ctx := context.Background()
    results, err := bm25Retriever.Retrieve(ctx, "框架 LLM")
    if err != nil {
        log.Fatal(err)
    }

    // 显示结果
    fmt.Printf("找到 %d 个结果:\n", len(results))
    for i, result := range results {
        fmt.Printf("%d. [%s] %s\n", i+1, result.Metadata["title"], result.Content)
    }
}
```

### 使用分数阈值

```go
// 只返回分数 >= 0.5 的文档
config := retriever.DefaultBM25Config()
config.K = 10
config.ScoreThreshold = 0.5

bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)

ctx := context.Background()
retrievalConfig := &rag.RetrievalConfig{
    K:              10,
    ScoreThreshold: 0.5,
}

results, _ := bm25Retriever.RetrieveWithConfig(ctx, query, retrievalConfig)

for _, result := range results {
    fmt.Printf("分数: %.4f | %s\n", result.Score, result.Document.Content)
}
```

### 中文文本支持

```go
import "github.com/smallnest/langgraphgo/rag/tokenizer"

// 创建中文文档
docs := []rag.Document{
    {ID: "doc1", Content: "Go语言支持并发编程"},
    {ID: "doc2", Content: "Python是一种易学的编程语言"},
}

// 使用中文分词器
chineseTokenizer := tokenizer.NewChineseTokenizer()

config := retriever.DefaultBM25Config()
bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    chineseTokenizer,
)

ctx := context.Background()
results, _ := bm25Retriever.Retrieve(ctx, "编程语言")
```

### 动态文档管理

```go
// 添加新文档
newDocs := []rag.Document{
    {ID: "doc3", Content: "新文档内容"},
}
bm25Retriever.AddDocuments(newDocs)

// 更新现有文档
bm25Retriever.UpdateDocument(rag.Document{
    ID:      "doc1",
    Content: "更新后的内容",
})

// 删除文档
bm25Retriever.DeleteDocument("doc2")

// 获取统计信息
stats := bm25Retriever.GetStats()
fmt.Printf("文档总数: %v\n", stats["num_documents"])
```

### 混合检索 (BM25 + 向量)

```go
// 创建 BM25 检索器（稀疏检索）
bm25Config := retriever.DefaultBM25Config()
bm25Retriever, _ := retriever.NewBM25Retriever(docs, bm25Config)

// 创建向量检索器（密集检索）
vectorConfig := rag.RetrievalConfig{K: 5}
vectorRetriever := retriever.NewVectorRetriever(
    vectorStore,
    embedder,
    vectorConfig,
)

// 使用等权重组合两者
hybridRetriever := retriever.NewHybridRetriever(
    []rag.Retriever{bm25Retriever, vectorRetriever},
    []float64{0.5, 0.5},
    rag.RetrievalConfig{K: 5},
)

// 使用混合检索查询
ctx := context.Background()
results, _ := hybridRetriever.Retrieve(ctx, "查询内容")
```

## BM25 参数

### k1 参数 (词频饱和度)

控制词频对分数的影响程度：

- **范围**: 1.2 - 2.0
- **较低值** (如 1.2): 减少词频影响
- **较高值** (如 2.0): 增加词频影响
- **使用场景**:
  - 短查询: 使用较高的 k1
  - 长查询: 使用较低的 k1

```go
config := retriever.DefaultBM25Config()
config.K1 = 1.5 // 默认值
```

### b 参数 (文档长度归一化)

控制文档长度对分数的影响：

- **范围**: 0.0 - 1.0
- **0.0**: 不进行长度归一化
- **1.0**: 完全长度归一化
- **推荐值**: 0.75 (默认)

```go
config := retriever.DefaultBM25Config()
config.B = 0.75 // 默认值，适用于大多数场景
```

## 分词器

### 默认单词分词器

```go
// 自动使用基于正则的单词分词
bm25Retriever, _ := retriever.NewBM25Retriever(docs, config)
```

### 中文分词器

```go
import "github.com/smallnest/langgraphgo/rag/tokenizer"

chineseTokenizer := tokenizer.NewChineseTokenizer()
bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    chineseTokenizer,
)
```

### 自定义正则分词器

```go
// 使用自定义模式创建分词器
regexTokenizer, _ := tokenizer.NewRegexTokenizer(`\b[a-zA-Z]+\b`)
bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    regexTokenizer,
)
```

### N-gram 分词器

```go
// 创建二元分词器
baseTokenizer := tokenizer.DefaultRegexTokenizer()
bigramTokenizer := tokenizer.NewNgramTokenizer(2, baseTokenizer)

bm25Retriever, _ := retriever.NewBM25RetrieverWithTokenizer(
    docs,
    config,
    bigramTokenizer,
)
```

## 预期输出

运行示例时，您应该看到类似以下的输出：

```
=== BM25 检索器演示 ===

1. 基础 BM25 检索
-----------------------
查询: framework for building LLM applications
找到 2 个结果:
  1. [LangGraph Overview] LangGraph is a framework for building stateful...
  2. [RAG Overview] RAG combines retrieval systems with generation...

2. BM25 分数阈值
----------------------------
查询: neural networks learning (阈值: 0.5)
找到 3 个高于阈值的结果:
  1. [分数: 3.1372] Deep learning uses neural networks...
  2. [分数: 0.9276] Machine learning algorithms learn...
  3. [分数: 0.9276] Neural networks are inspired by...

3. BM25 自定义分词器
------------------------------
查询: 编程语言
找到 2 个结果:
  1. Go语言支持并发编程
  2. Python是一种易学的编程语言

4. 混合检索设置
-------------------------
BM25 结果 'search similarity matching':
  1. [分数: 1.3852] Vector search uses embeddings...
  2. [分数: 0.9365] BM25 uses term frequency...

5. 动态文档管理
-------------------------------
初始文档数: 1
添加后: 3
已更新 doc1
删除 doc2 后: 2

6. 参数调优
-------------------
查询: quick fast
测试 k1 参数 (词频饱和度):
  k1=0.5: [doc1:0.98] [doc2:0.98]
  k1=1.5: [doc1:0.98] [doc2:0.98]
```

## 使用场景

1. **关键字搜索**: 当用户搜索特定术语时
2. **混合搜索**: 结合向量检索进行语义+关键字匹配
3. **多语言**: 支持英文、中文等多种语言
4. **动态索引**: 文档频繁变化的场景
5. **快速原型**: 无需向量数据库的快速检索

## 性能建议

- **精确关键字匹配**: 单独使用 BM25
- **语义理解**: 单独使用向量检索
- **最佳效果**: 结合 BM25 和向量进行混合检索
- **索引大小**: BM25 索引通常比向量索引小
- **查询速度**: 对于大型数据集，BM25 比向量检索快

## 对比: BM25 vs 向量检索

| 特性 | BM25 | 向量 |
|---------|------|--------|
| 类型 | 稀疏 (关键字) | 密集 (语义) |
| 索引大小 | 小 | 大 |
| 查询速度 | 快 | 较慢 |
| 精确匹配 | 优秀 | 较弱 |
| 语义理解 | 较弱 | 优秀 |
| 最适用于 | 关键字、技术术语 | 概念、含义 |

## 另请参阅

- [BM25 集成文档](../../docs/bm25_integration.md)
- [BM25 方案总结](../../docs/bm25_summary.md)
- [LangGraphGo 文档](../../README.md)

## 许可证

LangGraphGo 项目的一部分。详见主 LICENSE 文件。
