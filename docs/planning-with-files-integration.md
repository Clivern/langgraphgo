# Planning-with-Files Integration for LangGraphGo

## 📋 Summary

LangGraphGo **fully supports** the [planning-with-files](https://github.com/OthmanAdi/planning-with-files) workflow pattern. We've created a new `CreateManusAgent` function that combines the best of both approaches:

- ✅ **Persistent Markdown files** (task_plan.md, notes.md)
- ✅ **LLM-driven dynamic planning**
- ✅ **Automatic checkpointing and recovery**
- ✅ **Progress tracking with checkboxes**
- ✅ **Error logging and recovery**

---

## 🎯 What is planning-with-files?

Planning-with-files is a Claude Code skill that implements the **Manus AI** workflow pattern - the secret behind Manus's $2 billion acquisition by Meta.

### Core Principles

1. **Filesystem as Memory** - Store planning state in persistent Markdown files
2. **Attention Manipulation** - Re-read plan before decisions to stay focused
3. **Error Persistence** - Log failures to avoid repeating mistakes
4. **Goal Tracking** - Visual checkboxes show progress
5. **Append-Only Context** - Never modify history, only add to it

### Three-File Pattern

```
work/
├── task_plan.md     # Track phases and progress
├── notes.md         # Store research and findings
└── output.md        # Final deliverable
```

---

## 🔄 Comparison: LangGraphGo Approaches

| Feature | CreatePlanningAgent | CreateManusAgent (NEW) |
|---------|---------------------|----------------------|
| **Storage** | In-memory State | Markdown files + State |
| **Planning** | LLM generates JSON | LLM generates Markdown |
| **Progress** | Message history | Checkboxes in task_plan.md |
| **Errors** | In messages | Logged to notes.md |
| **Recovery** | Checkpoint auto-resume | Checkpoint + file resume |
| **Human Edit** | Via UpdateState | Edit task_plan.md directly |
| **Best For** | Quick automated tasks | Complex multi-step research |

---

## 🚀 Usage Example

```go
package main

import (
    "context"
    "github.com/smallnest/langgraphgo/graph"
    "github.com/smallnest/langgraphgo/prebuilt"
    "github.com/tmc/langchaingo/llms"
)

func main() {
    // 1. Define available nodes
    nodes := []graph.TypedNode[map[string]any]{
        {
            Name:        "research",
            Description: "Research and gather information",
            Function:    researchNode,
        },
        {
            Name:        "compile",
            Description: "Compile findings into notes",
            Function:    compileNode,
        },
        {
            Name:        "write",
            Description: "Write final deliverable",
            Function:    writeNode,
        },
    }

    // 2. Configure Manus agent
    config := prebuilt.ManusConfig{
        WorkDir:    "./work",
        PlanPath:   "./work/task_plan.md",
        NotesPath:  "./work/notes.md",
        OutputPath: "./work/output.md",
        AutoSave:   true,
        Verbose:    true,
    }

    // 3. Create the agent
    agent, err := prebuilt.CreateManusAgent(
        model,
        nodes,
        []tools.Tool{},
        config,
    )
    if err != nil {
        panic(err)
    }

    // 4. Execute
    initialState := map[string]any{
        "messages": []llms.MessageContent{
            {
                Role: llms.ChatMessageTypeHuman,
                Parts: []llms.ContentPart{
                    llms.TextPart("Research TypeScript benefits and write a summary"),
                },
            },
        },
        "goal": "Research and document TypeScript benefits",
    }

    result, err := agent.Invoke(context.Background(), initialState)
    if err != nil {
        panic(err)
    }

    // 5. Check generated files
    // work/task_plan.md - Shows phases with checkboxes
    // work/notes.md - Contains research findings
    // work/output.md - Final deliverable
}
```

---

## 📁 Generated File Format

### task_plan.md

```markdown
%% Goal
Research and document the benefits of TypeScript for development teams

%% Phases
- [x] Phase 1: Research TypeScript Benefits
  Description: Search for and analyze TypeScript documentation
  Node: research

- [x] Phase 2: Compile Findings
  Description: Organize research findings into notes.md
  Node: compile

- [ ] Phase 3: Write Summary
  Description: Generate final markdown summary
  Node: write
```

### notes.md

```markdown
# Research Notes

## TypeScript Benefits
- Type safety
- Better IDE support
- Easier refactoring

## Error Log
[Any errors encountered during execution]
```

### output.md

```markdown
# Final Output

Generated at: 2025-01-07 15:30:45

[Final deliverable content...]
```

---

## 🎨 Key Features

### 1. Persistent Planning

Plans are automatically saved to `task_plan.md` with progress tracking via checkboxes:

```go
- [x] Phase 1: Complete  ✓
- [ ] Phase 2: Pending
- [ ] Phase 3: Pending
```

### 2. Error Logging

Errors are automatically logged to `notes.md` with timestamps:

```markdown
## Error [2025-01-07 15:30:45]
Error in phase 2 (compile): connection timeout
```

### 3. Resume Capability

The agent can resume from where it left off:

```go
// Agent automatically:
// 1. Reads task_plan.md
// 2. Parses completed phases
// 3. Continues from pending phase
```

### 4. Human-in-the-Loop

Enable manual intervention:

```go
agent.InterruptBefore([]string{"planner"})

// User can edit task_plan.md to adjust plan
// Agent will read updated plan on resume
```

---

## 🔧 Advanced Configuration

```go
config := prebuilt.ManusConfig{
    // Working directory (auto-created)
    WorkDir:    "./my-work",

    // File paths (relative to WorkDir)
    PlanPath:   "./my-work/task_plan.md",
    NotesPath:  "./my-work/notes.md",
    OutputPath: "./my-work/output.md",

    // Auto-save plans after each phase
    AutoSave:   true,

    // Verbose logging
    Verbose:    true,
}
```

---

## 📚 Additional Resources

- [Original planning-with-files repository](https://github.com/OthmanAdi/planning-with-files)
- [LangGraphGo documentation](https://github.com/smallnest/langgraphgo)
- [Manus AI acquisition news](https://www.techcrunch.com/2025/01/meta-acquires-manus-2b)

---

## ✅ Implementation Checklist

- [x] Create `CreateManusAgent` function
- [x] Implement Markdown plan parsing
- [x] Add checkbox progress tracking
- [x] Implement error logging to notes.md
- [x] Add automatic plan saving
- [x] Create test suite
- [x] Write documentation
- [ ] Add integration examples
- [ ] Performance benchmarking
- [ ] Add streaming support

---

## 🎓 When to Use

### Use Manus Agent for:

- ✅ **Multi-step research tasks** (3+ phases)
- ✅ **Documentation projects**
- ✅ **Long-running workflows**
- ✅ **Tasks requiring manual review**
- ✅ **Projects with knowledge accumulation**

### Use Standard Planning Agent for:

- ⚡ **Quick automation tasks**
- ⚡ **API workflows**
- ⚡ **Data processing pipelines**
- ⚡ **Simple sequential tasks**

---

## 🙏 Acknowledgments

- [OthmanAdi](https://github.com/OthmanAdi) for creating [planning-with-files](https://github.com/OthmanAdi/planning-with-files)
- [Manus AI](https://www.manus.ai) for pioneering context engineering patterns
- [Anthropic](https://www.anthropic.com) for Claude Code and the Agent Skills framework

---

**License**: MIT (same as LangGraphGo)
