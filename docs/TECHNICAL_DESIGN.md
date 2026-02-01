# Living Lands Discord Bot - Technical Design Document (Go)

## 1. Overview

A Discord bot for the Living Lands Reloaded Hytale mod written in Go. Provides lore-friendly interactions, player account linking, mod Q&A with RAG, and channel navigation. Features a local LLM (Ollama) for in-character conversations.

## 2. Goals

- Welcome new members with Hytale lore-friendly messages
- Link Discord accounts to Hytale usernames
- Answer mod questions accurately using RAG
- Direct users to appropriate channels
- Engage users with an in-character LLM personality

## 3. Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Discord Gateway                         │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   Discord Bot (Go)                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Commands    │  │   Events     │  │   Services   │     │
│  │  - /link     │  │  - Welcome   │  │  - Account   │     │
│  │  - /ask      │  │  - Message   │  │    Linking   │     │
│  │  - /guide    │  │  - Button    │  │  - RAG       │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┼────────────┬────────────┐
        ▼            ▼            ▼            ▼
┌──────────────┐ ┌────────┐ ┌─────────┐ ┌────────────┐
│  PostgreSQL  │ │ Redis  │ │ Ollama  │ │ ChromaDB   │
│  - Users     │ │ Cache  │ │ LLM     │ │ Vectors    │
│  - Links     │ │ Memory │ │ Mistral │ │ Mod Docs   │
└──────────────┘ └────────┘ └─────────┘ └────────────┘
```

## 4. Tech Stack

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| Language | Go | 1.25.6 | Application code |
| Discord Library | discordgo | v0.29.0 | Discord API wrapper |
| Web Framework | Fiber | v2.52+ | HTTP API for Hytale |
| ORM | GORM | v1.25+ | Database abstraction |
| Migration | golang-migrate | v4.17+ | Schema versioning |
| LLM Client | Native HTTP | - | Ollama integration |
| Vector DB | chroma-go | v0.1+ | ChromaDB Go SDK |
| Cache | go-redis | v9.4+ | Redis client |
| Validation | go-playground/validator | v10+ | Struct validation |
| Config | envconfig | v1.4+ | Environment parsing |
| Logging | slog (std lib) | - | Structured logging |
| Container | Docker | 24+ | Deployment |
| Process Mgmt | Docker Compose | 2.20+ | Service orchestration |

## 5. Database Schema

```sql
-- Users table (Discord + Hytale linking)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    discord_id BIGINT UNIQUE NOT NULL,
    discord_username VARCHAR(32) NOT NULL,
    hytale_username VARCHAR(32),
    hytale_uuid UUID,
    verification_code VARCHAR(8),
    verified_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Conversation threads (LLM context)
CREATE TABLE conversations (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    channel_id BIGINT NOT NULL,
    thread_ts TIMESTAMP DEFAULT NOW(),
    messages JSONB DEFAULT '[]',
    ttl TIMESTAMP DEFAULT NOW() + INTERVAL '30 minutes'
);

-- Channel routing rules
CREATE TABLE channel_routes (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(50) UNIQUE NOT NULL,
    channel_id BIGINT NOT NULL,
    description TEXT,
    emoji VARCHAR(10)
);

-- Welcome message templates
CREATE TABLE welcome_templates (
    id SERIAL PRIMARY KEY,
    message TEXT NOT NULL,
    weight INTEGER DEFAULT 1,
    active BOOLEAN DEFAULT TRUE
);

-- Mod documentation embeddings metadata
CREATE TABLE doc_sources (
    id SERIAL PRIMARY KEY,
    source_type VARCHAR(20) NOT NULL, -- 'config', 'changelog', 'wiki'
    source_path TEXT NOT NULL,
    last_updated TIMESTAMP DEFAULT NOW(),
    checksum VARCHAR(64)
);

-- Create indexes for performance
CREATE INDEX idx_users_discord_id ON users(discord_id);
CREATE INDEX idx_users_verification_code ON users(verification_code);
CREATE INDEX idx_conversations_user_id ON conversations(user_id);
CREATE INDEX idx_conversations_ttl ON conversations(ttl);
```

## 6. Core Components

### 6.1 Project Structure

```
livinglands-bot/
├── cmd/
│   └── bot/
│       └── main.go              # Entry point
├── internal/
│   ├── bot/
│   │   ├── bot.go               # Discord bot setup
│   │   ├── commands.go          # Slash command handlers
│   │   ├── events.go            # Discord event handlers
│   │   └── interactions.go      # Button/modal handlers
│   ├── config/
│   │   └── config.go            # Environment configuration
│   ├── database/
│   │   ├── db.go                # GORM setup
│   │   ├── migrate.go           # Migration runner
│   │   └── models/
│   │       ├── user.go
│   │       ├── conversation.go
│   │       └── channel_route.go
│   ├── services/
│   │   ├── account.go           # Account linking service
│   │   ├── llm.go              # LLM/Ollama service
│   │   ├── rag.go              # RAG pipeline
│   │   ├── welcome.go          # Welcome service
│   │   └── channel.go          # Channel routing
│   ├── api/
│   │   ├── server.go            # Echo HTTP server
│   │   └── handlers/
│   │       └── verify.go        # Hytale verification endpoint
│   └── utils/
│       ├── logger.go            # slog configuration
│       └── validation.go        # Custom validators
├── pkg/
│   └── ollama/                  # Ollama client package
│       └── client.go
├── configs/
│   ├── personality.yaml         # Bot personality config
│   └── prompts/                 # System prompt templates
├── migrations/
│   └── *.sql                    # Database migrations
├── scripts/
│   └── index_docs.go           # Document indexing script
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── go.sum
```

### 6.2 Discord Bot Module

**File:** `internal/bot/bot.go`

```go
package bot

import (
    "github.com/bwmarrin/discordgo"
)

type Bot struct {
    session *discordgo.Session
    config  *config.Config
    services *Services
    logger  *slog.Logger
}

type Services struct {
    Account  *services.AccountService
    LLM      *services.LLMService
    RAG      *services.RAGService
    Welcome  *services.WelcomeService
    Channel  *services.ChannelService
}

func New(cfg *config.Config, svcs *Services, logger *slog.Logger) (*Bot, error) {
    dg, err := discordgo.New("Bot " + cfg.Discord.Token)
    if err != nil {
        return nil, err
    }
    
    dg.Identify.Intents = discordgo.IntentsGuildMembers | 
                          discordgo.IntentsGuildMessages |
                          discordgo.IntentsMessageContent
    
    bot := &Bot{
        session: dg,
        config: cfg,
        services: svcs,
        logger: logger,
    }
    
    // Register handlers
    dg.AddHandler(bot.onReady)
    dg.AddHandler(bot.onGuildMemberAdd)
    dg.AddHandler(bot.onInteractionCreate)
    
    return bot, nil
}

func (b *Bot) Start() error {
    return b.session.Open()
}

func (b *Bot) Stop() error {
    return b.session.Close()
}
```

### 6.3 Slash Commands

**File:** `internal/bot/commands.go`

```go
var (
    commands = []*discordgo.ApplicationCommand{
        {
            Name:        "link",
            Description: "Link your Hytale account",
        },
        {
            Name:        "ask",
            Description: "Ask about Living Lands",
            Options: []*discordgo.ApplicationCommandOption{
                {
                    Type:        discordgo.ApplicationCommandOptionString,
                    Name:        "question",
                    Description: "Your question",
                    Required:    true,
                },
            },
        },
        {
            Name:        "guide",
            Description: "Get directions to channels",
        },
    }
)

func (b *Bot) onInteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
    switch i.Type {
    case discordgo.InteractionApplicationCommand:
        b.handleCommand(s, i)
    case discordgo.InteractionMessageComponent:
        b.handleComponent(s, i)
    case discordgo.InteractionModalSubmit:
        b.handleModal(s, i)
    }
}

func (b *Bot) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
    data := i.ApplicationCommandData()
    
    switch data.Name {
    case "link":
        b.handleLinkCommand(s, i)
    case "ask":
        b.handleAskCommand(s, i)
    case "guide":
        b.handleGuideCommand(s, i)
    }
}
```

### 6.4 Configuration

**File:** `internal/config/config.go`

```go
package config

import (
    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    Discord struct {
        Token   string `envconfig:"DISCORD_TOKEN" required:"true"`
        GuildID string `envconfig:"DISCORD_GUILD_ID" required:"true"`
    }
    
    Database struct {
        Host     string `envconfig:"DB_HOST" default:"localhost"`
        Port     int    `envconfig:"DB_PORT" default:"5432"`
        User     string `envconfig:"DB_USER" default:"bot"`
        Password string `envconfig:"DB_PASSWORD" required:"true"`
        Name     string `envconfig:"DB_NAME" default:"livinglands"`
    }
    
    Ollama struct {
        URL             string `envconfig:"OLLAMA_URL" default:"http://localhost:11434"`
        Model           string `envconfig:"LLM_MODEL" default:"mistral:7b-instruct"`
        EmbeddingModel  string `envconfig:"EMBEDDING_MODEL" default:"nomic-embed-text"`
        MaxContextMsgs  int    `envconfig:"MAX_CONTEXT_MESSAGES" default:"10"`
    }
    
    Redis struct {
        URL string `envconfig:"REDIS_URL" default:"redis://localhost:6379"`
    }
    
    Chroma struct {
        URL string `envconfig:"CHROMA_URL" default:"http://localhost:8000"`
    }
    
    Hytale struct {
        APISecret       string `envconfig:"HYTALE_API_SECRET" required:"true"`
        VerifyCodeExpiry int   `envconfig:"VERIFY_CODE_EXPIRY" default:"600"`
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
    return &cfg, nil
}
```

### 6.5 Account Linking Service

**File:** `internal/services/account.go`

```go
package services

import (
    "crypto/rand"
    "encoding/base32"
    "fmt"
    "time"
    
    "gorm.io/gorm"
    "livinglands-bot/internal/database/models"
)

type AccountService struct {
    db     *gorm.DB
    expiry time.Duration
    logger *slog.Logger
}

func NewAccountService(db *gorm.DB, expirySeconds int, logger *slog.Logger) *AccountService {
    return &AccountService{
        db:     db,
        expiry: time.Duration(expirySeconds) * time.Second,
        logger: logger,
    }
}

// GenerateVerificationCode creates a new 8-char code for Discord user
func (s *AccountService) GenerateVerificationCode(discordID int64, discordUsername string) (string, error) {
    code := generateCode(8)
    
    user := &models.User{
        DiscordID:       discordID,
        DiscordUsername: discordUsername,
        VerificationCode: code,
    }
    
    err := s.db.Where("discord_id = ?", discordID).
        Assign(user).
        FirstOrCreate(user).Error
    
    if err != nil {
        return "", err
    }
    
    return code, nil
}

// VerifyLink validates code from Hytale and links accounts
func (s *AccountService) VerifyLink(code string, hytaleUsername string, hytaleUUID string) error {
    var user models.User
    
    err := s.db.Where("verification_code = ?", code).First(&user).Error
    if err != nil {
        return fmt.Errorf("invalid verification code")
    }
    
    // Check expiry (code valid for 10 minutes)
    if time.Since(user.UpdatedAt) > s.expiry {
        return fmt.Errorf("verification code expired")
    }
    
    // Update with Hytale info
    user.HytaleUsername = hytaleUsername
    user.HytaleUUID = hytaleUUID
    user.VerifiedAt = time.Now()
    user.VerificationCode = "" // Clear code
    
    return s.db.Save(&user).Error
}

func generateCode(length int) string {
    b := make([]byte, length)
    rand.Read(b)
    return base32.StdEncoding.EncodeToString(b)[:length]
}
```

### 6.6 LLM Service with Ollama

**File:** `pkg/ollama/client.go`

```go
package ollama

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    baseURL    string
    httpClient *http.Client
}

type GenerateRequest struct {
    Model   string `json:"model"`
    Prompt  string `json:"prompt"`
    System  string `json:"system,omitempty"`
    Context []int  `json:"context,omitempty"`
    Stream  bool   `json:"stream"`
    Options Options `json:"options,omitempty"`
}

type Options struct {
    Temperature float64 `json:"temperature,omitempty"`
    NumPredict  int     `json:"num_predict,omitempty"`
}

type GenerateResponse struct {
    Response string `json:"response"`
    Context  []int  `json:"context"`
    Done     bool   `json:"done"`
}

func NewClient(baseURL string) *Client {
    return &Client{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
    req.Stream = false
    
    body, err := json.Marshal(req)
    if err != nil {
        return nil, err
    }
    
    httpReq, err := http.NewRequestWithContext(ctx, "POST", 
        c.baseURL+"/api/generate", bytes.NewReader(body))
    if err != nil {
        return nil, err
    }
    
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("ollama returned %d", resp.StatusCode)
    }
    
    var genResp GenerateResponse
    if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
        return nil, err
    }
    
    return &genResp, nil
}

func (c *Client) Embed(ctx context.Context, model, text string) ([]float32, error) {
    req := struct {
        Model string `json:"model"`
        Prompt string `json:"prompt"`
    }{
        Model:  model,
        Prompt: text,
    }
    
    body, _ := json.Marshal(req)
    
    httpReq, _ := http.NewRequestWithContext(ctx, "POST",
        c.baseURL+"/api/embeddings", bytes.NewReader(body))
    httpReq.Header.Set("Content-Type", "application/json")
    
    resp, err := c.httpClient.Do(httpReq)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var embedResp struct {
        Embedding []float32 `json:"embedding"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
        return nil, err
    }
    
    return embedResp.Embedding, nil
}
```

**File:** `internal/services/llm.go`

```go
package services

import (
    "context"
    "fmt"
    
    "livinglands-bot/pkg/ollama"
)

type Personality struct {
    Name        string `yaml:"name"`
    Role        string `yaml:"role"`
    Tone        string `yaml:"tone"`
    Knowledge   string `yaml:"knowledge"`
    SystemPrompt string `yaml:"system_prompt"`
}

type LLMService struct {
    client     *ollama.Client
    model      string
    personality Personality
    logger     *slog.Logger
}

func NewLLMService(client *ollama.Client, model string, personality Personality, logger *slog.Logger) *LLMService {
    return &LLMService{
        client:      client,
        model:       model,
        personality: personality,
        logger:      logger,
    }
}

func (s *LLMService) GenerateResponse(ctx context.Context, userMessage string, context []string) (string, error) {
    // Build prompt with context
    prompt := s.buildPrompt(userMessage, context)
    
    req := ollama.GenerateRequest{
        Model:   s.model,
        Prompt:  prompt,
        System:  s.personality.SystemPrompt,
        Options: ollama.Options{
            Temperature: 0.7,
            NumPredict:  200,
        },
    }
    
    resp, err := s.client.Generate(ctx, req)
    if err != nil {
        s.logger.Error("llm generation failed", "error", err)
        return "", err
    }
    
    return resp.Response, nil
}

func (s *LLMService) buildPrompt(message string, context []string) string {
    var prompt string
    
    // Add RAG context
    if len(context) > 0 {
        prompt += "Use this information to answer:\n"
        for _, ctx := range context {
            prompt += fmt.Sprintf("- %s\n", ctx)
        }
        prompt += "\n"
    }
    
    prompt += fmt.Sprintf("User: %s\nAssistant:", message)
    return prompt
}
```

### 6.7 RAG Pipeline

**File:** `internal/services/rag.go`

```go
package services

import (
    "context"
    
    "github.com/amikos-tech/chroma-go"
    "livinglands-bot/pkg/ollama"
)

type RAGService struct {
    chromaClient *chroma.Client
    ollamaClient *ollama.Client
    collection   *chroma.Collection
    embedModel   string
    logger       *slog.Logger
}

func NewRAGService(chromaURL string, ollamaClient *ollama.Client, embedModel string, logger *slog.Logger) (*RAGService, error) {
    client, err := chroma.NewClient(chromaURL)
    if err != nil {
        return nil, err
    }
    
    collection, err := client.GetOrCreateCollection("livinglands_docs", nil)
    if err != nil {
        return nil, err
    }
    
    return &RAGService{
        chromaClient: client,
        ollamaClient: ollamaClient,
        collection:   collection,
        embedModel:   embedModel,
        logger:       logger,
    }, nil
}

func (s *RAGService) Query(ctx context.Context, question string, nResults int) ([]string, error) {
    // Generate embedding for question
    embedding, err := s.ollamaClient.Embed(ctx, s.embedModel, question)
    if err != nil {
        return nil, err
    }
    
    // Query ChromaDB
    results, err := s.collection.Query(
        [][]float32{embedding},
        nResults,
        nil,
        nil,
        nil,
    )
    if err != nil {
        return nil, err
    }
    
    // Extract documents
    var contexts []string
    if len(results.Documents) > 0 {
        for _, doc := range results.Documents[0] {
            contexts = append(contexts, doc)
        }
    }
    
    return contexts, nil
}

func (s *RAGService) AddDocuments(ctx context.Context, docs []Document) error {
    var embeddings [][]float32
    var texts []string
    var metadatas []map[string]interface{}
    var ids []string
    
    for i, doc := range docs {
        embed, err := s.ollamaClient.Embed(ctx, s.embedModel, doc.Text)
        if err != nil {
            s.logger.Error("failed to embed document", "doc", doc.ID, "error", err)
            continue
        }
        
        embeddings = append(embeddings, embed)
        texts = append(texts, doc.Text)
        metadatas = append(metadatas, doc.Metadata)
        ids = append(ids, doc.ID)
    }
    
    _, err := s.collection.Add(ids, embeddings, metadatas, texts)
    return err
}

type Document struct {
    ID       string
    Text     string
    Metadata map[string]interface{}
}
```

### 6.8 HTTP API (Hytale Integration)

**File:** `internal/api/server.go`

```go
package api

import (
    "net/http"
    
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "livinglands-bot/internal/config"
    "livinglands-bot/internal/services"
)

type Server struct {
    echo     *echo.Echo
    config   *config.Config
    services *services.Services
    logger   *slog.Logger
}

func New(cfg *config.Config, svcs *services.Services, logger *slog.Logger) *Server {
    e := echo.New()
    e.HideBanner = true
    
    // Middleware
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())
    e.Use(middleware.CORS())
    
    s := &Server{
        echo:     e,
        config:   cfg,
        services: svcs,
        logger:   logger,
    }
    
    // Routes
    e.GET("/health", s.healthCheck)
    e.POST("/api/v1/verify", s.verifyLink, s.authMiddleware)
    
    return s
}

func (s *Server) Start() error {
    return s.echo.Start(":8000")
}

func (s *Server) Stop() error {
    return s.echo.Shutdown(context.Background())
}

func (s *Server) healthCheck(c echo.Context) error {
    return c.JSON(http.StatusOK, map[string]string{
        "status": "healthy",
    })
}

func (s *Server) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        secret := c.Request().Header.Get("X-API-Secret")
        if secret != s.config.Hytale.APISecret {
            return c.JSON(http.StatusUnauthorized, map[string]string{
                "error": "unauthorized",
            })
        }
        return next(c)
    }
}
```

**File:** `internal/api/handlers/verify.go`

```go
package handlers

import (
    "net/http"
    
    "github.com/labstack/echo/v4"
    "livinglands-bot/internal/services"
)

type VerifyRequest struct {
    Code           string `json:"code" validate:"required,len=8"`
    HytaleUsername string `json:"hytale_username" validate:"required"`
    HytaleUUID     string `json:"hytale_uuid" validate:"required,uuid"`
}

type VerifyHandler struct {
    accountService *services.AccountService
}

func NewVerifyHandler(svc *services.AccountService) *VerifyHandler {
    return &VerifyHandler{accountService: svc}
}

func (h *VerifyHandler) Handle(c echo.Context) error {
    var req VerifyRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "invalid request",
        })
    }
    
    if err := c.Validate(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": err.Error(),
        })
    }
    
    if err := h.accountService.VerifyLink(req.Code, req.HytaleUsername, req.HytaleUUID); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": err.Error(),
        })
    }
    
    return c.JSON(http.StatusOK, map[string]string{
        "status": "success",
    })
}
```

## 7. Self-Hosting Deployment

### 7.1 Docker Compose

```yaml
version: '3.8'

services:
  bot:
    build:
      context: .
      dockerfile: Dockerfile
    env_file: .env
    environment:
      - DB_HOST=postgres
      - REDIS_URL=redis://redis:6379
      - CHROMA_URL=http://chromadb:8000
      - OLLAMA_URL=http://ollama:11434
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_started
      ollama:
        condition: service_started
      chromadb:
        condition: service_started
    volumes:
      - ./configs:/app/configs:ro
    ports:
      - "8000:8000"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "-qO-", "http://localhost:8000/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  ollama:
    image: ollama/ollama:latest
    volumes:
      - ollama_data:/root/.ollama
    environment:
      - OLLAMA_ORIGINS=*
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: livinglands
      POSTGRES_USER: bot
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U bot -d livinglands"]
      interval: 5s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data
    restart: unless-stopped

  chromadb:
    image: chromadb/chroma:latest
    volumes:
      - chroma_data:/chroma/chroma
    restart: unless-stopped

volumes:
  ollama_data:
  postgres_data:
  redis_data:
  chroma_data:
```

### 7.2 Dockerfile

```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bot cmd/bot/main.go

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates wget

# Copy binary from builder
COPY --from=builder /app/bot .
COPY --from=builder /app/configs ./configs

# Create non-root user
RUN adduser -D -u 1000 botuser
USER botuser

EXPOSE 8000

CMD ["./bot"]
```

### 7.3 Environment Variables (.env)

```bash
# Discord
DISCORD_TOKEN=your_bot_token_here
DISCORD_GUILD_ID=your_guild_id_here

# Database
DB_PASSWORD=secure_random_password
DB_HOST=postgres
DB_PORT=5432
DB_USER=bot
DB_NAME=livinglands

# LLM
OLLAMA_URL=http://ollama:11434
LLM_MODEL=mistral:7b-instruct
EMBEDDING_MODEL=nomic-embed-text
MAX_CONTEXT_MESSAGES=10

# Redis
REDIS_URL=redis://redis:6379

# ChromaDB
CHROMA_URL=http://chromadb:8000

# Hytale Integration
HYTALE_API_SECRET=webhook_secret_here
VERIFY_CODE_EXPIRY=600

# Bot Config
RATE_LIMIT_PER_MINUTE=5
LOG_LEVEL=info
PERSONALITY_FILE=configs/personality.yaml
```

### 7.4 Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores | 4 cores |
| RAM | 4GB | 8GB |
| Storage | 20GB SSD | 40GB SSD |
| Network | 10 Mbps | 25 Mbps |

**Note:** Ollama with Mistral 7B requires ~4GB RAM. For lighter setup, use `llama2:7b` (~3GB) or `phi` (~2GB).

### 7.5 Installation Steps

```bash
# 1. Clone repository
git clone https://github.com/yourusername/livinglands-bot.git
cd livinglands-bot

# 2. Configure environment
cp .env.example .env
# Edit .env with your Discord token and secrets

# 3. Build and start services
docker compose up -d

# 4. Download LLM models
docker compose exec ollama ollama pull mistral:7b-instruct
docker compose exec ollama ollama pull nomic-embed-text

# 5. Run database migrations
docker compose exec bot ./bot migrate

# 6. Index mod documentation
docker compose exec bot ./bot index-docs --path /path/to/mod/docs

# 7. Verify health
curl http://localhost:8000/health
```

### 7.6 Backup Strategy

```bash
#!/bin/bash
# backup.sh - Run daily via cron

set -e

BACKUP_DIR="/backups/livinglands/$(date +%Y%m%d_%H%M%S)"
mkdir -p $BACKUP_DIR

echo "Starting backup to $BACKUP_DIR"

# Database backup
docker compose exec -T postgres pg_dump -U bot livinglands > $BACKUP_DIR/database.sql
echo "Database backed up"

# Volume backups
docker run --rm \
    -v livinglands_postgres_data:/data \
    -v $BACKUP_DIR:/backup alpine \
    tar czf /backup/postgres.tar.gz -C /data .

docker run --rm \
    -v livinglands_ollama_data:/data \
    -v $BACKUP_DIR:/backup alpine \
    tar czf /backup/ollama.tar.gz -C /data .

docker run --rm \
    -v livinglands_chroma_data:/data \
    -v $BACKUP_DIR:/backup alpine \
    tar czf /backup/chroma.tar.gz -C /data .

echo "All volumes backed up"

# Optional: Sync to cloud
# rclone sync $BACKUP_DIR remote:backups/livinglands

# Cleanup old backups (keep 7 days)
find /backups/livinglands -type d -mtime +7 -exec rm -rf {} + 2>/dev/null || true

echo "Backup complete: $BACKUP_DIR"
```

### 7.7 Upgrade Process

```bash
# 1. Pull latest code
git pull origin main

# 2. Rebuild bot image
docker compose build bot

# 3. Run migrations
docker compose run --rm bot ./bot migrate

# 4. Restart services
docker compose up -d

# 5. Verify
docker compose logs -f bot
```

## 8. Security Considerations

1. **Token Security:** Never commit `.env` file, use Docker secrets for production
2. **API Authentication:** HMAC or API key validation for Hytale webhook
3. **Rate Limiting:** Redis-based per-user rate limiting on LLM queries
4. **Input Validation:** Use `go-playground/validator` for all inputs
5. **SQL Injection:** GORM provides parameterized queries by default
6. **XSS Prevention:** Discord handles message rendering, but validate user inputs
7. **Container Security:** Non-root user, read-only root filesystem where possible
8. **Network Isolation:** Services communicate via internal Docker network only
9. **Audit Logging:** Log all sensitive operations with user IDs and timestamps

## 9. Monitoring

**Health Endpoint:**
```go
// GET /health
{
  "status": "healthy",
  "services": {
    "discord": "connected",
    "database": "connected",
    "ollama": "available",
    "redis": "connected"
  },
  "version": "1.0.0",
  "uptime": "72h15m30s"
}
```

**Metrics (Prometheus):**
- `discord_messages_total` - Total messages processed
- `llm_requests_duration_seconds` - LLM response latency
- `ollama_errors_total` - Ollama error count
- `active_conversations` - Current active conversation threads

**Logging:**
- Structured JSON logging via `slog`
- Correlation IDs for tracing requests
- Separate log levels: DEBUG, INFO, WARN, ERROR

## 10. Development Workflow

```bash
# Setup
go mod init github.com/yourusername/livinglands-bot
go mod tidy

# Run locally (requires local Postgres, Redis, Ollama)
go run cmd/bot/main.go

# Run with hot reload (using air)
air

# Run tests
go test ./...

# Run with race detector
go run -race cmd/bot/main.go

# Build for production
go build -ldflags="-w -s" -o bot cmd/bot/main.go

# Database migrations (using golang-migrate)
migrate -path migrations -database "postgres://bot:password@localhost/livinglands?sslmode=disable" up

# Lint
golangci-lint run

# Format code
go fmt ./...
```

## 11. Go-Specific Patterns

### Context Cancellation
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

resp, err := s.llm.Generate(ctx, req)
if err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        return "", fmt.Errorf("llm request timed out")
    }
    return "", err
}
```

### Error Wrapping
```go
if err != nil {
    return fmt.Errorf("failed to generate verification code for user %d: %w", discordID, err)
}
```

### Graceful Shutdown
```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := bot.Stop(); err != nil {
    logger.Error("bot shutdown error", "error", err)
}
if err := api.Stop(); err != nil {
    logger.Error("api shutdown error", "error", err)
}
```

## 12. Future Enhancements

- [ ] Web dashboard (React + Go API)
- [ ] Metrics endpoint (Prometheus format)
- [ ] Distributed tracing (OpenTelemetry)
- [ ] Multi-server support
- [ ] Hytale server status monitoring
- [ ] Player statistics display
- [ ] Voice channel TTS integration
- [ ] Plugin system for custom commands

## 13. References

- **discordgo docs:** https://github.com/bwmarrin/discordgo
- **Echo Framework:** https://echo.labstack.com/
- **GORM:** https://gorm.io/
- **Ollama API:** https://github.com/ollama/ollama/blob/main/docs/api.md
- **ChromaDB Go:** https://github.com/amikos-tech/chroma-go
- **go-redis:** https://github.com/redis/go-redis
- **validator:** https://github.com/go-playground/validator

---

**Document Version:** 1.0  
**Last Updated:** 2026-01-31  
**Status:** Draft
