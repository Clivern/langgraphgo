# Markdown Tool Documentation

This package provides specialized tools for working with Markdown files, particularly designed for the [Manus-style planning workflow](https://github.com/OthmanAdi/planning-with-files).

## Overview

The Markdown tool provides functions for:

- ✅ **Reading Markdown files** with frontmatter support
- ✅ **Writing Markdown files** with optional frontmatter
- ✅ **Updating checkboxes** in task plans
- ✅ **Appending to sections** in organized notes
- ✅ **Logging errors** with timestamps
- ✅ **Parsing task plans** to extract goals and phases
- ✅ **Generating task plans** from structured data

## Core Types

### MarkdownFile

```go
type MarkdownFile struct {
    Path        string
    Frontmatter map[string]any
    Content     string
}
```

Represents a Markdown file with optional YAML frontmatter.

### TaskPhase

```go
type TaskPhase struct {
    Number      int
    Name        string
    Description string
    Node        string
    Complete    bool
}
```

Represents a single phase in a task plan workflow.

## API Reference

### Reading Markdown

#### ReadMarkdown

```go
mf, err := ReadMarkdown("task_plan.md")
if err != nil {
    log.Fatal(err)
}

fmt.Println(mf.Content)
fmt.Println(mf.Frontmatter)
```

Reads a Markdown file and parses frontmatter (YAML between `---` delimiters).

### Writing Markdown

#### WriteMarkdown

```go
content := "# My Document\n\nContent goes here."
frontmatter := map[string]any{
    "title": "My Document",
    "date":  "2025-01-07",
}

err := WriteMarkdown("output.md", content, frontmatter)
```

Writes content to a Markdown file with optional frontmatter.

### Task Plan Operations

#### CreateTaskPlan

```go
phases := []TaskPhase{
    {
        Number:      1,
        Name:        "Research",
        Description: "Search for information",
        Node:        "research",
    },
    {
        Number:      2,
        Name:        "Write",
        Description: "Write deliverable",
        Node:        "write",
    },
}

err := CreateTaskPlan("task_plan.md", "Research TypeScript", phases)
```

Creates a new task plan Markdown file.

#### ParseTaskPlan

```go
goal, phases, err := ParseTaskPlan("task_plan.md")
fmt.Println("Goal:", goal)
for _, phase := range phases {
    fmt.Printf("Phase %d: %s (complete: %v)\n",
        phase.Number, phase.Name, phase.Complete)
}
```

Parses a task plan file to extract goal and phases.

#### UpdatePhaseStatus

```go
// Mark phase as complete
err := UpdatePhaseStatus("task_plan.md", "Research", true)
```

Updates the checkbox status of a specific phase.

#### GetCompletedPhases / GetPendingPhases

```go
completed, err := GetCompletedPhases("task_plan.md")
pending, err := GetPendingPhases("task_plan.md")

fmt.Println("Completed:", completed)
fmt.Println("Pending:", pending)
```

Returns lists of completed or pending phase names.

### Section Operations

#### AppendToMarkdownSection

```go
err := AppendToMarkdownSection("notes.md", "Research",
    "New research finding")
```

Appends content to a specific section in a Markdown file.

#### ExtractSectionContent

```go
content, err := ExtractSectionContent("notes.md", "Research")
fmt.Println(content)
```

Extracts all content from a specific section.

### Error Logging

#### LogErrorToMarkdown

```go
err := LogErrorToMarkdown("notes.md", "Connection timeout")
```

Logs an error with timestamp to a Markdown file.

### Checkbox Operations

#### UpdateMarkdownCheckboxes

```go
updates := map[string]bool{
    "Research": true,
    "Write":    false,
}

updatedContent, err := UpdateMarkdownCheckboxes("task_plan.md", updates)
```

Updates multiple checkboxes in a task plan.

## Usage Examples

### Example 1: Create and Update a Task Plan

```go
package main

import (
    "fmt"
    "github.com/smallnest/langgraphgo/tool"
)

func main() {
    // Define phases
    phases := []tool.TaskPhase{
        {Number: 1, Name: "Research", Description: "Research phase", Node: "research"},
        {Number: 2, Name: "Write", Description: "Write phase", Node: "write"},
        {Number: 3, Name: "Review", Description: "Review phase", Node: "review"},
    }

    // Create task plan
    err := tool.CreateTaskPlan("task_plan.md",
        "Research and document TypeScript benefits", phases)
    if err != nil {
        panic(err)
    }

    fmt.Println("Task plan created")

    // Update phase status
    err = tool.UpdatePhaseStatus("task_plan.md", "Research", true)
    if err != nil {
        panic(err)
    }

    // Check status
    completed, _ := tool.GetCompletedPhases("task_plan.md")
    pending, _ := tool.GetPendingPhases("task_plan.md")

    fmt.Printf("Completed: %v\n", completed)
    fmt.Printf("Pending: %v\n", pending)
}
```

### Example 2: Manage Research Notes

```go
// Create notes file
err := tool.WriteMarkdown("notes.md", "# Research Notes\n", nil)

// Add research section
err = tool.AppendToMarkdownSection("notes.md", "Research",
    "TypeScript provides type safety")

// Add findings
err = tool.AppendToMarkdownSection("notes.md", "Research",
    "Better IDE support with autocomplete")

// Log an error
err = tool.LogErrorToMarkdown("notes.md", "API rate limit reached")
```

### Example 3: Parse and Display Task Plan

```go
// Read task plan
goal, phases, err := tool.ParseTaskPlan("task_plan.md")
if err != nil {
    panic(err)
}

fmt.Printf("📋 Goal: %s\n\n", goal)
fmt.Println("Phases:")

for _, phase := range phases {
    status := "✅"
    if !phase.Complete {
        status = "⬜"
    }
    fmt.Printf("  %s Phase %d: %s\n", status, phase.Number, phase.Name)
    fmt.Printf("      Description: %s\n", phase.Description)
    fmt.Printf("      Node: %s\n", phase.Node)
}
```

## Task Plan Format

The tool expects task plans in this format:

```markdown
%% Goal

Your goal description here

%% Phases

- [ ] Phase 1: Phase Name
  Description: What this phase does
  Node: node_name

- [x] Phase 2: Another Phase
  Description: What this phase does
  Node: another_node
```

## Integration with Manus Agent

The Markdown tool is used by the `CreateManusAgent` function in `prebuilt/manus_planning_agent.go`:

```go
// Create task plan
planText := tool.GenerateTaskPlanMarkdown(goal, phases)
err := tool.WriteFile(config.PlanPath, planText)

// Update checkboxes
err = tool.UpdatePhaseStatus(config.PlanPath, phaseName, true)

// Log errors
err = tool.LogErrorToMarkdown(config.NotesPath, errMsg)
```

## Best Practices

### 1. Always Check Errors

```go
if err := tool.UpdatePhaseStatus(path, "Research", true); err != nil {
    log.Printf("Warning: failed to update phase: %v", err)
    // Continue execution or handle error
}
```

### 2. Use Structured Data

```go
phases := []tool.TaskPhase{
    {Number: 1, Name: "Research", Node: "research"},
    {Number: 2, Name: "Write", Node: "write"},
}

// Generate consistent markdown
content := tool.GenerateTaskPlanMarkdown(goal, phases)
```

### 3. Validate Before Use

```go
// Check if plan exists before parsing
if _, err := os.Stat("task_plan.md"); err == nil {
    goal, phases, err := tool.ParseTaskPlan("task_plan.md")
    // Use parsed data
}
```

## Testing

The package includes comprehensive tests:

```bash
go test -v ./tool/ -run Markdown
```

Tests cover:
- Reading and writing Markdown files
- Frontmatter parsing
- Checkbox updates
- Task plan parsing and generation
- Section operations
- Error logging

## File Permissions

All file operations use secure permissions:

- Write operations: `0600` (read/write for owner only)
- Files are created with restricted permissions by default

## Error Handling

Functions return errors that can be checked and handled:

```go
err := tool.UpdatePhaseStatus("task_plan.md", "Research", true)
if err != nil {
    if os.IsNotExist(err) {
        // File doesn't exist, create it first
    } else if errors.Is(err, os.ErrPermission) {
        // Permission error
    }
}
```

## Performance Considerations

- Files are read entirely into memory
- For large files, consider streaming or chunked processing
- Checkbox updates rewrite the entire file
- Use batch updates when possible

## Thread Safety

The package is not thread-safe for concurrent writes to the same file. Use external synchronization if needed.

## See Also

- [prebuilt.CreateManusAgent](../prebuilt/manus_planning_agent.go) - Manus Agent implementation
- [planning-with-files](https://github.com/OthmanAdi/planning-with-files) - Original inspiration
- [tool/file_tool.go](./file_tool.go) - Basic file operations

## License

MIT License - see LICENSE file for details
