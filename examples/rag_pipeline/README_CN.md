# RAG 管道示例

本示例演示一个完整的、模块化的 RAG 管道。

## 概述

本示例演示如何使用 `RAGPipeline` 组件实现：
- **构建知识库**：一键式将本地文档目录导入向量库。
- **智能问答**：构建一个能够基于特定私有文档回答问题的 Agent。

## 核心架构

1. **加载 (Load)**：使用 `TextLoader` 加载本地文件。
2. **拆分 (Split)**：使用 `RecursiveCharacterTextSplitter` 进行文档切分。
3. **嵌入 (Embed)**：集成 LangChain OpenAI embeddings 接口。
4. **存储 (Store)**：内存向量数据库。
5. **流水线 (Pipeline)**：使用 `RAGPipeline` 编排检索和生成逻辑。

## 用法

确保已设置 `OPENAI_API_KEY` 环境变量。

```bash
cd examples/rag_pipeline
go run main.go
```
