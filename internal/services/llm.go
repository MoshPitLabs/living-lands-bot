package services

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"living-lands-bot/pkg/language"
	"living-lands-bot/pkg/ollama"
)

// ResponseMode determines the complexity of the LLM response.
type ResponseMode int

const (
	// ModeFast - Conversational responses with minimal system prompt
	ModeFast ResponseMode = iota
	// ModeStandard - Normal responses with condensed system prompt
	ModeStandard
	// ModeDeep - Technical responses with full system prompt and RAG context
	ModeDeep
)

func (m ResponseMode) String() string {
	switch m {
	case ModeFast:
		return "fast"
	case ModeStandard:
		return "standard"
	case ModeDeep:
		return "deep"
	default:
		return "unknown"
	}
}

// Personality represents the LLM's character and behavior.
type Personality struct {
	Name               string `yaml:"name"`
	Role               string `yaml:"role"`
	Tone               string `yaml:"tone"`
	Knowledge          string `yaml:"knowledge"`
	SystemPrompt       string `yaml:"system_prompt"`
	FastModePrompt     string `yaml:"fast_mode_prompt"`
	StandardModePrompt string `yaml:"standard_mode_prompt"`
	DeepModePrompt     string `yaml:"deep_mode_prompt"`
}

// LLMConfig holds tunable parameters for LLM generation.
type LLMConfig struct {
	// Fast mode - for conversational queries (greetings, thanks, etc.)
	FastMaxTokens   int
	FastTemperature float64
	FastTopK        int
	FastTopP        float64

	// Standard mode - for simple questions
	StandardMaxTokens   int
	StandardTemperature float64
	StandardTopK        int
	StandardTopP        float64

	// Deep mode - for technical questions with RAG
	DeepMaxTokens   int
	DeepTemperature float64
	DeepTopK        int
	DeepTopP        float64

	// Common settings
	RepeatPenalty float64
	NumContext    int
}

// DefaultLLMConfig returns optimized default settings.
func DefaultLLMConfig() LLMConfig {
	return LLMConfig{
		// Fast mode: very quick, short responses
		FastMaxTokens:   120,
		FastTemperature: 0.5,
		FastTopK:        20,
		FastTopP:        0.85,

		// Standard mode: balanced speed and quality
		StandardMaxTokens:   120,
		StandardTemperature: 0.6,
		StandardTopK:        30,
		StandardTopP:        0.9,

		// Deep mode: thorough technical responses
		DeepMaxTokens:   180,
		DeepTemperature: 0.7,
		DeepTopK:        40,
		DeepTopP:        0.95,

		// Common: encourage diverse responses
		RepeatPenalty: 1.1,
		NumContext:    2048, // Reduced context window for speed
	}
}

// LLMService handles LLM generation with RAG context.
type LLMService struct {
	client      *ollama.Client
	model       string
	personality Personality
	config      LLMConfig
	logger      *slog.Logger

	// Condensed system prompts for different modes
	fastSystemPrompt     string
	standardSystemPrompt string
	deepSystemPrompt     string
}

// LLMMetrics holds timing and token information for observability.
type LLMMetrics struct {
	Mode            ResponseMode
	TotalDuration   time.Duration
	PromptTokens    int
	GeneratedTokens int
	TokensPerSecond float64
	PromptEvalTime  time.Duration
	GenerationTime  time.Duration
}

// NewLLMService initializes an LLM service with personality configuration.
func NewLLMService(ollamaClient *ollama.Client, model string, personalityFile string, logger *slog.Logger) (*LLMService, error) {
	return NewLLMServiceWithConfig(ollamaClient, model, personalityFile, DefaultLLMConfig(), logger)
}

// NewLLMServiceWithConfig initializes an LLM service with custom configuration.
func NewLLMServiceWithConfig(ollamaClient *ollama.Client, model string, personalityFile string, config LLMConfig, logger *slog.Logger) (*LLMService, error) {
	// Load personality from YAML file
	personality, err := loadPersonality(personalityFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load personality: %w", err)
	}

	s := &LLMService{
		client:      ollamaClient,
		model:       model,
		personality: personality,
		config:      config,
		logger:      logger,
	}

	// Build condensed system prompts for different modes
	s.buildCondensedPrompts()

	logger.Info("llm service initialized",
		"model", model,
		"personality", personality.Name,
		"role", personality.Role,
		"fast_tokens", config.FastMaxTokens,
		"standard_tokens", config.StandardMaxTokens,
		"deep_tokens", config.DeepMaxTokens,
	)

	return s, nil
}

// buildCondensedPrompts creates optimized system prompts for each mode.
func (s *LLMService) buildCondensedPrompts() {
	// Fast mode: Minimal prompt for greetings and simple responses
	s.fastSystemPrompt = s.personality.FastModePrompt

	// Standard mode: Condensed personality for general questions
	s.standardSystemPrompt = s.personality.StandardModePrompt

	// Deep mode: Use full system prompt from personality.yaml
	s.deepSystemPrompt = s.personality.DeepModePrompt
}

// loadPersonality reads and parses the personality YAML file.
func loadPersonality(filePath string) (Personality, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Personality{}, fmt.Errorf("failed to read personality file: %w", err)
	}

	var p Personality
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Personality{}, fmt.Errorf("failed to parse personality yaml: %w", err)
	}

	// Validate required fields
	if p.SystemPrompt == "" {
		return Personality{}, fmt.Errorf("personality missing system_prompt")
	}

	return p, nil
}

// DetermineMode selects the appropriate response mode based on intent.
func DetermineMode(intent QueryIntent, hasRAGContext bool) ResponseMode {
	switch intent {
	case IntentConversational:
		return ModeFast
	case IntentNavigation, IntentAccountHelp:
		return ModeFast
	case IntentIdentity:
		return ModeStandard
	case IntentKnowledge:
		if hasRAGContext {
			return ModeDeep
		}
		return ModeStandard
	default:
		return ModeStandard
	}
}

// GenerateResponse generates an LLM response with RAG context.
func (s *LLMService) GenerateResponse(ctx context.Context, userMessage string, ragContext []string) (string, error) {
	return s.GenerateResponseWithIntent(ctx, userMessage, ragContext, IntentKnowledge)
}

// GenerateResponseWithIntent generates a response using the appropriate mode for the intent.
func (s *LLMService) GenerateResponseWithIntent(ctx context.Context, userMessage string, ragContext []string, intent QueryIntent) (string, error) {
	startTime := time.Now()

	// Sanitize user input to prevent prompt injection
	userMessage = SanitizePromptInput(userMessage)

	// Determine the response mode based on intent
	mode := DetermineMode(intent, len(ragContext) > 0)

	// Detect the language of the user's message
	detectedLang, confidence := language.Detect(userMessage)

	// Build prompt with RAG context (if any)
	prompt := s.buildPrompt(userMessage, ragContext, mode)

	// Get system prompt for this mode
	systemPrompt := s.getSystemPrompt(mode, detectedLang)

	// Get generation options for this mode
	options := s.getOptions(mode)

	// Call Ollama
	req := ollama.GenerateRequest{
		Model:   s.model,
		Prompt:  prompt,
		System:  systemPrompt,
		Stream:  false,
		Options: options,
	}

	resp, err := s.client.Generate(ctx, req)
	if err != nil {
		s.logger.Error("llm generation failed",
			"error", err,
			"mode", mode.String(),
			"intent", intent.String(),
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
		return "", fmt.Errorf("llm generation failed: %w", err)
	}

	// Clean up response
	answer := strings.TrimSpace(resp.Response)

	// Remove trailing prompt template artifacts
	// Look for double newline followed by "User:" or "User:" at the end
	patterns := []string{"\n\nUser:", "\nUser:", "\nUser :", "\n\nAssistant:", "\nAssistant:"}
	for _, pattern := range patterns {
		if idx := strings.Index(answer, pattern); idx != -1 {
			answer = strings.TrimSpace(answer[:idx])
		}
	}

	// Final trim
	answer = strings.TrimSpace(answer)

	// Calculate and log metrics
	metrics := s.calculateMetrics(resp, mode, startTime)
	s.logMetrics(userMessage, detectedLang, confidence, ragContext, answer, metrics)

	return answer, nil
}

// getSystemPrompt returns the appropriate system prompt for the mode.
func (s *LLMService) getSystemPrompt(mode ResponseMode, lang language.Language) string {
	var systemPrompt string

	switch mode {
	case ModeFast:
		systemPrompt = s.fastSystemPrompt
	case ModeStandard:
		systemPrompt = s.standardSystemPrompt
	case ModeDeep:
		systemPrompt = s.deepSystemPrompt
	default:
		systemPrompt = s.standardSystemPrompt
	}

	// Add language instruction if non-English
	if lang.IsNonEnglish() {
		systemPrompt = fmt.Sprintf("%s\n\nIMPORTANT: Respond in %s.", systemPrompt, lang.String())
	}

	return systemPrompt
}

// getOptions returns generation options tuned for the response mode.
func (s *LLMService) getOptions(mode ResponseMode) ollama.Options {
	switch mode {
	case ModeFast:
		return ollama.Options{
			Temperature:   s.config.FastTemperature,
			NumPredict:    s.config.FastMaxTokens,
			TopK:          s.config.FastTopK,
			TopP:          s.config.FastTopP,
			RepeatPenalty: s.config.RepeatPenalty,
			NumCtx:        s.config.NumContext,
		}
	case ModeStandard:
		return ollama.Options{
			Temperature:   s.config.StandardTemperature,
			NumPredict:    s.config.StandardMaxTokens,
			TopK:          s.config.StandardTopK,
			TopP:          s.config.StandardTopP,
			RepeatPenalty: s.config.RepeatPenalty,
			NumCtx:        s.config.NumContext,
		}
	case ModeDeep:
		return ollama.Options{
			Temperature:   s.config.DeepTemperature,
			NumPredict:    s.config.DeepMaxTokens,
			TopK:          s.config.DeepTopK,
			TopP:          s.config.DeepTopP,
			RepeatPenalty: s.config.RepeatPenalty,
			NumCtx:        s.config.NumContext,
		}
	default:
		return ollama.Options{
			Temperature:   s.config.StandardTemperature,
			NumPredict:    s.config.StandardMaxTokens,
			TopK:          s.config.StandardTopK,
			TopP:          s.config.StandardTopP,
			RepeatPenalty: s.config.RepeatPenalty,
			NumCtx:        s.config.NumContext,
		}
	}
}

// buildPrompt constructs the final prompt with RAG context.
func (s *LLMService) buildPrompt(userMessage string, ragContext []string, mode ResponseMode) string {
	var prompt strings.Builder

	// Only add RAG context for Deep mode with actual context
	if mode == ModeDeep && len(ragContext) > 0 {
		prompt.WriteString("Relevant documentation (use only if it answers the question):\n")
		for i, ctx := range ragContext {
			// Truncate very long contexts to save tokens
			truncated := ctx
			if len(truncated) > 500 {
				truncated = truncated[:500] + "..."
			}
			prompt.WriteString(fmt.Sprintf("%d. %s\n", i+1, truncated))
		}
		prompt.WriteString("\n---\n\n")
	}

	// Add the user's question
	prompt.WriteString(fmt.Sprintf("User: %s\nAssistant:", userMessage))

	return prompt.String()
}

// calculateMetrics extracts timing and token metrics from the response.
func (s *LLMService) calculateMetrics(resp *ollama.GenerateResponse, mode ResponseMode, startTime time.Time) LLMMetrics {
	metrics := LLMMetrics{
		Mode:            mode,
		TotalDuration:   time.Since(startTime),
		PromptTokens:    resp.PromptEvalCount,
		GeneratedTokens: resp.EvalCount,
	}

	// Calculate tokens per second if we have eval data
	if resp.EvalDuration > 0 && resp.EvalCount > 0 {
		evalSeconds := float64(resp.EvalDuration) / 1e9
		metrics.TokensPerSecond = float64(resp.EvalCount) / evalSeconds
	}

	// Convert nanoseconds to durations
	if resp.PromptEvalDuration > 0 {
		metrics.PromptEvalTime = time.Duration(resp.PromptEvalDuration)
	}
	if resp.EvalDuration > 0 {
		metrics.GenerationTime = time.Duration(resp.EvalDuration)
	}

	return metrics
}

// logMetrics logs generation metrics for observability.
func (s *LLMService) logMetrics(userMessage string, lang language.Language, confidence int, ragContext []string, answer string, metrics LLMMetrics) {
	// Log at info level for monitoring
	s.logger.Info("llm response generated",
		"mode", metrics.Mode.String(),
		"duration_ms", metrics.TotalDuration.Milliseconds(),
		"prompt_tokens", metrics.PromptTokens,
		"generated_tokens", metrics.GeneratedTokens,
		"tokens_per_sec", fmt.Sprintf("%.1f", metrics.TokensPerSecond),
		"rag_context_count", len(ragContext),
		"response_length", len(answer),
	)

	// Log detailed debug info
	s.logger.Debug("llm generation details",
		"user_message", userMessage,
		"detected_language", lang.String(),
		"language_confidence", confidence,
		"prompt_eval_ms", metrics.PromptEvalTime.Milliseconds(),
		"generation_ms", metrics.GenerationTime.Milliseconds(),
	)
}
