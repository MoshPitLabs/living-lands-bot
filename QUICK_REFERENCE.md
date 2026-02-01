# Living Lands Bot - Quick Reference Guide

## 🚀 Getting Started

### First Time Setup
```bash
# 1. Start dependencies
docker compose up -d postgres redis ollama chromadb

# 2. Download LLM models
docker compose exec ollama ollama pull mistral:7b-instruct
docker compose exec ollama ollama pull nomic-embed-text

# 3. Run migrations
go run cmd/bot/main.go migrate

# 4. Index documentation
go run cmd/bot/main.go index-docs --path ./docs

# 5. Start bot
go run cmd/bot/main.go
```

## 📋 CLI Commands

| Command | Usage | Notes |
|---------|-------|-------|
| `migrate` | `./bot migrate` | Run database migrations |
| `index-docs` | `./bot index-docs --path <dir>` | Index files for RAG |
| `help` | `./bot help` | Show help message |
| (none) | `./bot` | Start bot normally |

## 🔧 Environment Variables

```bash
# Discord
DISCORD_TOKEN=your_bot_token
DISCORD_GUILD_ID=your_guild_id

# Database
DB_PASSWORD=password
DB_HOST=localhost
DB_PORT=5432
DB_USER=bot
DB_NAME=livinglands

# Redis (NEW!)
REDIS_URL=redis://localhost:6379
RATE_LIMIT_PER_MINUTE=5

# Ollama
OLLAMA_URL=http://localhost:11434
LLM_MODEL=mistral:7b-instruct
EMBEDDING_MODEL=nomic-embed-text

# ChromaDB
CHROMA_URL=http://localhost:8000

# Hytale
HYTALE_API_SECRET=webhook_secret
VERIFY_CODE_EXPIRY=600

# Misc
LOG_LEVEL=info
HTTP_ADDR=:8000
PERSONALITY_FILE=configs/personality.yaml
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run service tests with verbose output
go test ./internal/services -v

# Run with coverage
go test ./internal/services -cover

# Run specific test
go test ./internal/services -run TestRateLimiter
```

## 🏗️ Project Structure

```
living-lands-bot/
├── cmd/bot/main.go              # Entry point
├── internal/
│   ├── bot/                     # Discord bot
│   │   ├── bot.go
│   │   ├── commands.go
│   │   ├── events.go
│   │   └── interactions.go
│   ├── services/                # Business logic
│   │   ├── account.go
│   │   ├── channel.go
│   │   ├── indexer.go           # NEW!
│   │   ├── llm.go
│   │   ├── rag.go
│   │   ├── rate_limiter.go      # NEW!
│   │   ├── welcome.go
│   │   └── *_test.go            # 32 tests total
│   ├── config/
│   ├── database/
│   ├── api/
│   └── utils/
├── pkg/ollama/                  # Ollama client
├── migrations/                  # SQL migrations
├── configs/                     # YAML configs
└── docker-compose.yml
```

## 💾 Database

```bash
# Create migration
migrate create -ext sql -dir migrations -seq add_feature

# Apply migrations
./bot migrate

# Check migrations
sqlite3 livinglands.db ".tables"
```

## 🐳 Docker Compose

```bash
# Start all services
docker compose up -d

# Start specific services
docker compose up -d postgres redis ollama chromadb

# View logs
docker compose logs -f bot

# Stop all
docker compose down

# Rebuild bot image
docker compose build bot
```

## 📊 Monitoring

```bash
# Health check
curl http://localhost:8000/health

# View bot logs
docker compose logs bot

# Monitor Redis
redis-cli
> KEYS "rate_limit:*"
> INFO stats
```

## 🔍 Debugging

```bash
# Enable debug logging
LOG_LEVEL=debug go run cmd/bot/main.go

# Build binary for debugging
go build -o bot cmd/bot/main.go
./bot

# Test database connection
go run cmd/bot/main.go migrate

# Test indexing
./bot index-docs --path ./test-docs
```

## 🔑 Key Files

| File | Purpose | Language |
|------|---------|----------|
| `cmd/bot/main.go` | Entry point | Go |
| `internal/bot/bot.go` | Bot core | Go |
| `internal/bot/commands.go` | Slash commands | Go |
| `internal/services/rate_limiter.go` | Rate limiting | Go |
| `internal/services/indexer.go` | Document indexing | Go |
| `configs/personality.yaml` | Bot personality | YAML |
| `migrations/*.sql` | DB schema | SQL |

## ✨ New Features (MVP)

### Rate Limiting
- Protects `/ask` command from abuse
- 5 requests per minute per user (configurable)
- Redis-backed persistence
- User-friendly error messages

### Document Indexing
- CLI: `./bot index-docs --path <directory>`
- Supports .md and .txt files
- Automatic chunking and embedding
- Checksum-based deduplication

### Tests
- 32 unit tests (all passing)
- Service layer coverage
- No database dependencies
- Fast execution (0.004s)

## 🐛 Troubleshooting

| Problem | Solution |
|---------|----------|
| Bot won't start | Check `DISCORD_TOKEN` in `.env` |
| Rate limiting not working | Check `REDIS_URL` connection |
| Documents not indexing | Check file permissions and path |
| Tests failing | Run `go mod tidy` then rebuild |
| High memory usage | Monitor Ollama container |

## 📚 Documentation

- **Technical Details:** See `docs/MVP_IMPLEMENTATION.md`
- **Full Reference:** See `COMPLETION_REPORT.md`
- **Implementation:** See `IMPLEMENTATION_SUMMARY.md`
- **User Guide:** See `README.md`

## 🚢 Deployment

```bash
# Production checklist
- [ ] Set DISCORD_TOKEN
- [ ] Set DB_PASSWORD
- [ ] Set REDIS_URL
- [ ] Run migrations
- [ ] Index documents
- [ ] Test /link command
- [ ] Test /ask command
- [ ] Test /guide command
- [ ] Monitor logs for errors
```

## 📝 Common Tasks

### Adding a new Discord command
1. Update `commands.go` - Add command definition
2. Implement handler function
3. Add tests in `*_test.go`
4. Register in `RegisterCommands()`

### Indexing new documentation
```bash
./bot index-docs --path ./path/to/docs
```

### Checking rate limits
```bash
redis-cli
> KEYS "rate_limit:*"
> GET rate_limit:user123
```

### Viewing bot logs
```bash
docker compose logs -f bot
# or locally
LOG_LEVEL=debug go run cmd/bot/main.go 2>&1 | grep ERROR
```

## 🔗 Useful Links

- [Discord.py Documentation](https://discordgo.readthedocs.io/)
- [Ollama API](https://github.com/ollama/ollama/blob/main/docs/api.md)
- [ChromaDB Documentation](https://docs.trychroma.com/)
- [GORM Documentation](https://gorm.io/)
- [Go Testing Package](https://golang.org/pkg/testing/)

## ⚡ Performance Tips

1. **Rate Limiter:**
   - Increase `RATE_LIMIT_PER_MINUTE` for trusted users
   - Monitor Redis memory usage

2. **Document Indexing:**
   - Index in batches of 10-20 files
   - Monitor Ollama CPU usage
   - Increase chunk overlap for similar documents

3. **Bot Performance:**
   - Set `LOG_LEVEL=info` in production
   - Monitor goroutine count
   - Use `GOMAXPROCS` if needed

## 🎯 What's Next?

1. **Phase 6 Tasks:**
   - Integration testing
   - E2E testing
   - Performance optimization
   - Monitoring setup

2. **Future Features:**
   - Database-driven channel guide
   - Admin rate limit bypass
   - Advanced metrics
   - Web dashboard

---

**Last Updated:** February 1, 2026  
**Version:** 0.1.0-MVP  
**Status:** Ready for Integration Testing
