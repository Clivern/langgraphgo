package main

import (
	"context"
	"encoding/json"
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

// Ensure comicToolWrapper implements prebuilt.ToolWithSchema
var _ prebuilt.ToolWithSchema = (*comicToolWrapper)(nil)

// comicToolWrapper wraps a tool with a more friendly name and description
type comicToolWrapper struct {
	tool      tools.Tool
	newName   string
	newDesc   string
	schema    map[string]any
	toolType  string // "generate_comic_storyboard", "generate_comic_image", or "merge_comic_to_pdf"
}

func (w *comicToolWrapper) Name() string        { return w.newName }
func (w *comicToolWrapper) Description() string { return w.newDesc }
func (w *comicToolWrapper) Call(ctx context.Context, input string) (string, error) {
	// Convert JSON input to command-line args format
	if w.toolType == "generate_comic_storyboard" {
		var params struct {
			Topic string `json:"topic"`
			Style string `json:"style"`
			Pages int    `json:"pages"`
		}
		if err := json.Unmarshal([]byte(input), &params); err == nil {
			// Convert to command-line args
			args := []string{"--topic", params.Topic}
			if params.Style != "" {
				args = append(args, "--style", params.Style)
			}
			if params.Pages > 0 {
				args = append(args, "--pages", fmt.Sprintf("%d", params.Pages))
			}
			argsJSON, _ := json.Marshal(map[string]any{"args": args})
			return w.tool.Call(ctx, string(argsJSON))
		}
	} else if w.toolType == "generate_comic_image" {
		var params struct {
			Prompt string `json:"prompt"`
			Path   string `json:"path"`
		}
		if err := json.Unmarshal([]byte(input), &params); err == nil {
			args := []string{"--prompt", params.Prompt, "--image", params.Path}
			argsJSON, _ := json.Marshal(map[string]any{"args": args})
			return w.tool.Call(ctx, string(argsJSON))
		}
	} else if w.toolType == "merge_comic_to_pdf" {
		var params struct {
			Directory string `json:"directory"`
		}
		if err := json.Unmarshal([]byte(input), &params); err == nil {
			argsJSON, _ := json.Marshal(map[string]any{"args": []string{params.Directory}})
			return w.tool.Call(ctx, string(argsJSON))
		}
	}
	// Fallback to original input
	return w.tool.Call(ctx, input)
}

// Schema returns a custom JSON schema for this tool
func (w *comicToolWrapper) Schema() map[string]any {
	if w.schema != nil {
		return w.schema
	}
	return nil
}

func main() {
	// 1. Initialize LLM
	if os.Getenv("OPENAI_API_KEY") == "" {
		log.Fatal("OPENAI_API_KEY is not set")
	}

	// Configure LLM with support for custom base URL (e.g., Baidu Qianfan)
	model := os.Getenv("OPENAI_API_MODEL")
	if model == "" {
		model = os.Getenv("OPENAI_MODEL")
	}
	if model == "" {
		model = "gpt-4o"
	}

	var opts []openai.Option
	opts = append(opts, openai.WithModel(model))

	// Support for custom OpenAI-compatible APIs (e.g., Baidu Qianfan)
	if baseURL := os.Getenv("OPENAI_API_BASE"); baseURL != "" {
		opts = append(opts, openai.WithBaseURL(baseURL))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Load Skills from the skills directory
	skillsDir := "./skills"
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		skillsDir = "comic_skill_example/skills"
	}

	packages, err := goskills.ParseSkillPackages(skillsDir)
	if err != nil {
		log.Fatalf("Failed to parse skill packages: %v", err)
	}

	if len(packages) == 0 {
		log.Fatal("No skills found in " + skillsDir)
	}

	// 3. Convert Skills to Tools
	var allTools []tools.Tool
	var allSystemMessages strings.Builder

	// Tool name remapping for better LLM understanding
	toolNameMap := map[string]string{
		"run_scripts_generate_comic_ts": "generate_comic_storyboard",
		"run_scripts_main_ts":            "generate_comic_image",
		"run_scripts_merge_to_pdf_ts":    "merge_comic_to_pdf",
	}

	// Tool schemas for better LLM understanding
	// Use a simpler format that's more compatible with OpenAI's function calling
	toolSchemas := map[string]map[string]any{
		"generate_comic_storyboard": {
			"type": "object",
			"properties": map[string]any{
				"topic": map[string]any{
					"type":        "string",
					"description": "The topic of the comic to create",
				},
				"style": map[string]any{
					"type":        "string",
					"description": "Visual style (e.g., warm, classic, dramatic)",
				},
				"pages": map[string]any{
					"type":        "integer",
					"description": "Number of pages to generate",
				},
			},
			"required": []string{"topic"},
		},
		"generate_comic_image": {
			"type": "object",
			"properties": map[string]any{
				"prompt": map[string]any{
					"type":        "string",
					"description": "Image generation prompt",
				},
				"path": map[string]any{
					"type":        "string",
					"description": "Output file path",
				},
			},
			"required": []string{"prompt", "path"},
		},
		"merge_comic_to_pdf": {
			"type": "object",
			"properties": map[string]any{
				"directory": map[string]any{
					"type":        "string",
					"description": "Path to the comic directory",
				},
			},
			"required": []string{"directory"},
		},
	}

	allSystemMessages.WriteString("You are a helpful assistant with access to tools. When users ask to create a comic, you MUST call the generate_comic_storyboard function.\n\n")
	allSystemMessages.WriteString("Available functions:\n")
	allSystemMessages.WriteString("- generate_comic_storyboard: Creates a complete comic storyboard with prompts\n")
	allSystemMessages.WriteString("- generate_comic_image: Generates a single comic image (requires prompt and path)\n")
	allSystemMessages.WriteString("- merge_comic_to_pdf: Merges comic images into a PDF\n\n")
	allSystemMessages.WriteString("Workflow:\n")
	allSystemMessages.WriteString("1. Call generate_comic_storyboard to create the comic storyboard\n")
	allSystemMessages.WriteString("2. If the output shows '=== IMAGE_GENERATION_REQUIRED ===', call generate_comic_image for each page\n")
	allSystemMessages.WriteString("3. Call merge_comic_to_pdf to merge all images into a PDF\n\n")
	allSystemMessages.WriteString("CRITICAL: Always call functions instead of providing text descriptions.\n")

	for _, skill := range packages {
		fmt.Printf("Loading skill: %s - %s\n", skill.Meta.Name, skill.Meta.Description)

		skillTools, err := adapter.SkillsToTools(skill)
		if err != nil {
			log.Printf("Failed to convert skill %s to tools: %v", skill.Meta.Name, err)
			continue
		}

		for _, t := range skillTools {
			// Wrap tool with better name if in remap
			if newName, exists := toolNameMap[t.Name()]; exists {
				// Create a wrapper tool with the better name
				schema := toolSchemas[newName]
				wrappedTool := &comicToolWrapper{
					tool:     t,
					newName:  newName,
					newDesc:  t.Description(),
					schema:   schema,
					toolType: newName,
				}
				allTools = append(allTools, wrappedTool)
				fmt.Printf("  - Tool: %s (was: %s)\n", newName, t.Name())
			} else {
				allTools = append(allTools, t)
				fmt.Printf("  - Tool: %s\n", t.Name())
			}
		}
	}

	if len(allTools) == 0 {
		log.Fatal("No tools found from skills")
	}

	fmt.Printf("\nTotal tools loaded: %d\n\n", len(allTools))

	// 4. Create Agent with all skills
	// For debugging, let's filter to only use the comic generation tool
	var comicTools []tools.Tool
	for _, t := range allTools {
		if t.Name() == "generate_comic_storyboard" || t.Name() == "generate_comic_image" || t.Name() == "merge_comic_to_pdf" {
			comicTools = append(comicTools, t)
		}
	}

	if len(comicTools) == 0 {
		log.Fatal("Comic tools not found")
	}

	fmt.Printf("Using %d comic tools\n", len(comicTools))

	// Debug: print tool definitions
	fmt.Println("\n=== Tool Definitions ===")
	for _, t := range comicTools {
		fmt.Printf("Tool: %s\n", t.Name())
		fmt.Printf("  Description: %s\n", t.Description())
		// Check if tool implements Schema
		if st, ok := t.(interface{ Schema() map[string]any }); ok {
			if schema := st.Schema(); schema != nil {
				fmt.Printf("  Has Schema: YES\n")
			} else {
				fmt.Printf("  Has Schema: NO (nil)\n")
			}
		} else {
			fmt.Printf("  Has Schema: NO (interface mismatch)\n")
		}
	}
	fmt.Println("=== End Tool Definitions ===\n")

	// The agent will use the LLM to decide which tools to call and in what order
	// Increase maxIterations to allow for multiple tool calls (generate_comic -> multiple image gens -> merge)
	systemMsgStr := allSystemMessages.String()
	fmt.Printf("\n=== System Message ===\n%s\n=== End System Message ===\n\n", systemMsgStr)

	agent, err := prebuilt.CreateAgentMap(llm, comicTools, 20,
		prebuilt.WithSystemMessage(systemMsgStr),
	)
	if err != nil {
		log.Fatal(err)
	}

	// 5. Parse command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <comic description>")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  go run main.go \"Create a comic about a little girl picking mushrooms in the forest\"")
		fmt.Println()
		fmt.Println("Available skills:")
		for _, skill := range packages {
			fmt.Printf("  - %s: %s\n", skill.Meta.Name, skill.Meta.Description)
		}
		os.Exit(1)
	}

	input := strings.Join(os.Args[1:], " ")

	// 6. Run Agent
	fmt.Printf("🎨 Creating comic with agent...\n")
	fmt.Printf("📝 Request: %s\n\n", input)

	ctx := context.Background()
	resp, err := agent.Invoke(ctx, map[string]any{
		"messages": []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, input),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// 7. Print result
	fmt.Println("\n===========================================")
	fmt.Println("Agent Response:")
	fmt.Println("===========================================")

	if messages, ok := resp["messages"].([]llms.MessageContent); ok && len(messages) > 0 {
		for i, msg := range messages {
			fmt.Printf("\n[Message %d - Role: %s]\n", i+1, msg.Role)
			for j, part := range msg.Parts {
				switch p := part.(type) {
				case llms.TextContent:
					fmt.Printf("  [Part %d - Text]: %s\n", j+1, string(p.Text))
				case llms.ToolCall:
					fmt.Printf("  [Part %d - ToolCall]: %s\n", j+1, p.FunctionCall.Name)
					fmt.Printf("    Arguments: %s\n", p.FunctionCall.Arguments)
				case llms.ToolCallResponse:
					fmt.Printf("  [Part %d - ToolResponse]: %s\n", j+1, p)
				default:
					fmt.Printf("  [Part %d - Unknown]: %v\n", j+1, part)
				}
			}
		}
	} else {
		fmt.Printf("Response: %v\n", resp)
	}

	fmt.Println("\n===========================================")
	fmt.Println("Done!")
	fmt.Println("===========================================")
}
