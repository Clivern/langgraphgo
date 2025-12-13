package prebuilt

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/tools"
)

func TestCreateReactAgentTyped_GenerateContentError(t *testing.T) {
	mockLLM := &MockLLMError{}

	agent, err := CreateReactAgentTyped(mockLLM, []tools.Tool{}, 3)
	require.NoError(t, err)
	require.NotNil(t, agent)

	initialState := ReactAgentState{
		Messages: []llms.MessageContent{
			llms.TextParts(llms.ChatMessageTypeHuman, "Test"),
		},
	}

	_, err = agent.Invoke(context.Background(), initialState)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mock LLM GenerateContent error")
}
