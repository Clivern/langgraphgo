package goskills

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/smallnest/goskills"
	goskillstool "github.com/smallnest/goskills/tool"
	langgraphtool "github.com/smallnest/langgraphgo/tool"
	"github.com/tmc/langchaingo/tools"
)

// ToolConfig 定义工具的配置，可以从 SKILL.md 或单独的配置文件中读取
// 注意：如果 SKILL.md 中已经定义了 tools 字段，此配置将被用作补充覆盖
type ToolConfig struct {
	// NameMapping 定义工具名称映射，从脚本名称到更友好的名称
	NameMapping map[string]string `json:"nameMapping"`

	// SchemaOverrides 定义工具 schema 的覆盖
	SchemaOverrides map[string]map[string]any `json:"schemaOverrides"`

	// DescriptionOverrides 定义工具描述的覆盖
	DescriptionOverrides map[string]string `json:"descriptionOverrides"`
}

// buildToolConfigFromSkill 从 SKILL.md 中的工具定义自动构建 ToolConfig
func buildToolConfigFromSkill(skill *goskills.SkillPackage) *ToolConfig {
	if len(skill.Meta.Tools) == 0 {
		return nil
	}

	config := &ToolConfig{
		NameMapping:          make(map[string]string),
		DescriptionOverrides: make(map[string]string),
		SchemaOverrides:      make(map[string]map[string]any),
	}

	for _, toolDef := range skill.Meta.Tools {
		// 构建名称映射：从工具名到工具名（保持一致）
		config.NameMapping[toolDef.Name] = toolDef.Name

		// 设置描述
		if toolDef.Description != "" {
			config.DescriptionOverrides[toolDef.Name] = toolDef.Description
		}

		// 构建 schema
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
		} else {
			// 默认参数: args 数组
			schema["properties"] = map[string]any{
				"args": map[string]any{
					"type":        "array",
					"description": "Arguments to pass to the script.",
					"items": map[string]any{
						"type": "string",
					},
				},
			}
		}

		schema["additionalProperties"] = false
		config.SchemaOverrides[toolDef.Name] = schema
	}

	return config
}

// SkillTool implements tools.Tool for goskills.
type SkillTool struct {
	name        string
	description string
	scriptMap   map[string]string
	skillPath   string
	config      *ToolConfig    // 工具配置
	schema      map[string]any // 工具的 JSON schema
}

var _ tools.Tool = &SkillTool{}
var _ interface{ Schema() map[string]any } = &SkillTool{}

func (t *SkillTool) Name() string {
	// 如果配置中有名称映射，使用映射后的名称
	if t.config != nil && t.config.NameMapping != nil {
		if newName, ok := t.config.NameMapping[t.name]; ok {
			return newName
		}
	}
	return t.name
}

func (t *SkillTool) Description() string {
	// 使用映射后的名称获取覆盖的描述
	if t.config != nil && t.config.DescriptionOverrides != nil {
		mappedName := t.Name()
		if desc, ok := t.config.DescriptionOverrides[mappedName]; ok {
			return desc
		}
	}
	return t.description
}

func (t *SkillTool) Schema() map[string]any {
	// 如果配置中有 schema 覆盖，使用覆盖的 schema
	if t.config != nil && t.config.SchemaOverrides != nil {
		mappedName := t.Name()
		if schema, ok := t.config.SchemaOverrides[mappedName]; ok {
			return schema
		}
	}
	return t.schema
}

func (t *SkillTool) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"name":        t.name,
		"description": t.description,
		"skillPath":   t.skillPath,
		"scriptMap":   t.scriptMap,
		"mappedName":  t.Name(), // 映射后的名称
	})
}

func (t *SkillTool) Call(ctx context.Context, input string) (string, error) {
	// 使用原始名称进行路由，因为脚本路径是用原始名称存储的
	originalName := t.name

	switch originalName {
	case "run_shell_code":
		var params struct {
			Code string         `json:"code"`
			Args map[string]any `json:"args"`
		}
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal run_shell_code arguments: %w", err)
		}
		shellTool := goskillstool.ShellTool{}
		return shellTool.Run(params.Args, params.Code)

	case "run_shell_script":
		var params struct {
			ScriptPath string   `json:"scriptPath"`
			Args       []string `json:"args"`
		}
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal run_shell_script arguments: %w", err)
		}
		return langgraphtool.RunShellScript(params.ScriptPath, params.Args)

	case "run_python_code":
		var params struct {
			Code string         `json:"code"`
			Args map[string]any `json:"args"`
		}
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal run_python_code arguments: %w", err)
		}
		pythonTool := goskillstool.PythonTool{}
		return pythonTool.Run(params.Args, params.Code)

	case "run_python_script":
		var params struct {
			ScriptPath string   `json:"scriptPath"`
			Args       []string `json:"args"`
		}
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal run_python_script arguments: %w", err)
		}
		return goskillstool.RunPythonScript(params.ScriptPath, params.Args)

	case "read_file":
		var params struct {
			FilePath string `json:"filePath"`
		}
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal read_file arguments: %w", err)
		}
		path := params.FilePath
		if !filepath.IsAbs(path) && t.skillPath != "" {
			resolvedPath := filepath.Join(t.skillPath, path)
			if _, err := os.Stat(resolvedPath); err == nil {
				path = resolvedPath
			}
		}
		return goskillstool.ReadFile(path)

	case "write_file":
		var params struct {
			FilePath string `json:"filePath"`
			Content  string `json:"content"`
		}
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal write_file arguments: %w", err)
		}
		err := goskillstool.WriteFile(params.FilePath, params.Content)
		if err == nil {
			return fmt.Sprintf("Successfully wrote to file: %s", params.FilePath), nil
		}
		return "", err

	case "wikipedia_search":
		var params struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal wikipedia_search arguments: %w", err)
		}
		return goskillstool.WikipediaSearch(params.Query)

	case "tavily_search":
		var params struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal tavily_search arguments: %w", err)
		}
		return goskillstool.TavilySearch(params.Query)

	case "web_fetch":
		var params struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal([]byte(input), &params); err != nil {
			return "", fmt.Errorf("failed to unmarshal web_fetch arguments: %w", err)
		}
		return goskillstool.WebFetch(params.URL)

	default:
		if scriptPath, ok := t.scriptMap[originalName]; ok {
			// 尝试解析为命名参数格式（来自 SKILL.md 工具定义）
			var namedParams map[string]any
			err := json.Unmarshal([]byte(input), &namedParams)

			if err == nil && len(namedParams) > 0 {
				// 成功解析为命名参数，转换为命令行参数
				var args []string

				// 参数映射：将 SKILL.md 中的参数名转换为脚本参数名
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

				// 按照已知的顺序处理参数（或者按字母顺序保持一致性）
				paramOrder := []string{"topic", "style", "pages", "aspect", "path", "prompt", "ar", "quality", "directory"}

				for _, key := range paramOrder {
					if value, ok := namedParams[key]; ok && value != nil {
						if flag, ok := paramMapping[key]; ok {
							args = append(args, flag)
							args = append(args, fmt.Sprintf("%v", value))
						}
					}
				}

				if strings.HasSuffix(scriptPath, ".py") {
					return goskillstool.RunPythonScript(scriptPath, args)
				} else if strings.HasSuffix(scriptPath, ".ts") || strings.HasSuffix(scriptPath, ".js") {
					return langgraphtool.RunTypeScriptScript(scriptPath, args)
				} else {
					return langgraphtool.RunShellScript(scriptPath, args)
				}
			}

			// 回退到旧的 args 数组格式
			var params struct {
				Args []string `json:"args"`
			}
			if input != "" {
				if err := json.Unmarshal([]byte(input), &params); err != nil {
					return "", fmt.Errorf("failed to unmarshal script arguments: %w", err)
				}
			}
			if strings.HasSuffix(scriptPath, ".py") {
				return goskillstool.RunPythonScript(scriptPath, params.Args)
			} else if strings.HasSuffix(scriptPath, ".ts") || strings.HasSuffix(scriptPath, ".js") {
				return langgraphtool.RunTypeScriptScript(scriptPath, params.Args)
			} else {
				return langgraphtool.RunShellScript(scriptPath, params.Args)
			}
		}
		return "", fmt.Errorf("unknown tool: %s", originalName)
	}
}

// SkillsToToolsOptions 定义转换选项
type SkillsToToolsOptions struct {
	// ToolConfig 提供工具配置（名称映射、schema 覆盖等）
	// 注意：如果 SKILL.md 中已经定义了 tools，会自动生成配置，
	//      此配置仅用于覆盖或补充 SKILL.md 中的定义
	ToolConfig *ToolConfig
}

// SkillsToTools converts a goskills.SkillPackage to a slice of tools.Tool.
// 自动从 SKILL.md 读取工具定义，如果没有定义则自动生成。
// 支持通过 ToolConfig 覆盖或补充 SKILL.md 中的定义。
func SkillsToTools(skill *goskills.SkillPackage, opts ...SkillsToToolsOptions) ([]tools.Tool, error) {
	var config *ToolConfig

	// 1. 首先尝试从 SKILL.md 自动构建配置
	skillConfig := buildToolConfigFromSkill(skill)
	if skillConfig != nil {
		config = skillConfig
	}

	// 2. 如果用户提供了配置，合并覆盖
	if len(opts) > 0 && opts[0].ToolConfig != nil {
		if config == nil {
			config = &ToolConfig{
				NameMapping:          make(map[string]string),
				DescriptionOverrides: make(map[string]string),
				SchemaOverrides:      make(map[string]map[string]any),
			}
		}
		// 合并 NameMapping
		for k, v := range opts[0].ToolConfig.NameMapping {
			config.NameMapping[k] = v
		}
		// 合并 DescriptionOverrides
		for k, v := range opts[0].ToolConfig.DescriptionOverrides {
			config.DescriptionOverrides[k] = v
		}
		// 合并 SchemaOverrides
		for k, v := range opts[0].ToolConfig.SchemaOverrides {
			config.SchemaOverrides[k] = v
		}
	}

	availableTools, scriptMap := goskills.GenerateToolDefinitions(skill)
	var result []tools.Tool

	for _, t := range availableTools {
		if t.Function.Name == "" {
			continue
		}

		// 创建描述，如果可用的话包含参数 schema
		desc := t.Function.Description

		// 从函数参数构建默认 schema
		var schema map[string]any
		if t.Function.Parameters != nil {
			// 尝试类型断言为 map[string]any
			if params, ok := t.Function.Parameters.(map[string]any); ok {
				// Parameters 是一个完整的 schema 对象，包含 type, properties, required 等
				// 直接使用它作为 schema，但确保包含 additionalProperties
				schema = params
				if _, exists := schema["additionalProperties"]; !exists {
					schema["additionalProperties"] = false
				}
			}
		}

		result = append(result, &SkillTool{
			name:        t.Function.Name,
			description: desc,
			scriptMap:   scriptMap,
			skillPath:   skill.Path,
			config:      config,
			schema:      schema,
		})
	}

	return result, nil
}

// MCPToTools converts MCP tools to langchaingo tools.
// Note: goskills also supports MCP. We can add a helper for that too if needed,
// but the user specifically asked for "Skills封装".
