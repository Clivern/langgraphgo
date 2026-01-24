# 从零构建 AI 漫画生成智能体：LangGraphGo + Skills 实战指南

![封面图](cover-image/cover.png)

> 本文将带你深入了解如何使用 LangGraphGo 框架结合 Skills 插件系统，从零开始构建一个能够自动生成漫画的 AI 智能体。我们会深入剖析技术架构，分享踩坑经验，并提供完整的代码实现。

## 前言

![AI智能体架构概念图](cover-image/01-architecture.png)

在 AI 应用开发领域，智能体（Agent）架构正变得越来越重要。与传统的单一 LLM 调用不同，智能体能够：

- 自主规划任务执行步骤
- 调用外部工具完成任务
- 根据执行结果动态调整策略

本文将以一个**漫画生成智能体**为例，展示如何使用 LangGraphGo 框架，使用当前炙手可热的Skill技术，构建复杂的多步骤 AI 应用。这个智能体能够：

1. 根据用户输入生成漫画分镜脚本
2. 自动调用图像生成模型生成每一页画面
3. 将所有页面合并成完整的 PDF 漫画

> 本示例使用宝玉的漫画 Skill作为漫画生成的核心工具，演示了langGraphGo与Skills插件系统的无缝集成。

## 技术栈概览

![技术栈展示图](cover-image/02-techstack.png)

```
┌─────────────────────────────────────────────────────────────┐
│                         用户输入                              │
│                  "创作一个采蘑菇的小姑娘的漫画"                  │
└─────────────────────────┬───────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────────────┐
│                    LangGraphGo Agent                         │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐  │
│  │   LLM 核心    │───▶│  工具调度器   │───▶│  状态管理器   │  │
│  │   (ERNIE)    │    │ (Tool Router) │    │   (State)    │  │
│  └──────────────┘    └──────────────┘    └──────────────┘  │
└─────────────────────────┬───────────────────────────────────┘
                          │
          ┌───────────────┼───────────────┐
          ▼               ▼               ▼
   ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
   │ 分镜生成工具 │ │ 图像生成工具 │ │ PDF合并工具 │
   │(baoyu-comic)│ │(image-gen)  │ │(baoyu-comic)│
   │  (.ts脚本)  │ │  (.ts脚本)  │ │  (.ts脚本)  │
   └─────────────┘ └─────────────┘ └─────────────┘
                          │
            ┌─────────────┴─────────────┐
            ▼                           ▼
     ┌─────────────┐          ┌─────────────┐
     │ pdf skill   │          │  其他技能   │
     │ (.py脚本)   │          │  (未使用)    │
     └─────────────┘          └─────────────┘
```

**核心技术组件：**

- **LangGraphGo**: Go 语言实现的 LangGraph 框架，提供状态图（StateGraph）能力
- **GoSkills v0.6.1+**: 技能插件系统，将脚本封装为 LLM 可调用的工具
- **TypeScript 脚本**: 实际执行业务逻辑的脚本层，使用 npx tsx 执行
- **ERNIE 5.0 Thinking Preview**: 百度文心一言大模型，工具调用稳定，负责理解和规划

## 一、项目架构设计

![项目目录结构图](cover-image/03-project-structure.png)

### 1.1 目录结构

```
comic_skill_example/
├── main.go                 # 入口文件，Agent 创建和执行
├── go.mod                  # Go 模块依赖
└── skills/                 # 技能插件目录
    ├── baoyu-comic/        # 漫画分镜生成技能
    │   ├── SKILL.md        # 技能定义（含工具元数据）
    │   └── scripts/
    │       ├── generate-comic.ts    # 分镜生成脚本
    │       └── merge-to-pdf.ts      # PDF 合并脚本
    ├── baoyu-image-gen/     # 图像生成技能
    │   ├── SKILL.md         # 技能定义
    │   └── scripts/
    │       └── main.ts      # 图像生成脚本
    └── pdf/                 # PDF 处理技能（Python）
        ├── SKILL.md         # 技能定义
        └── scripts/
            ├── check_bounding_boxes.py
            ├── convert_pdf_to_images.py
            ├── extract_form_field_info.py
            └── ...
```

**说明：** 系统会自动加载 `skills/` 目录下的所有技能包，但漫画生成 Agent 只使用其中 3 个核心工具（`generate_comic_storyboard`、`generate_comic_image`、`merge_comic_to_pdf`），这些工具来自 `baoyu-comic` 和 `baoyu-image-gen` 两个技能包。

### 1.2 技能定义系统（SKILL.md）

每个技能通过 `SKILL.md` 文件定义，使用 YAML frontmatter 声明工具：

```yaml
---
name: baoyu-comic
description: Knowledge comic creator supporting multiple styles...
tools:
  - name: generate_comic_storyboard
    script: scripts/generate-comic.ts
    description: 创建完整的漫画分镜脚本和提示词
    parameters:
      topic:
        type: string
        description: 要创作的漫画主题
        required: true
      style:
        type: string
        description: 视觉风格（如：warm 温暖、classic 经典）
        required: false
      pages:
        type: integer
        description: 要生成的页数
        required: false
      aspect:
        type: string
        description: 宽高比（如：3:4、4:3、16:9）
        required: false
---
```

**设计亮点：**

1. **声明式工具定义** - 工具名称、参数、描述全部在 SKILL.md 中声明
2. **自动 Schema 生成** - 系统自动根据参数定义生成 OpenAPI Schema
3. **零 Go 代码修改** - 添加新工具无需修改 Go 代码，只需编辑 SKILL.md

## 二、核心实现解析

![工具配置自动发现流程图](cover-image/04-tool-discovery.png)

### 2.1 工具配置自动发现机制

传统方式需要在 Go 代码中硬编码工具配置（60+ 行），我们实现了从 SKILL.md 自动读取。

**关键实现：adapter/goskills/goskills.go**

```go
// buildToolConfigFromSkill 从 SKILL.md 中的工具定义自动构建 ToolConfig
func buildToolConfigFromSkill(skill *goskills.SkillPackage) *ToolConfig {
    if len(skill.Meta.Tools) == 0 {
        return nil
    }

    config := &ToolConfig{
        NameMapping:         make(map[string]string),
        DescriptionOverrides: make(map[string]string),
        SchemaOverrides:     make(map[string]map[string]any),
    }

    for _, toolDef := range skill.Meta.Tools {
        // 构建名称映射：从工具名到工具名（保持一致）
        config.NameMapping[toolDef.Name] = toolDef.Name

        // 设置描述
        if toolDef.Description != "" {
            config.DescriptionOverrides[toolDef.Name] = toolDef.Description
        }

        // 构建 Schema
        schema := map[string]any{
            "type":       "object",
            "properties": make(map[string]any),
        }

        if len(toolDef.Parameters) > 0 {
            var required []string
            for paramName, param := range toolDef.Parameters {
                prop := map[string]any{
                    "type": param.Type,
                }
                if param.Description != "" {
                    prop["description"] = param.Description
                }
                schema["properties"].(map[string]any)[paramName] = prop
                if param.Required {
                    required = append(required, paramName)
                }
            }
            if len(required) > 0 {
                schema["required"] = required
            }
        }

        schema["additionalProperties"] = false
        config.SchemaOverrides[toolDef.Name] = schema
    }

    return config
}

// SkillsToTools 自动从 SKILL.md 读取工具定义
func SkillsToTools(skill *goskills.SkillPackage, opts ...SkillsToToolsOptions) ([]tools.Tool, error) {
    var config *ToolConfig

    // 1. 首先尝试从 SKILL.md 自动构建配置
    skillConfig := buildToolConfigFromSkill(skill)
    if skillConfig != nil {
        config = skillConfig
    }

    // 2. 如果用户提供了配置，合并覆盖
    if len(opts) > 0 && opts[0].ToolConfig != nil {
        // 合并逻辑...
    }

    // 3. 生成工具...
}
```

**优势对比：**

| 特性 | 硬编码方式 | 自动发现方式 |
|------|------------|-------------|
| 代码量 | 60+ 行 | 0 行 |
| 维护成本 | 高（双份修改） | 低（单一数据源） |
| 扩展性 | 需要重新编译 | 无需改 Go 代码 |
| 类型安全 | 编译时检查 | 运行时检查 |

### 2.2 命名参数到命令行参数的转换

![参数格式转换流程图](cover-image/05-param-conversion.png)

LLM 返回的是 JSON 格式的命名参数：
```json
{
  "topic": "采蘑菇的小姑娘",
  "style": "warm",
  "pages": 1
}
```

但 TypeScript 脚本期望的是命令行参数格式：
```bash
npx tsx generate-comic.ts --topic "采蘑菇的小姑娘" --style "warm" --pages 1
```

**关键实现：adapter/goskills/goskills.go**

```go
func (t *SkillTool) Call(ctx context.Context, input string) (string, error) {
    originalName := t.name

    switch originalName {
    // ... 预定义工具的处理
    default:
        if scriptPath, ok := t.scriptMap[originalName]; ok {
            // 1. 尝试解析为命名参数格式
            var namedParams map[string]any
            err := json.Unmarshal([]byte(input), &namedParams)

            if err == nil && len(namedParams) > 0 {
                // 2. 转换为命令行参数
                var args []string

                paramMapping := map[string]string{
                    "topic":     "--topic",
                    "style":     "--style",
                    "pages":     "--pages",
                    "aspect":    "--aspect",
                    "path":      "--image",
                    "prompt":    "--prompt",
                    "ar":        "--ar",
                    "quality":   "--quality",
                    "directory": "--directory",
                }

                paramOrder := []string{"topic", "style", "pages", "aspect",
                                       "path", "prompt", "ar", "quality", "directory"}

                for _, key := range paramOrder {
                    if value, ok := namedParams[key]; ok && value != nil {
                        if flag, ok := paramMapping[key]; ok {
                            args = append(args, flag)
                            args = append(args, fmt.Sprintf("%v", value))
                        }
                    }
                }

                // 3. 根据脚本类型执行
                if strings.HasSuffix(scriptPath, ".py") {
                    return goskillstool.RunPythonScript(scriptPath, args)
                } else if strings.HasSuffix(scriptPath, ".ts") || strings.HasSuffix(scriptPath, ".js") {
                    return langgraphtool.RunTypeScriptScript(scriptPath, args)
                } else {
                    return langgraphtool.RunShellScript(scriptPath, args)
                }
            }

            // 回退到旧的 args 数组格式...
        }
    }
}
```

### 2.3 TypeScript 脚本执行层

使用 `npx tsx` 直接执行 TypeScript，无需编译：

```go
// tool/shell_tool.go

func RunTypeScriptScript(scriptPath string, args []string) (string, error) {
    cmdArgs := append([]string{"tsx", scriptPath}, args...)
    cmd := exec.Command("npx", cmdArgs...)

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        return "", fmt.Errorf("failed to run typescript script: %w\nStdout: %s\nStderr: %s",
            err, stdout.String(), stderr.String())
    }

    return stdout.String() + stderr.String(), nil
}
```

**为什么选择 tsx？**

- ✅ 无需预编译，开发效率高
- ✅ 支持 TypeScript 和 ESM
- ✅ 与 Node.js 生态完全兼容
- ✅ 支持最新的 JS 语法

**注意：** Skills 系统支持多种脚本类型的混合使用：

| 脚本类型 | 执行方式 | 适用场景 |
|---------|---------|---------|
| TypeScript (.ts) | `npx tsx script.ts` | 业务逻辑、图像生成 |
| JavaScript (.js) | `npx tsx script.js` | 简单脚本 |
| Python (.py) | `python script.py` | 数据处理、PDF 操作 |
| Shell (.sh) | `bash script.sh` | 系统操作 |

在本项目中：
- **baoyu-comic** 和 **baoyu-image-gen** 使用 TypeScript
- **pdf** 技能使用 Python（未在漫画生成流程中使用）

## 三、完整工作流程

![智能体工作流程图](cover-image/06-workflow.png)

### 3.1 智能体执行流程图

```
用户输入
   │
   ▼
┌─────────────────────────────────────────────────┐
│  Agent 节点: LLM 规划 + 工具调用                │
│  输入: 用户请求 + 工具定义                       │
│  输出: 结构化工具调用                            │
└────────────────────┬────────────────────────────┘
                     │
                     ▼
          ┌──────────────────────┐
          │  Tools 节点: 执行工具 │
          └──────────────────────┘
                     │
        ┌────────────┴────────────┐
        ▼                         ▼
   分镜生成脚本                图像生成脚本
   (generate-comic.ts)        (main.ts)
        │                         │
        ▼                         ▼
   分镜 JSON 文件            漫画图像文件
        │                         │
        └────────────┬────────────┘
                     ▼
              PDF 合并脚本
           (merge-to-pdf.ts)
                     │
                     ▼
              完整漫画 PDF
```

### 3.2 主程序完整代码

```go
// main.go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/smallnest/goskills"
    adapter "github.com/smallnest/langgraphgo/adapter/goskills"
    "github.com/smallnest/langgraphgo/prebuilt"
    "github.com/tmc/langchaingo/llms"
    "github.com/tmc/langchaingo/llms/openai"
    "github.com/tmc/langchaingo/tools"
)

func main() {
    // 1. 初始化 LLM
    // 推荐使用 ERNIE 5.0 Thinking Preview，工具调用更稳定
    // 如需使用，设置环境变量：
    //   export OPENAI_API_KEY=your-ernie-api-key
    //   export OPENAI_BASE_URL=https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/ernie-5.0-thinking-preview
    llm, err := openai.New()
    if err != nil {
        log.Fatal(err)
    }

    // 2. 从 skills 目录加载技能包
    skillsDir := "./skills"
    if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
        skillsDir = "comic_skill_example/skills"
    }

    packages, err := goskills.ParseSkillPackages(skillsDir)
    if err != nil {
        log.Fatalf("解析技能包失败: %v", err)
    }

    if len(packages) == 0 {
        log.Fatal("在 " + skillsDir + " 中未找到任何技能")
    }

    // 3. 将技能转换为工具（工具配置会从 SKILL.md 自动读取）
    var allTools []tools.Tool

    for _, skill := range packages {
        fmt.Printf("正在加载技能: %s - %s\n", skill.Meta.Name, skill.Meta.Description)

        // 工具配置会从 SKILL.md 的 tools 字段自动读取
        skillTools, err := adapter.SkillsToTools(skill)
        if err != nil {
            log.Printf("转换技能 %s 为工具失败: %v", skill.Meta.Name, err)
            continue
        }

        allTools = append(allTools, skillTools...)
    }

    // 4. 筛选出漫画相关工具
    var comicTools []tools.Tool
    for _, t := range allTools {
        if t.Name() == "generate_comic_storyboard" ||
           t.Name() == "generate_comic_image" ||
           t.Name() == "merge_comic_to_pdf" {
            comicTools = append(comicTools, t)
        }
    }

    // 5. 构建系统提示词
    systemMsg := `你是一个有用的助手，可以访问工具来创作漫画。当用户要求创建漫画时，你必须调用 generate_comic_storyboard 函数。

可用函数：
- generate_comic_storyboard: 创建完整的漫画分镜脚本和提示词
- generate_comic_image: 生成单张漫画图像（需要提示词和路径）
- merge_comic_to_pdf: 将漫画图像合并成 PDF

工作流程：
1. 调用 generate_comic_storyboard 创建漫画分镜
2. 如果输出显示 '=== IMAGE_GENERATION_REQUIRED ==='，则为每一页调用 generate_comic_image
3. 调用 merge_comic_to_pdf 将所有图像合并成 PDF

重要提示：始终调用函数，而不是提供文字描述。`

    // 6. 创建 Agent
    agent, err := prebuilt.CreateAgentMap(llm, comicTools, 20,
        prebuilt.WithSystemMessage(systemMsg),
    )
    if err != nil {
        log.Fatal(err)
    }

    // 7. 执行
    ctx := context.Background()
    resp, err := agent.Invoke(ctx, map[string]any{
        "messages": []llms.MessageContent{
            llms.TextParts(llms.ChatMessageTypeHuman, os.Args[1]),
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    // 8. 输出结果
    if messages, ok := resp["messages"].([]llms.MessageContent); ok {
        for _, msg := range messages {
            fmt.Printf("[%s] %s\n", msg.Role, msg.Parts)
        }
    }
}
```

## 四、踩坑与解决方案

![踩坑排查示意图](cover-image/07-troubleshooting.png)

### 4.1 DeepSeek V3 工具调用不稳定问题

**现象：** DeepSeek V3 返回的工具调用格式不稳定，有时无法正确解析
```
<｜tool▁calls▁begin｜><｜tool▁call▁begin｜>function<｜tool▁sep｜>generate_comic_storyboard
{"topic":"采蘑菇的小姑娘"}
<｜tool▁call▁end｜>
```

**解决方案：** 更换为 ERNIE 5.0 Thinking Preview（文心一言），工具调用更稳定

```go
// 使用千帆平台配置
llm, err := openai.New(
    openai.WithToken("your-ernie-api-key"),
    openai.WithBaseURL("https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat/ernie-5.0-thinking-preview"),
)
```

### 4.2 TypeScript 脚本执行问题

**现象：** Bun API 与 Node.js 不兼容
```
Cannot find package 'bun'
```

**解决方案：**
- 移除 Bun 特定 API（如 `Bun.write`）
- 使用 Node.js 兼容的 API（如 `writeFileSync`）
- 使用 `npx tsx` 替代 `bun run` 执行

### 4.3 参数格式转换问题

**现象：** LLM 返回命名参数，脚本期望命令行参数

**解决方案：** 在工具执行层实现自动转换（见 2.2 节）

### 4.4 中文文件名支持

**现象：** PDF 合并脚本的正则表达式无法匹配中文文件名

**解决方案：**
```typescript
// 添加 Unicode 中文字符范围
const pagePattern = /^(\d+)-(cover|page)(-[\w\u4e00-\u9fff-]+)?\.(png|jpg|jpeg)$/i;
```

### 4.5 调试日志控制

**现象：** `[DEBUG]` 日志过多影响正常输出

**解决方案：** 使用 `WithVerbose()` 选项控制
```go
agent, err := prebuilt.CreateAgentMap(llm, tools, 20,
    prebuilt.WithSystemMessage(systemMsg),
    prebuilt.WithVerbose(true), // 仅在需要时启用
)
```

## 五、最佳实践总结

![最佳实践总结图](cover-image/08-best-practices.png)

### 5.1 技能设计原则

1. **单一职责** - 每个技能专注一个领域（分镜、图像、PDF）
2. **声明式配置** - 工具定义在 SKILL.md 中，而非硬编码
3. **语言选型** - 脚本层使用 TypeScript/Python，发挥各自优势

### 5.2 错误处理

```typescript
// 脚本中要提供清晰的错误信息
async function main() {
    try {
        // ...
    } catch (error) {
        console.error("Error:", error);
        console.error("Error message:", error?.message);
        process.exit(1);
    }
}
```

### 5.3 扩展性考虑

当需要添加新工具时：

1. 在 `SKILL.md` 中添加工具定义
2. 在 `scripts/` 目录添加对应脚本
3. 无需修改任何 Go 代码

**示例：添加"水印添加"工具**

```yaml
# SKILL.md
tools:
  - name: add_watermark
    script: scripts/add-watermark.ts
    description: 为漫画添加水印
    parameters:
      image:
        type: string
        required: true
      watermark:
        type: string
        required: true
```

```typescript
// scripts/add-watermark.ts
// 实现水印逻辑
```

## 六、性能优化建议

### 6.1 并发图像生成

```go
// 并发生成所有页面
var wg sync.WaitGroup
semaphore := make(chan struct{}, 3) // 限制并发数

for _, page := range pages {
    wg.Add(1)
    go func(p Page) {
        defer wg.Done()
        semaphore <- struct{}{}        // 获取信号量
        defer func() { <-semaphore }() // 释放信号量

        generateImage(p)
    }(page)
}
wg.Wait()
```

### 6.2 缓存机制

```go
// 对相同参数的请求使用缓存
type CacheKey struct {
    Topic string
    Style string
    Pages int
}

var storyboardCache = sync.Map{}
```

## 七、未来展望

### 7.1 可能的改进方向

1. **多模态输入** - 支持图片、视频作为创作素材
2. **风格迁移** - 一键切换漫画风格
3. **交互式编辑** - 支持用户在生成过程中介入调整
4. **分布式部署** - 将图像生成等耗时任务分布到多台机器

### 7.2 社区生态

欢迎贡献：
- 新的技能插件
- 性能优化方案
- 文档改进
- Bug 修复

## 八、参考资料

- [LangGraphGo GitHub](https://github.com/smallnest/langgraphgo) - Go 语言的 LangGraph 实现
- [GoSkills v0.6.1+](https://github.com/smallnest/goskills) - 技能插件系统（支持 SKILL.md 工具定义）
- [LangChain 中文文档](https://www.langchain.com.cn/)
- [DeepSeek API 文档](https://platform.deepseek.com/api-docs/)

---

**作者简介：** 鸟窝，专注于 Go 语言、LLM 应用架构设计。本文所有代码已开源，欢迎 Star ⭐

代码：  [langgraphgo comic skill example](https://github.com/smallnest/langgraphgo/tree/master/examples/comic_skill_example)