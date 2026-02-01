# Living Lands Discord Bot - Agent Instructions

## Project Overview

**Living Lands Discord Bot** is a companion Discord bot for the Living Lands Reloaded Hytale mod. It provides lore-friendly interactions, player account linking, mod Q&A with RAG, and channel navigation with a local LLM personality.

- **Language:** Go 1.25.6
- **Discord Library:** discordgo v0.29.0
- **Web Framework:** Fiber v2.52+
- **Databases:** PostgreSQL (relational) + ChromaDB (vector) + Redis (cache)
- **LLM:** Ollama (Mistral 7B) with RAG
- **Deployment:** Docker Compose (self-hosted)
- **Current Status:** Development phase

## Architecture

```
Discord Bot (Go)
├── cmd/bot/main.go              # Entry point
├── internal/
│   ├── bot/                     # Discord bot handlers
│   │   ├── bot.go              # Bot initialization
│   │   ├── commands.go         # Slash commands
│   │   ├── events.go           # Event handlers
│   │   └── interactions.go     # Button/modal handlers
│   ├── config/                  # Environment configuration
│   ├── database/               # GORM models & migrations
│   ├── services/               # Business logic
│   │   ├── account.go         # Account linking
│   │   ├── llm.go            # Ollama integration
│   │   ├── rag.go            # RAG pipeline
│   │   ├── welcome.go        # Welcome messages
│   │   └── channel.go        # Channel routing
│   ├── api/                    # HTTP API for Hytale
│   │   ├── server.go          # Fiber server
│   │   └── handlers/          # Route handlers
│   └── utils/                  # Utilities
├── pkg/ollama/                 # Ollama HTTP client
├── configs/                    # Personality YAML
├── migrations/                 # SQL migrations
└── docker-compose.yml
```

## Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Language | Go 1.25.6 | Application code |
| Discord | discordgo v0.29.0 | Discord API wrapper |
| Web Framework | Fiber v2.52+ | HTTP API (replaces Echo) |
| ORM | GORM v1.25+ | Database abstraction |
| Migrations | golang-migrate v4.17+ | Schema versioning |
| LLM | Ollama (Mistral 7B) | Local inference |
| Vector DB | chroma-go v0.1+ | Document embeddings |
| Cache | go-redis v9.4+ | Redis client |
| Validation | go-playground/validator v10+ | Struct validation |
| Config | envconfig v1.4+ | Environment parsing |
| Logging | slog (std lib) | Structured logging |

## Required Gateway Intents

```go
discordgo.IntentsGuildMembers      // Welcome new users (privileged!)
discordgo.IntentsGuildMessages     // Read messages
discordgo.IntentsMessageContent    // Read message content (privileged)
```

**Note:** `GUILD_MEMBERS` and `MESSAGE_CONTENT` require enabling in Discord Developer Portal.

## Key Patterns

### Discord Rate Limits
- Global: 50 requests/second
- Must respond to interactions within 3 seconds (or defer)
- Follow-ups allowed for 15 minutes after deferral
- Always check `X-RateLimit-*` headers

### Command Handler Pattern

```go
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

### Interaction Response Pattern

```go
// Immediate response (within 3 seconds)
s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseChannelMessageWithSource,
    Data: &discordgo.InteractionResponseData{
        Content: "Processing your request...",
    },
})

// Or defer for longer operations
s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
    Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
})

// Follow up later
s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
    Content: "Here's your answer!",
})
```

### LLM Service Pattern

```go
func (s *LLMService) GenerateResponse(ctx context.Context, question string, context []string) (string, error) {
    prompt := s.buildPrompt(question, context)
    
    req := ollama.GenerateRequest{
        Model:  s.model,
        Prompt: prompt,
        System: s.personality.SystemPrompt,
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
```

### RAG Pipeline Pattern

```go
func (s *RAGService) Query(ctx context.Context, question string, nResults int) ([]string, error) {
    // 1. Generate embedding
    embedding, err := s.ollamaClient.Embed(ctx, s.embedModel, question)
    if err != nil {
        return nil, err
    }
    
    // 2. Query ChromaDB
    results, err := s.collection.Query([][]float32{embedding}, nResults, nil, nil, nil)
    if err != nil {
        return nil, err
    }
    
    // 3. Extract documents
    var contexts []string
    for _, doc := range results.Documents[0] {
        contexts = append(contexts, doc)
    }
    
    return contexts, nil
}
```

### Fiber HTTP API Pattern

```go
func NewServer(cfg *config.Config, svcs *services.Services, logger *slog.Logger) *Server {
    app := fiber.New(fiber.Config{
        DisableStartupMessage: true,
    })
    
    // Middleware
    app.Use(logger.New())
    app.Use(recover.New())
    app.Use(cors.New())
    
    s := &Server{
        app:      app,
        config:   cfg,
        services: svcs,
        logger:   logger,
    }
    
    // Routes
    app.Get("/health", s.healthCheck)
    app.Post("/api/v1/verify", s.authMiddleware, s.verifyLink)
    
    return s
}

func (s *Server) healthCheck(c *fiber.Ctx) error {
    return c.JSON(fiber.Map{"status": "healthy"})
}
```

### Database Model Pattern

```go
type User struct {
    ID               uint      `gorm:"primarykey"`
    DiscordID        int64     `gorm:"uniqueIndex;not null"`
    DiscordUsername  string    `gorm:"not null"`
    HytaleUsername   string
    HytaleUUID       string
    VerificationCode string
    VerifiedAt       *time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

### Graceful Shutdown Pattern

```go
quit := make(chan os.Signal, 1)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

<-quit

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := bot.Stop(); err != nil {
    logger.Error("bot shutdown error", "error", err)
}
if err := server.ShutdownWithContext(ctx); err != nil {
    logger.Error("server shutdown error", "error", err)
}
```

## Logging

**Use slog (standard library)** for structured logging:

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Usage
logger.Info("user linked account", 
    "discord_id", user.DiscordID,
    "hytale_username", user.HytaleUsername,
)

logger.Error("llm request failed", 
    "error", err,
    "user_id", userID,
)
```

## Configuration

**Environment variables only** (12-factor app style):

```go
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
    // ... other configs
}
```

## Error Handling

**Always wrap errors with context:**

```go
if err != nil {
    return fmt.Errorf("failed to generate verification code for user %d: %w", discordID, err)
}
```

**Handle Discord API errors specifically:**

```go
if discordErr, ok := err.(*discordgo.RESTError); ok {
    if discordErr.Response.StatusCode == 429 {
        // Rate limited - extract retry-after header
        retryAfter := discordErr.Response.Header.Get("Retry-After")
    }
}
```

## Implementation Rules

1. **Database separation:**
   - PostgreSQL for relational data (users, links, routing rules)
   - ChromaDB for vector embeddings only
   - Redis for caching and session state

2. **Context propagation:** Always pass `context.Context` through the call chain for cancellation/timeouts

3. **Rate limiting:** Implement per-user rate limiting on LLM queries (Redis counter)

4. **Interaction timeouts:** Always defer if operation takes > 2 seconds

5. **Error resilience:** Bot should continue running even if LLM or DB is temporarily unavailable

6. **Structured logging:** Use slog with key-value pairs, never fmt.Printf

7. **No global state:** Pass dependencies via constructors (dependency injection)

8. **Webhook security:** Verify Hytale webhook requests with HMAC/API key

## Linear Project Management

**IMPORTANT:** Always use Linear MCP tools to keep the project board in sync with development progress.

### Linear Workflow Integration

1. **Before Starting Work**
   - Check Linear for assigned issues: `linear_list_issues` (filter by assignee, status)
   - Update issue status to "in_progress" when beginning work: `linear_update_issue`
   - Review issue details for context: `linear_get_issue`

2. **During Development**
   - Keep Linear issues updated as progress is made
   - Add comments with progress updates: `linear_add_comment`
   - Create new issues for discovered bugs/features: `linear_create_issue`

3. **Before Creating PRs**
   - Ensure related Linear issue exists
   - Update Linear issue with completion status
   - Reference Linear issue key in PR description (e.g., "Closes: LLB-123")

4. **Using linearapp Agent**
   Delegate complex Linear workflows to the `linearapp` agent:
   - Sprint planning and cycle management
   - Breaking down features into subtasks
   - Organizing backlog with labels and priorities
   - Generating status reports

   Example:
   ```bash
   # Let linearapp agent handle sprint planning
   /linearapp Create a sprint for Phase 1 (Database setup)
   ```

### Linear Issue References in Git

**Branch Naming with Linear:**
- Feature: `feature/LLB-123-short-description`
- Bug fix: `fix/LLB-456-short-description`
- Use Linear issue key as branch prefix for automatic linking

**Commit Messages with Linear:**
- Reference Linear issues: `feat: add /link command (LLB-123)`
- Auto-close issues: `fix: handle rate limit errors (Closes LLB-456)`

**PR Description Template:**
```markdown
## Summary
Brief overview of what this PR accomplishes

**Closes:** LLB-123, LLB-456
**Related:** LLB-789 (partial implementation)
**Version:** 0.1.0

### Features
- Feature 1
- Feature 2

### Bug Fixes
- Fix 1
```

### Proactive Linear Checks

Claude Code should AUTOMATICALLY check Linear when:
- Starting a new feature or bug fix
- Creating a pull request
- Completing a development phase
- User mentions "task", "issue", "sprint", or "backlog"

## Agent Usage Guidelines

### Primary Agents (Auto-proactive)

| Agent | When to Use | For What |
|-------|-------------|----------|
| **code-review** | After writing significant code | Review Go code quality, error handling, concurrency |
| **architecture-review** | After implementing features | Check architectural patterns, service boundaries |
| **git-flow-manager** | Branch/PR management | Feature branches, releases |
| **error-detective** | Debugging issues | Analyze logs, investigate errors |
| **golang-backend-api** | Core backend development | Go-specific patterns, Discord API integration |
| **linearapp** | Sprint planning, issue tracking | Creating cycles, organizing backlog, status reports |

### When to Use Each Agent

**golang-backend-api:**
- Discord bot command implementations
- Database model design
- API endpoint handlers
- Service layer logic
- Fiber HTTP routing

**code-review:**
- After implementing commands
- After adding database migrations
- Before creating PRs
- After service implementations

**architecture-review:**
- Adding new services
- Refactoring service boundaries
- Database schema changes
- Major feature additions

**linearapp:**
- Sprint planning and cycle creation
- Breaking down features into subtasks
- Organizing backlog with labels
- Generating status reports
- Issue estimation and prioritization

## Pull Request Guidelines

### Branch Naming
- Feature: `feature/short-description`
- Bug fix: `fix/short-description`
- Refactor: `refactor/component-name`

### Commit Messages
Follow conventional commits:
- `feat: add /link command`
- `fix: handle rate limit errors`
- `refactor: extract llm client to pkg/`

### PR Checklist

**Code Quality:**
- [ ] Follows Go conventions (gofmt, golint)
- [ ] Proper error handling with wrapping
- [ ] Context cancellation respected
- [ ] No global variables
- [ ] Structured logging used

**Performance:**
- [ ] Database queries are efficient
- [ ] LLM requests have timeouts
- [ ] Goroutines properly synchronized
- [ ] No goroutine leaks

**Testing:**
- [ ] Unit tests for services
- [ ] Manual Discord testing
- [ ] Rate limiting tested
- [ ] Graceful shutdown verified

**Documentation:**
- [ ] Comments for exported functions
- [ ] README.md updated (if user-facing)
- [ ] TECHNICAL_DESIGN.md updated (if architecture changes)

## Deployment

### Local Development

```bash
# Start dependencies
docker compose up -d postgres redis ollama chromadb

# Run bot
go run cmd/bot/main.go
```

### Production Deployment

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env with production values

# 2. Start all services
docker compose up -d

# 3. Download LLM models
docker compose exec ollama ollama pull mistral:7b-instruct
docker compose exec ollama ollama pull nomic-embed-text

# 4. Run migrations
docker compose exec bot ./bot migrate

# 5. Index documents
docker compose exec bot ./bot index-docs --path /path/to/docs
```

### Backup

```bash
# Daily automated backup (run via cron)
./scripts/backup.sh
```

## Environment Variables

See `.env.example` for all required variables:
- `DISCORD_TOKEN` - Bot token from Discord Developer Portal
- `DISCORD_GUILD_ID` - Your Discord server ID
- `DB_PASSWORD` - PostgreSQL password
- `HYTALE_API_SECRET` - Secret for Hytale webhook verification
- `OLLAMA_URL` - Ollama endpoint (default: http://ollama:11434)

## MCP Server Integration

### Discord MCP Server

Use the Discord MCP server for bot announcements and webhook management:

```bash
# List configured webhooks
discord_list_webhooks

# Add a new webhook
discord_add_webhook --name "releases" --url "https://discord.com/api/webhooks/YOUR_ID/YOUR_TOKEN"

# Send announcement
discord_send_announcement \
  --version "v0.1.0" \
  --headline "Initial Release!" \
  --changes "Account linking,LLM Q&A,Channel navigation" \
  --style "release"
```

**Available Tools:**
- `discord_send_message` - General messages
- `discord_send_announcement` - Release announcements
- `discord_send_teaser` - Preview announcements
- `discord_send_changelog` - Detailed changelogs
- `discord_add_webhook` - Configure webhooks
- `discord_list_webhooks` - View webhooks

### GitHub MCP Server

Use for pull request management and code operations:

**Key Tools:**
- `github_create_pull_request` - Create PRs
- `github_pull_request_review_write` - Request reviews
- `github_list_issues` - View issues
- `github_search_code` - Find code patterns
- `github_push_files` - Commit changes

### Linear MCP Server

Use for project tracking (primary):

**Key Tools:**
- `linear_list_issues` - View issues by filter
- `linear_get_issue` - Get issue details
- `linear_update_issue` - Update status/assignee
- `linear_create_issue` - Create new issues
- `linear_add_comment` - Add progress updates
- `linear_list_cycles` - View sprints
- `linear_create_cycle` - Create new cycles

---

## Pull Request Guidelines

### PR Creation Process

1. **Branch Naming**
   - Feature: `feature/LLB-123-short-description`
   - Bug fix: `fix/LLB-456-short-description`
   - Refactor: `refactor/component-name`
   - Use Linear issue key as branch prefix

2. **Commit Messages**
   - Follow conventional commits: `type: description`
   - Types: `feat`, `fix`, `refactor`, `perf`, `docs`, `test`
   - Example: `feat: add /link command (LLB-123)`
   - Reference Linear issues in commits

3. **PR Checklist**

**Code Quality:**
- [ ] Follows Go conventions (gofmt, golint)
- [ ] Proper error handling with wrapping
- [ ] Context cancellation respected
- [ ] No global variables
- [ ] Structured logging used

**Performance:**
- [ ] Database queries are efficient
- [ ] LLM requests have timeouts
- [ ] Goroutines properly synchronized
- [ ] No goroutine leaks

**Testing:**
- [ ] Unit tests for services
- [ ] Manual Discord testing
- [ ] Rate limiting tested
- [ ] Graceful shutdown verified

**Documentation:**
- [ ] Comments for exported functions
- [ ] README.md updated (if user-facing)
- [ ] TECHNICAL_DESIGN.md updated (if architecture changes)

## Key References

### Documentation
- `docs/TECHNICAL_DESIGN.md` - Full architecture and API details
- `docs/IMPLEMENTATION_PLAN.md` - Phased development plan
- `docs/ROADMAP.md` - Future features and versions
- Discord Developer Portal: https://discord.com/developers/applications

### External Libraries
- **discordgo:** https://github.com/bwmarrin/discordgo
- **Fiber:** https://docs.gofiber.io/
- **GORM:** https://gorm.io/
- **Ollama API:** https://github.com/ollama/ollama/blob/main/docs/api.md

---

**Document Version:** 1.1  
**Last Updated:** 2026-01-31  
**Status:** Draft
