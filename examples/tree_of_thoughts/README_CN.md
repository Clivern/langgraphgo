# Tree of Thoughts 示例

## 概述

本示例演示了 **Tree of Thoughts (ToT, 思维树)** 模式，这是一种基于搜索的推理框架，其中问题求解被建模为通过树结构的搜索过程。ToT 将线性的思维链扩展为多路径探索搜索，在每个步骤生成多个候选"思维"，评估其可行性，并修剪无望的分支，同时扩展最有希望的分支。

## 什么是 Tree of Thoughts？

Tree of Thoughts 是一种代理架构，通过搜索树实现对问题空间的系统探索。与线性推理方法不同，ToT 同时探索多个解决方案路径，并使用评估来引导搜索朝向目标前进。

## 五个关键阶段

1. **分解 (Decomposition)**：将问题分解为一系列步骤或思维
2. **思维生成 (Thought Generation)**：生成多个潜在的下一步，在搜索树中创建分支
3. **状态评估 (State Evaluation)**：评估每个思维的：
   - 有效性：此移动是否遵循规则？
   - 进度：我们是否更接近解决方案？
   - 启发式：此路径有多大希望？
4. **剪枝与扩展 (Pruning & Expansion)**：修剪无效或无望的分支，从最有希望的活跃分支继续
5. **求解 (Solution)**：继续直到达到目标状态；解决方案是从根到目标的路径

## 架构

```
                问题初始状态
                        │
        ┌───────────────┼───────────────┐
        │               │               │
    思维 1           思维 2           思维 3
        │               │               │
    ┌───┼───┐       ┌───┼───┐       [已剪枝]
    │       │       │       │
 状态   状态   状态   状态
  1.1     1.2     2.1     2.2
    │       X       │       X
    │   [无效]      │   [低分]
 目标               │
[找到!]         继续...
```

## 核心组件

### 1. ThoughtState 接口

表示搜索树中的状态：

```go
type ThoughtState interface {
    IsValid() bool            // 检查状态是否遵循规则
    IsGoal() bool            // 检查这是否是解决方案
    GetDescription() string  // 人类可读的描述
    Hash() string           // 用于循环检测的唯一标识符
}
```

### 2. ThoughtGenerator 接口

生成可能的下一个状态：

```go
type ThoughtGenerator interface {
    Generate(ctx context.Context, current ThoughtState) ([]ThoughtState, error)
}
```

### 3. ThoughtEvaluator 接口

评估状态质量：

```go
type ThoughtEvaluator interface {
    Evaluate(ctx context.Context, state ThoughtState, pathLength int) (float64, error)
}
```

## 配置

```go
type TreeOfThoughtsConfig struct {
    Generator    ThoughtGenerator  // 创建新状态
    Evaluator    ThoughtEvaluator // 对状态评分
    MaxDepth     int              // 最大搜索深度（默认：10）
    MaxPaths     int              // 维护的最大活跃路径数（默认：5）
    Verbose      bool             // 启用详细日志
    InitialState ThoughtState     // 起始状态
}
```

## 示例：过河问题

包含的示例解决了经典的狼羊卷菜过河问题：

**问题**：农夫需要将狼、羊和卷菜运过河。

**约束条件**：
1. 船一次只能载农夫和最多一个其他物品
2. 狼不能与羊单独在一起（狼吃羊）
3. 羊不能与卷菜单独在一起（羊吃卷菜）

### 解决方案

Tree of Thoughts 代理通过系统搜索找到解决方案：

```
步骤 1：农夫带羊过河
步骤 2：农夫独自返回
步骤 3：农夫带狼过河
步骤 4：农夫带羊返回
步骤 5：农夫带卷菜过河
步骤 6：农夫独自返回
步骤 7：农夫带羊过河
```

## 运行示例

```bash
cd examples/tree_of_thoughts
go run main.go
```

## 工作原理

1. **初始化**：所有物品从左岸开始
2. **扩展**：生成所有可能的合法移动（农夫独自、农夫+狼、农夫+羊、农夫+卷菜）
3. **验证**：检查每个状态是否违反规则（例如，狼和羊单独在一起）
4. **循环检测**：使用状态哈希避免重访相同状态
5. **评估**：基于进度对状态评分（右岸有多少物品）
6. **剪枝**：移除无效状态并只保留前 N 个最有希望的路径
7. **目标检查**：继续直到找到所有物品都在右岸的状态

## 何时使用 Tree of Thoughts

**最适合**：
- 具有明确规则和目标状态的逻辑谜题
- 动作顺序和约束至关重要的复杂规划问题
- 需要探索多种策略的问题
- 约束满足问题

**不理想的场景**：
- 简单、直接的任务
- 没有明确验证规则的问题
- 评估主观或模糊的任务

## 优势

- **健壮性**：与单次通过方法相比，系统探索大大减少了错误
- **正确性**：可靠性来自算法的合理性，而非模型记忆
- **组合复杂性**：擅长处理具有大可能性空间的问题
- **可验证**：每个步骤都可以根据规则进行验证

## 劣势

- **计算成本**：需要更多的 LLM 调用和状态管理操作
- **速度**：由于探索，比线性方法慢
- **评估器依赖**：搜索有效性严重依赖于评估函数的质量

## 核心洞察

与依赖 LLM 记忆知识的思维链提示不同，Tree of Thoughts 通过可验证的搜索**发现**解决方案。即使 LLM 在某个分支做出次优选择，代理也会继续探索其他分支，为具有明确定义规则的问题保证最终的正确性。

## 实现要点

示例演示了：
- 自定义状态表示 (`RiverState`)
- 基于规则的验证（检查禁止的组合）
- 状态生成（所有可能的船只横渡）
- 启发式评估（朝向目标的进度）
- 循环检测（避免重访状态）

## 参考

- [All Agentic Architectures](https://github.com/FareedKhan-dev/all-agentic-architectures) - 原始 Tree of Thoughts 实现
- 模式 #9：Tree of Thoughts (ToT)
- [Tree of Thoughts 论文](https://arxiv.org/abs/2305.10601) - "Tree of Thoughts: Deliberate Problem Solving with Large Language Models"

## 许可证

此实现是 langgraphgo 项目的一部分。
