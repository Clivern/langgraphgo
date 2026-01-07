package prebuilt_test

import (
	"context"
	"fmt"
	"os"

	"github.com/smallnest/langgraphgo/graph"
	"github.com/smallnest/langgraphgo/prebuilt"
	"github.com/tmc/langchaingo/llms"
)

// Example_manusAgent demonstrates how to use CreateManusAgent
// with persistent Markdown files for planning and tracking
func Example_manusAgent() {
	// Define available nodes that can be used in the plan
	nodes := []graph.TypedNode[map[string]any]{
		{
			Name:        "research",
			Description: "Research and gather information from external sources",
			Function: func(ctx context.Context, state map[string]any) (map[string]any, error) {
				messages := state["messages"].([]llms.MessageContent)

				fmt.Println("🔍 Researching...")

				msg := llms.MessageContent{
					Role:  llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.TextPart("Research complete: Found 15 relevant sources")},
				}

				return map[string]any{
					"messages": append(messages, msg),
				}, nil
			},
		},
		{
			Name:        "compile",
			Description: "Compile findings into organized notes",
			Function: func(ctx context.Context, state map[string]any) (map[string]any, error) {
				messages := state["messages"].([]llms.MessageContent)

				fmt.Println("📝 Compiling findings...")

				msg := llms.MessageContent{
					Role:  llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.TextPart("Notes compiled: 5 key findings organized")},
				}

				return map[string]any{
					"messages": append(messages, msg),
				}, nil
			},
		},
		{
			Name:        "write",
			Description: "Write final deliverable based on research",
			Function: func(ctx context.Context, state map[string]any) (map[string]any, error) {
				messages := state["messages"].([]llms.MessageContent)

				fmt.Println("✍️  Writing final output...")

				msg := llms.MessageContent{
					Role:  llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.TextPart("Final document written: 2000 words summary")},
				}

				return map[string]any{
					"messages": append(messages, msg),
				}, nil
			},
		},
		{
			Name:        "review",
			Description: "Review and validate the output",
			Function: func(ctx context.Context, state map[string]any) (map[string]any, error) {
				messages := state["messages"].([]llms.MessageContent)

				fmt.Println("✅ Reviewing...")

				msg := llms.MessageContent{
					Role:  llms.ChatMessageTypeAI,
					Parts: []llms.ContentPart{llms.TextPart("Review complete: Output validated successfully")},
				}

				return map[string]any{
					"messages": append(messages, msg),
				}, nil
			},
		},
	}

	// Configure the Manus agent
	_ = prebuilt.ManusConfig{
		WorkDir:    "./work",
		PlanPath:   "./work/task_plan.md",
		NotesPath:  "./work/notes.md",
		OutputPath: "./work/output.md",
		AutoSave:   true,
		Verbose:    true,
	}

	// Create initial state with user request
	_ = map[string]any{
		"messages": []llms.MessageContent{
			{
				Role:  llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{llms.TextPart("Research TypeScript benefits and write a summary")},
			},
		},
		"goal": "Research and document the benefits of TypeScript for development teams",
	}

	fmt.Println("🚀 Manus Agent Example")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("This example demonstrates:")
	fmt.Println("1. Persistent Markdown planning (task_plan.md)")
	fmt.Println("2. Research notes storage (notes.md)")
	fmt.Println("3. Progress tracking with checkboxes")
	fmt.Println("4. Final output generation (output.md)")
	fmt.Println()
	fmt.Println("Available nodes:")
	for i, node := range nodes {
		fmt.Printf("  %d. %s: %s\n", i+1, node.Name, node.Description)
	}
	fmt.Println()
	fmt.Println("Generated files:")
	fmt.Println("  📄 task_plan.md - Workflow plan with progress checkboxes")
	fmt.Println("  📄 notes.md - Research findings and error logs")
	fmt.Println("  📄 output.md - Final deliverable")
	fmt.Println()

	// Clean up work directory for demo
	os.RemoveAll("./work")
}

// Example_manusAgentWithErrors shows error handling and recovery
func Example_manusAgentWithErrors() {
	fmt.Println("🔄 Manus Agent with Error Handling")
	fmt.Println("===================================")
	fmt.Println()
	fmt.Println("The Manus agent handles errors by:")
	fmt.Println("1. Logging errors to notes.md")
	fmt.Println("2. Updating checkboxes in task_plan.md")
	fmt.Println("3. Maintaining state for recovery")
	fmt.Println()
	fmt.Println("Example error flow:")
	fmt.Println("  ❌ Phase 2 fails → error logged to notes.md")
	fmt.Println("  📋 task_plan.md shows Phase 1 complete, Phase 2 pending")
	fmt.Println("  🔄 Agent can resume and retry Phase 2")
	fmt.Println()
	fmt.Println("Error logging format in notes.md:")
	fmt.Println("  ## Error [2025-01-07 15:30:45]")
	fmt.Println("  Error in phase 2 (compile): connection timeout")
}

// Example_manusAgentFileStructure shows the file structure
func Example_manusAgentFileStructure() {
	fmt.Println("📁 Manus Agent File Structure")
	fmt.Println("=============================")
	fmt.Println()
	fmt.Println("work/")
	fmt.Println("├── task_plan.md          # Workflow plan with checkboxes")
	fmt.Println("│   %% Goal")
	fmt.Println("│   Research TypeScript benefits")
	fmt.Println("│   ")
	fmt.Println("│   %% Phases")
	fmt.Println("│   - [x] Phase 1: Research")
	fmt.Println("│   - [ ] Phase 2: Compile")
	fmt.Println("│   - [ ] Phase 3: Write")
	fmt.Println("│")
	fmt.Println("├── notes.md              # Research findings & errors")
	fmt.Println("│   # Research Notes")
	fmt.Println("│   ")
	fmt.Println("│   ## TypeScript Benefits")
	fmt.Println("│   - Type safety")
	fmt.Println("│   - Better IDE support")
	fmt.Println("│   ")
	fmt.Println("│   ## Error Log")
	fmt.Println("│   [Error entries here]")
	fmt.Println("│")
	fmt.Println("└── output.md             # Final deliverable")
	fmt.Println("    # TypeScript Benefits Summary")
	fmt.Println("    ...")
	fmt.Println()
}

// Example_manusVsPlanningAgent compares both approaches
func Example_manusVsPlanningAgent() {
	fmt.Println("📊 Manus Agent vs Planning Agent")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("Planning Agent (prebuilt.CreatePlanningAgent):")
	fmt.Println("  ✅ Dynamic workflow generation")
	fmt.Println("  ✅ JSON-based plan format")
	fmt.Println("  ✅ In-memory state management")
	fmt.Println("  ✅ Fast execution")
	fmt.Println()
	fmt.Println("Manus Agent (prebuilt.CreateManusAgent):")
	fmt.Println("  ✅ Persistent Markdown files")
	fmt.Println("  ✅ Human-readable plans")
	fmt.Println("  ✅ Progress tracking with checkboxes")
	fmt.Println("  ✅ Error logging to notes.md")
	fmt.Println("  ✅ Resume capability")
	fmt.Println("  ✅ Knowledge accumulation")
	fmt.Println()
	fmt.Println("When to use:")
	fmt.Println("  • Planning Agent - Quick tasks, automated workflows")
	fmt.Println("  • Manus Agent - Complex multi-step tasks, research, documentation")
	fmt.Println()
}

// Example_manusAgentIntegration shows real usage pattern
func Example_manusAgentIntegration() {
	fmt.Println("💻 Integration Example")
	fmt.Println("=====================")
	fmt.Println()
	fmt.Println("// 1. Define your nodes")
	fmt.Println("nodes := []graph.TypedNode[map[string]any]{")
	fmt.Println("    {Name: \"research\", Description: \"...\", Function: ...},")
	fmt.Println("    {Name: \"compile\", Description: \"...\", Function: ...},")
	fmt.Println("    {Name: \"write\", Description: \"...\", Function: ...},")
	fmt.Println("}")
	fmt.Println()
	fmt.Println("// 2. Configure Manus agent")
	fmt.Println("config := prebuilt.ManusConfig{")
	fmt.Println("    WorkDir:    \"./work\",")
	fmt.Println("    PlanPath:   \"./work/task_plan.md\",")
	fmt.Println("    NotesPath:  \"./work/notes.md\",")
	fmt.Println("    OutputPath: \"./work/output.md\",")
	fmt.Println("    AutoSave:   true,")
	fmt.Println("    Verbose:    true,")
	fmt.Println("}")
	fmt.Println()
	fmt.Println("// 3. Create the agent")
	fmt.Println("agent, err := prebuilt.CreateManusAgent(")
	fmt.Println("    model,")
	fmt.Println("    nodes,")
	fmt.Println("    []tools.Tool{},")
	fmt.Println("    config,")
	fmt.Println(")")
	fmt.Println()
	fmt.Println("// 4. Execute")
	fmt.Println("result, err := agent.Invoke(ctx, initialState)")
	fmt.Println()
	fmt.Println("// 5. Check results in work/")
	fmt.Println("//    - task_plan.md shows progress")
	fmt.Println("//    - notes.md contains research")
	fmt.Println("//    - output.md has final deliverable")
	fmt.Println()
}
