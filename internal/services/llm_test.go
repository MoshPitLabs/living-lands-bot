package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetermineMode(t *testing.T) {
	tests := []struct {
		name          string
		intent        QueryIntent
		hasRAGContext bool
		expectedMode  ResponseMode
	}{
		// Conversational intents always use fast mode
		{
			name:          "greeting uses fast mode",
			intent:        IntentConversational,
			hasRAGContext: false,
			expectedMode:  ModeFast,
		},
		{
			name:          "greeting with RAG still uses fast mode",
			intent:        IntentConversational,
			hasRAGContext: true,
			expectedMode:  ModeFast,
		},

		// Navigation and account help use fast mode
		{
			name:          "navigation uses fast mode",
			intent:        IntentNavigation,
			hasRAGContext: false,
			expectedMode:  ModeFast,
		},
		{
			name:          "account help uses fast mode",
			intent:        IntentAccountHelp,
			hasRAGContext: false,
			expectedMode:  ModeFast,
		},

		// Knowledge intent depends on RAG context
		{
			name:          "knowledge without RAG uses standard mode",
			intent:        IntentKnowledge,
			hasRAGContext: false,
			expectedMode:  ModeStandard,
		},
		{
			name:          "knowledge with RAG uses deep mode",
			intent:        IntentKnowledge,
			hasRAGContext: true,
			expectedMode:  ModeDeep,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mode := DetermineMode(tt.intent, tt.hasRAGContext)
			assert.Equal(t, tt.expectedMode, mode)
		})
	}
}

func TestResponseModeString(t *testing.T) {
	tests := []struct {
		mode     ResponseMode
		expected string
	}{
		{ModeFast, "fast"},
		{ModeStandard, "standard"},
		{ModeDeep, "deep"},
		{ResponseMode(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.mode.String())
		})
	}
}

func TestDefaultLLMConfig(t *testing.T) {
	cfg := DefaultLLMConfig()

	// Fast and Standard have same max tokens (both 120), Deep is higher
	assert.LessOrEqual(t, cfg.FastMaxTokens, cfg.StandardMaxTokens)
	assert.Less(t, cfg.StandardMaxTokens, cfg.DeepMaxTokens)

	// Temperature should increase with complexity
	assert.Less(t, cfg.FastTemperature, cfg.StandardTemperature)
	assert.Less(t, cfg.StandardTemperature, cfg.DeepTemperature)

	// All values should be reasonable
	assert.Greater(t, cfg.FastMaxTokens, 0)
	assert.Greater(t, cfg.DeepMaxTokens, 0)
	assert.Greater(t, cfg.NumContext, 0)
	assert.Greater(t, cfg.RepeatPenalty, 1.0)
}

func TestLLMConfigValues(t *testing.T) {
	cfg := DefaultLLMConfig()

	// Validate token limits are reasonable for Discord
	assert.LessOrEqual(t, cfg.FastMaxTokens, 150, "fast tokens should be under 150")
	assert.LessOrEqual(t, cfg.StandardMaxTokens, 150, "standard tokens should be under 150")
	assert.LessOrEqual(t, cfg.DeepMaxTokens, 300, "deep tokens should be under 300")

	// Validate temperature ranges
	assert.GreaterOrEqual(t, cfg.FastTemperature, 0.0)
	assert.LessOrEqual(t, cfg.FastTemperature, 1.0)
	assert.GreaterOrEqual(t, cfg.DeepTemperature, 0.0)
	assert.LessOrEqual(t, cfg.DeepTemperature, 1.0)

	// Validate TopP is valid probability
	assert.GreaterOrEqual(t, cfg.FastTopP, 0.0)
	assert.LessOrEqual(t, cfg.FastTopP, 1.0)
	assert.GreaterOrEqual(t, cfg.DeepTopP, 0.0)
	assert.LessOrEqual(t, cfg.DeepTopP, 1.0)
}
