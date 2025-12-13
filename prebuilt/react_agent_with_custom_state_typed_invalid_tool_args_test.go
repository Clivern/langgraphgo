package prebuilt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

// MockLLMInvalidToolArgs for testing tool call with invalid arguments JSON
type MockLLMInvalidToolArgs struct{}

func (m *MockLLMInvalidToolArgs) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	return &llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{
				Content: "Calling tool with invalid args",
				ToolCalls: []llms.ToolCall{
					{
						ID: "call_1",
						FunctionCall: &llms.FunctionCall{
							Name:      "test_tool",
							Arguments: `{"input":}`, // Invalid JSON
						},
					},
				},
			},
		},
	}, nil
}

func (m *MockLLMInvalidToolArgs) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	return "", nil
}

func TestCreateReactAgentWithCustomStateTyped_InvalidToolArgs(t *testing.T) {
	type CustomState struct {
		Messages       []llms.MessageContent
		IterationCount int
	}

	tool := &MockToolForReact{name: "test_tool", description: "Test tool"}
	mockLLM := &MockLLMInvalidToolArgs{}

	getMessages := func(s CustomState) []llms.MessageContent { return s.Messages }
	setMessages := func(s CustomState, msgs []llms.MessageContent) CustomState {
		s.Messages = append(s.Messages, msgs...)
		return s
	}
	getIterationCount := func(s CustomState) int { return s.IterationCount }
	setIterationCount := func(s CustomState, count int) CustomState {
		s.IterationCount = count
		return s
	}
	hasToolCalls := func(msgs []llms.MessageContent) bool {
		if len(msgs) == 0 {
			return false
		}
		lastMsg := msgs[len(msgs)-1]
		for _, part := range lastMsg.Parts {
			if _, ok := part.(llms.ToolCall); ok {
				return true
			}
		}
		return false
	}

	agent, err := CreateReactAgentWithCustomStateTyped(
		mockLLM,
		[]tools.Tool{tool},
		getMessages,
		setMessages,
		getIterationCount,
		setIterationCount,
		hasToolCalls,
		3,
	)
	require.NoError(t, err)
	require.NotNil(t, agent)

	initialState := CustomState{
		Messages: []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "Test"),
		},
	}

	_, err = agent.Invoke(context.Background(), initialState)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tool arguments")
}
