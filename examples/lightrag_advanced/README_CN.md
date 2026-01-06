# LightRAG 高级示例

一个展示 LightRAG 高级功能的示例，包含自定义配置和性能对比。

## 概述

本综合示例展示：
- **自定义提示模板**用于实体和关系提取
- **社区检测**用于全局检索
- **融合方法对比**（RRF vs 加权）
- **跨检索模式的性能基准测试**
- **知识图谱遍历**
- **文档操作**（添加、更新、删除）

## 前置要求

```bash
go mod tidy
```

## 运行示例

### 不使用 OpenAI API Key（使用 Mock LLM）

```bash
go run main.go
```

### 使用 OpenAI API Key

获得更好的实体提取效果：

```bash
export OPENAI_API_KEY="your-api-key-here"
go run main.go
```

## 展示的功能

### 1. 自定义配置

```go
config := rag.LightRAGConfig{
    Mode:                 "hybrid",
    Temperature:          0.7,
    ChunkSize:            512,
    MaxEntitiesPerChunk:  20,
    EnableCommunityDetection: true,
    PromptTemplates: map[string]string{
        "entity_extraction": "...",
        "relationship_extraction": "...",
    },
}
```

### 2. 检索模式对比

示例在所有四种模式下运行相同查询：
- **Naive**：快速，基础检索
- **Local**：以实体为中心，支持多跳推理
- **Global**：社区级别摘要
- **Hybrid**：两全其美

### 3. 融合方法对比

比较混合模式的两种融合策略：
- **RRF（倒数排名融合）**：基于排名的融合
- **Weighted**：基于分数的加权融合

### 4. 知识图谱操作

```go
// 查询知识图谱
result, err := graphKg.Query(ctx, &rag.GraphQuery{
    EntityTypes: []string{"CONCEPT", "TECHNOLOGY"},
    Limit:       5,
})
```

### 5. 文档操作

```go
// 添加新文档
lightrag.AddDocuments(ctx, []rag.Document{newDoc})

// 更新文档
lightrag.UpdateDocument(ctx, updatedDoc)
```

### 6. 性能基准测试

运行多个查询以测量：
- 每种模式的平均响应时间
- 成功率
- 延迟统计

## 示例文档

示例使用关于 AI 和机器学习的文档：
- Transformer 架构
- 神经网络
- 大语言模型（LLMs）
- 机器学习基础
- 注意力机制
- RAG（检索增强生成）
- 微调
- 嵌入

## 预期输出

```
=== LightRAG Advanced Example ===
This example demonstrates advanced features of LightRAG including:
- Custom prompt templates
- Community detection
- Different fusion methods
- Performance comparison between modes

Indexing documents...
Indexed 8 documents in 15ms

=== Retrieval Mode Comparison ===

--- Naive Mode ---
Response Time: 250µs
Sources Retrieved: 3
Confidence: 0.15

--- Local Mode ---
Response Time: 450µs
Sources Retrieved: 5
Query Entities: 2

--- Global Mode ---
Response Time: 380µs
Sources Retrieved: 4
Communities: 2

--- Hybrid Mode ---
Response Time: 520µs
Sources: 5
Local Confidence: 0.25
Global Confidence: 0.18
...
```

## 高级配置选项

### 本地检索

```go
LocalConfig: rag.LocalRetrievalConfig{
    TopK:               15,        // 实体数量
    MaxHops:            3,         // 图谱遍历深度
    IncludeDescriptions: true,
    EntityWeight:       0.8,       // 实体相关性权重
}
```

### 全局检索

```go
GlobalConfig: rag.GlobalRetrievalConfig{
    MaxCommunities:     10,        // 要检索的社区数
    IncludeHierarchy:   true,      // 包含层次结构
    MaxHierarchyDepth:  5,         // 层次深度
    CommunityWeight:    0.7,       // 社区相关性
}
```

### 混合融合

```go
HybridConfig: rag.HybridRetrievalConfig{
    LocalWeight:  0.6,             // 60% 本地
    GlobalWeight: 0.4,             // 40% 全局
    FusionMethod: "rrf",           // 或 "weighted"
    RFFK:         60,              // RRF 参数
}
```

## 性能建议

1. **分块大小**：较大的分块 = 更少的 API 调用，检索精度较低
2. **MaxHops**：限制在 2-3 以获得更好的性能
3. **TopK**：从 10-20 开始，根据结果调整
4. **社区检测**：对于小数据集（< 100 个文档）禁用

## 代码结构

- `main()`: 协调所有演示
- `OpenAILLMAdapter`: 包装 OpenAI LLM
- `MockLLM`: 演示用的 Mock 实现
- `createSampleDocuments()`: 创建测试文档

## 相关文档

- [LightRAG 简单示例](../lightrag_simple/) - 基本用法
- [LightRAG 文档](../../docs/lightrag.md) - 完整文档
