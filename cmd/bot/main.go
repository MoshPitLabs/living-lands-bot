package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"living-lands-bot/internal/api"
	"living-lands-bot/internal/bot"
	"living-lands-bot/internal/config"
	"living-lands-bot/internal/database"
	"living-lands-bot/internal/services"
	"living-lands-bot/internal/utils"
	"living-lands-bot/pkg/ollama"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	logger := utils.NewLogger(cfg.Bot.LogLevel)

	// Handle one-off CLI commands
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			handleMigrate(cfg, logger)
			return
		case "index-docs":
			handleIndexDocs(cfg, logger)
			return
		case "help":
			printHelp()
			return
		}
	}

	// Start normal bot mode
	startBot(cfg, logger)
}

func handleMigrate(cfg *config.Config, logger *slog.Logger) {
	db, err := database.Open(cfg)
	if err != nil {
		logger.Error("db open failed", "error", err)
		os.Exit(1)
	}
	// Ensure database connection is closed on exit
	defer func() {
		if sqlDB, err := db.Gorm.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Error("db close failed", "error", err)
			}
		}
	}()

	if err := database.RunMigrations(db, "migrations"); err != nil {
		logger.Error("migrations failed", "error", err)
		os.Exit(1)
	}

	logger.Info("migrations complete")
}

func handleIndexDocs(cfg *config.Config, logger *slog.Logger) {
	// Parse flags for index-docs command
	fs := flag.NewFlagSet("index-docs", flag.ExitOnError)
	pathFlag := fs.String("path", "", "Path to directory or file to index")

	// Skip first two args (program name and command name)
	if err := fs.Parse(os.Args[2:]); err != nil {
		logger.Error("flag parse failed", "error", err)
		os.Exit(1)
	}

	if *pathFlag == "" {
		logger.Error("--path flag is required")
		os.Exit(1)
	}

	// Initialize database
	db, err := database.Open(cfg)
	if err != nil {
		logger.Error("db open failed", "error", err)
		os.Exit(1)
	}
	// Ensure database connection is closed on exit
	defer func() {
		if sqlDB, err := db.Gorm.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Error("db close failed", "error", err)
			}
		}
	}()

	// Initialize Ollama client
	ollamaClient := ollama.NewClient(cfg.Ollama.URL)

	// Initialize RAG service
	ragService, err := services.NewRAGService(cfg.Chroma.URL, ollamaClient, cfg.Ollama.EmbeddingModel, logger)
	if err != nil {
		logger.Error("rag service init failed", "error", err)
		os.Exit(1)
	}

	// Initialize indexer
	indexer := services.NewDocumentIndexer(ragService, logger)

	// Index the documents
	indexCtx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	logger.Info("starting document indexing", "path", *pathFlag)
	if err := indexer.IndexDirectory(indexCtx, *pathFlag); err != nil {
		logger.Error("document indexing failed", "error", err)
		os.Exit(1)
	}

	// Get stats
	stats, err := indexer.GetIndexingStats(indexCtx)
	if err != nil {
		logger.Error("failed to get stats", "error", err)
	} else {
		logger.Info("indexing complete", "stats", stats)
	}
}

func startBot(cfg *config.Config, logger *slog.Logger) {
	// Open database
	db, err := database.Open(cfg)
	if err != nil {
		logger.Error("db open failed", "error", err)
		os.Exit(1)
	}
	// Ensure database is closed on exit (including error cases)
	defer func() {
		if sqlDB, err := db.Gorm.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				logger.Error("db close failed", "error", err)
			}
		}
	}()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: cfg.Redis.Addr,
	})
	// Ensure Redis is closed on exit (including error cases)
	defer func() {
		if err := redisClient.Close(); err != nil {
			logger.Error("redis close failed", "error", err)
		}
	}()

	// Test Redis connection
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(pingCtx).Err(); err != nil {
		logger.Error("redis connection failed", "error", err)
		os.Exit(1)
	}
	logger.Info("redis client initialized", "url", cfg.Redis.URL)

	// Initialize services
	accountService := services.NewAccountService(db.Gorm, cfg.Hytale.VerifyCodeExpiry, logger)
	welcomeService := services.NewWelcomeService(db.Gorm, logger)
	channelService := services.NewChannelService(db.Gorm, logger)
	rateLimiter := services.NewRateLimiter(redisClient, cfg.Bot.RateLimitPerMin, logger)

	// Initialize Ollama client with custom timeout
	ollamaTimeout := time.Duration(cfg.Ollama.RequestTimeout) * time.Second
	ollamaClient := ollama.NewClientWithTimeout(cfg.Ollama.URL, ollamaTimeout)
	logger.Info("ollama client initialized",
		"url", cfg.Ollama.URL,
		"timeout_seconds", cfg.Ollama.RequestTimeout,
	)

	// Initialize RAG service
	ragService, err := services.NewRAGService(cfg.Chroma.URL, ollamaClient, cfg.Ollama.EmbeddingModel, logger)
	if err != nil {
		logger.Error("rag service init failed", "error", err)
		os.Exit(1)
	}

	// Build LLM config from environment
	llmConfig := services.LLMConfig{
		FastMaxTokens:       cfg.LLM.FastMaxTokens,
		FastTemperature:     cfg.LLM.FastTemperature,
		FastTopK:            20,
		FastTopP:            0.85,
		StandardMaxTokens:   cfg.LLM.StandardMaxTokens,
		StandardTemperature: cfg.LLM.StandardTemperature,
		StandardTopK:        30,
		StandardTopP:        0.9,
		DeepMaxTokens:       cfg.LLM.DeepMaxTokens,
		DeepTemperature:     cfg.LLM.DeepTemperature,
		DeepTopK:            40,
		DeepTopP:            0.95,
		RepeatPenalty:       1.1,
		NumContext:          2048,
	}

	// Initialize LLM service with config
	llmService, err := services.NewLLMServiceWithConfig(ollamaClient, cfg.Ollama.Model, cfg.Bot.PersonalityFile, llmConfig, logger)
	if err != nil {
		logger.Error("llm service init failed", "error", err)
		os.Exit(1)
	}

	// Initialize bot and HTTP server
	dBot, err := bot.New(cfg, accountService, ragService, llmService, welcomeService, channelService, rateLimiter, logger)
	if err != nil {
		logger.Error("discord bot init failed", "error", err)
		os.Exit(1)
	}

	httpServer := api.NewServer(cfg, accountService, logger)

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Start HTTP server and keep it running even if Discord auth fails.
	errCh := make(chan error, 1)
	go func() { errCh <- httpServer.Start() }()

	// Start Discord with retry loop. Useful during initial setup when the token
	// may be missing/invalid, or Discord is temporarily unavailable.
	go func() {
		backoff := 5 * time.Second
		maxBackoff := 2 * time.Minute
		for {
			select {
			case <-rootCtx.Done():
				return
			default:
			}

			if err := dBot.Start(); err != nil {
				logger.Error("discord start failed", "error", err)
				select {
				case <-time.After(backoff):
				case <-rootCtx.Done():
					return
				}

				backoff *= 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}
				continue
			}

			// Reset backoff after a successful connect and wait for shutdown.
			backoff = 5 * time.Second
			<-rootCtx.Done()
			return
		}
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown requested")
	case err := <-errCh:
		logger.Error("http server exited", "error", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.ShutdownWithContext(ctx)
	_ = dBot.Stop()
	_ = redisClient.Close()
}

func printHelp() {
	help := `Living Lands Discord Bot

Usage:
  ./bot [command] [options]

Commands:
  migrate              Run database migrations
  index-docs           Index documents for RAG
    --path <path>      Path to directory or file to index (required)
  help                 Show this help message
  (no command)         Start the bot in normal mode

Examples:
  ./bot migrate
  ./bot index-docs --path ./docs
  ./bot
`
	println(help)
}
