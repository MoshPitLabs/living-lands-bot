package config

import (
	"fmt"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Discord struct {
		Token   string `envconfig:"DISCORD_TOKEN" required:"true"`
		GuildID string `envconfig:"DISCORD_GUILD_ID" required:"true"`
	}

	HTTP struct {
		Addr string `envconfig:"HTTP_ADDR" default:":8000"`
	}

	Database struct {
		Host     string `envconfig:"DB_HOST" default:"localhost"`
		Port     int    `envconfig:"DB_PORT" default:"5432"`
		User     string `envconfig:"DB_USER" default:"bot"`
		Password string `envconfig:"DB_PASSWORD" required:"true"`
		Name     string `envconfig:"DB_NAME" default:"livinglands"`
		SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
	}

	Redis struct {
		URL  string `envconfig:"REDIS_URL" default:"redis://localhost:6379"`
		Addr string // Parsed host:port for go-redis client
	}

	Chroma struct {
		URL string `envconfig:"CHROMA_URL" default:"http://localhost:8000"`
	}

	Ollama struct {
		URL            string `envconfig:"OLLAMA_URL" default:"http://localhost:11434"`
		Model          string `envconfig:"LLM_MODEL" default:"mistral:7b-instruct"`
		EmbeddingModel string `envconfig:"EMBEDDING_MODEL" default:"nomic-embed-text"`
		MaxContextMsgs int    `envconfig:"MAX_CONTEXT_MESSAGES" default:"10"`
		// Request timeout in seconds (should be longer than Discord's 30s window)
		RequestTimeout int `envconfig:"OLLAMA_TIMEOUT" default:"60"`
	}

	LLM struct {
		// Fast mode settings (conversational queries)
		FastMaxTokens   int     `envconfig:"LLM_FAST_MAX_TOKENS" default:"60"`
		FastTemperature float64 `envconfig:"LLM_FAST_TEMPERATURE" default:"0.5"`

		// Standard mode settings (simple questions)
		StandardMaxTokens   int     `envconfig:"LLM_STANDARD_MAX_TOKENS" default:"120"`
		StandardTemperature float64 `envconfig:"LLM_STANDARD_TEMPERATURE" default:"0.6"`

		// Deep mode settings (technical questions with RAG)
		DeepMaxTokens   int     `envconfig:"LLM_DEEP_MAX_TOKENS" default:"180"`
		DeepTemperature float64 `envconfig:"LLM_DEEP_TEMPERATURE" default:"0.7"`
	}

	Hytale struct {
		APISecret        string `envconfig:"HYTALE_API_SECRET" required:"true"`
		VerifyCodeExpiry int    `envconfig:"VERIFY_CODE_EXPIRY" default:"600"`
	}

	Bot struct {
		RateLimitPerMin int    `envconfig:"RATE_LIMIT_PER_MINUTE" default:"5"`
		LogLevel        string `envconfig:"LOG_LEVEL" default:"info"`
		PersonalityFile string `envconfig:"PERSONALITY_FILE" default:"configs/personality.yaml"`
	}
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	// discordgo expects "Bot <token>"; we store token only
	cfg.Discord.Token = strings.TrimSpace(cfg.Discord.Token)
	if cfg.Discord.Token == "" {
		return nil, fmt.Errorf("DISCORD_TOKEN is required")
	}

	// Parse Redis URL to extract host:port for go-redis client
	redisURL := cfg.Redis.URL
	if strings.HasPrefix(redisURL, "redis://") {
		// Remove the scheme
		redisURL = strings.TrimPrefix(redisURL, "redis://")
	}
	cfg.Redis.Addr = redisURL

	// Validate all configuration values
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate checks if all configuration values are valid.
// Returns a detailed error message if any validation fails.
func (c *Config) Validate() error {
	// Validate Discord config
	if c.Discord.GuildID == "" {
		return fmt.Errorf("DISCORD_GUILD_ID is required and cannot be empty")
	}

	// Validate Database config
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	if c.Database.Port < 1 || c.Database.Port > 65535 {
		return fmt.Errorf("DB_PORT must be between 1 and 65535, got %d", c.Database.Port)
	}
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	// Validate HTTP config
	if c.HTTP.Addr == "" {
		return fmt.Errorf("HTTP_ADDR is required")
	}

	// Validate Redis config
	if c.Redis.Addr == "" {
		return fmt.Errorf("REDIS_URL is required or REDIS_ADDR cannot be empty")
	}

	// Validate Chroma config
	if c.Chroma.URL == "" {
		return fmt.Errorf("CHROMA_URL is required")
	}

	// Validate Ollama config
	if c.Ollama.URL == "" {
		return fmt.Errorf("OLLAMA_URL is required")
	}
	if c.Ollama.Model == "" {
		return fmt.Errorf("LLM_MODEL is required")
	}
	if c.Ollama.EmbeddingModel == "" {
		return fmt.Errorf("EMBEDDING_MODEL is required")
	}
	if c.Ollama.RequestTimeout < 1 || c.Ollama.RequestTimeout > 600 {
		return fmt.Errorf("OLLAMA_TIMEOUT must be between 1 and 600 seconds, got %d", c.Ollama.RequestTimeout)
	}

	// Validate LLM config
	if c.LLM.FastMaxTokens < 1 || c.LLM.FastMaxTokens > 1000 {
		return fmt.Errorf("LLM_FAST_MAX_TOKENS must be between 1 and 1000, got %d", c.LLM.FastMaxTokens)
	}
	if c.LLM.StandardMaxTokens < 1 || c.LLM.StandardMaxTokens > 1000 {
		return fmt.Errorf("LLM_STANDARD_MAX_TOKENS must be between 1 and 1000, got %d", c.LLM.StandardMaxTokens)
	}
	if c.LLM.DeepMaxTokens < 1 || c.LLM.DeepMaxTokens > 1000 {
		return fmt.Errorf("LLM_DEEP_MAX_TOKENS must be between 1 and 1000, got %d", c.LLM.DeepMaxTokens)
	}
	if c.LLM.FastTemperature < 0 || c.LLM.FastTemperature > 2 {
		return fmt.Errorf("LLM_FAST_TEMPERATURE must be between 0 and 2, got %f", c.LLM.FastTemperature)
	}
	if c.LLM.StandardTemperature < 0 || c.LLM.StandardTemperature > 2 {
		return fmt.Errorf("LLM_STANDARD_TEMPERATURE must be between 0 and 2, got %f", c.LLM.StandardTemperature)
	}
	if c.LLM.DeepTemperature < 0 || c.LLM.DeepTemperature > 2 {
		return fmt.Errorf("LLM_DEEP_TEMPERATURE must be between 0 and 2, got %f", c.LLM.DeepTemperature)
	}

	// Validate Hytale config
	if c.Hytale.APISecret == "" {
		return fmt.Errorf("HYTALE_API_SECRET is required")
	}
	if c.Hytale.VerifyCodeExpiry < 60 || c.Hytale.VerifyCodeExpiry > 3600 {
		return fmt.Errorf("VERIFY_CODE_EXPIRY must be between 60 and 3600 seconds, got %d", c.Hytale.VerifyCodeExpiry)
	}

	// Validate Bot config
	if c.Bot.RateLimitPerMin < 1 || c.Bot.RateLimitPerMin > 1000 {
		return fmt.Errorf("RATE_LIMIT_PER_MINUTE must be between 1 and 1000, got %d", c.Bot.RateLimitPerMin)
	}
	if c.Bot.LogLevel == "" {
		return fmt.Errorf("LOG_LEVEL is required")
	}
	if c.Bot.PersonalityFile == "" {
		return fmt.Errorf("PERSONALITY_FILE is required")
	}

	return nil
}
