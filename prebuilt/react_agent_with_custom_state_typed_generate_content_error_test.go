package prebuilt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

func TestCreateReactAgentWithCustomStateTyped_GenerateContentError(t *testing.T) {
	type CustomState struct {
		Messages       []llms.MessageContent
		IterationCount int
	}

	mockLLM := &MockLLMError{}

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
	hasToolCalls := func(msgs []llms.MessageContent) bool { return false }

	agent, err := CreateReactAgentWithCustomStateTyped(
		mockLLM,
		[]tools.Tool{},
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
	assert.Contains(t, err.Error(), "mock LLM GenerateContent error")
}
