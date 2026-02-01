# Living Lands Discord Bot

A companion Discord bot for the **Living Lands Reloaded** Hytale mod. Features lore-friendly interactions, player account linking, intelligent mod Q&A using local LLM with RAG, and channel navigation.

## Features

- **Welcome System** - Lore-friendly welcome messages for new members
- **Account Linking** - Link Discord accounts to Hytale usernames via verification codes
- **Intelligent Q&A** - Answer mod questions using local LLM (Ollama) with RAG (Retrieval-Augmented Generation)
- **Channel Navigation** - Direct users to appropriate channels (bug reports, changelog, wiki)
- **Living Personality** - In-character bot responses that feel alive
- **Rate Limiting** - Protect against API abuse with per-user request limits (Redis-backed)
- **Document Indexing** - Automatically index Markdown and TXT files for RAG knowledge base
- **Comprehensive Testing** - Unit tests for all services (32 tests, all passing)

## Tech Stack

- **Language:** Go 1.25.6
- **Discord Library:** [discordgo](https://github.com/bwmarrin/discordgo) v0.29.0
- **Web Framework:** [Fiber](https://docs.gofiber.io/)
- **Database:** PostgreSQL (relational) + ChromaDB (vector) + Redis (cache)
- **LLM:** Ollama running Mistral 7B + nomic-embed-text
- **Deployment:** Docker Compose

## Architecture

```
Discord Bot (Go)
├── cmd/bot/main.go              # Entry point
├── internal/
│   ├── bot/                     # Discord bot handlers
│   │   ├── bot.go              # Bot initialization
│   │   ├── commands.go         # Slash commands (/link, /ask, /guide)
│   │   ├── events.go           # Event handlers (welcome, messages)
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
└── docker-compose.yml          # Service orchestration
```

## Prerequisites

- Go 1.25.6 or higher
- Docker & Docker Compose
- Discord Bot Token ([Create one here](https://discord.com/developers/applications))
- Hytale server with webhook capability (for account linking)

## Quick Start

### 1. Clone & Configure

```bash
git clone https://github.com/yourusername/livinglands-bot.git
cd livinglands-bot
cp .env.example .env
# Edit .env with your Discord token and other secrets
```

### 2. Start Dependencies

```bash
docker compose up -d postgres redis ollama chromadb
```

### 3. Download LLM Models

```bash
docker compose exec ollama ollama pull mistral:7b-instruct
docker compose exec ollama ollama pull nomic-embed-text
```

### 4. Run Database Migrations

```bash
./bot migrate
```

### 5. Index Mod Documentation (for RAG)

```bash
./bot index-docs --path ./docs
```

This will:
- Recursively find all .md and .txt files
- Split documents into semantic chunks (500 chars, 50 char overlap)
- Generate embeddings using Ollama
- Store in ChromaDB for retrieval
- Skip already-indexed files based on checksums

### 6. Start the Bot

```bash
go run cmd/bot/main.go
# or with docker
docker compose up -d bot
```

### 7. Verify Health

```bash
curl http://localhost:8000/health
```

## Discord Commands

| Command | Description |
|---------|-------------|
| `/link` | Generate verification code to link Hytale account |
| `/ask <question>` | Ask about Living Lands mod (AI-powered with RAG, rate limited to 5/min) |
| `/guide` | Get directions to channels (bug reports, changelog, wiki) |

## CLI Commands

```bash
# Start bot in normal mode
./bot

# Run database migrations
./bot migrate

# Index documentation for RAG knowledge base
./bot index-docs --path <directory_or_file>

# Show help
./bot help
```

### Example: Index Documentation

```bash
# Index a single file
./bot index-docs --path ./docs/guide.md

# Index entire directory recursively
./bot index-docs --path ./docs/

# Output:
# [INFO] starting document indexing path=./docs
# [INFO] file processed path=./docs/guide.md chunks=15
# [INFO] document indexing complete processed_files=8 total_chunks=127
```

## Environment Variables

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
RATE_LIMIT_PER_MINUTE=5         # Max /ask requests per user per minute
LOG_LEVEL=info                  # debug, info, warn, error
PERSONALITY_FILE=configs/personality.yaml
HTTP_ADDR=:8000                 # API server bind address
```

## Development

### Local Development

```bash
# Start only dependencies
docker compose up -d postgres redis ollama chromadb

# Run bot locally
go run cmd/bot/main.go

# Run with hot reload
air
```

### Running Tests

```bash
go test ./...
```

### Database Migrations

```bash
# Create migration
migrate create -ext sql -dir migrations -seq migration_name

# Apply migrations
migrate -path migrations -database "postgres://bot:password@localhost/livinglands?sslmode=disable" up

# Rollback
migrate -path migrations -database "postgres://bot:password@localhost/livinglands?sslmode=disable" down 1
```

## Docker Deployment

### Production Setup

```bash
# 1. Configure environment
cp .env.example .env
# Edit with production values

# 2. Build and start
docker compose up -d

# 3. Verify health
curl http://localhost:8000/health
```

### Backup

```bash
# Run automated backup
./scripts/backup.sh
```

### Upgrade

```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker compose build bot
docker compose up -d
```

## Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores | 4 cores |
| RAM | 4GB | 8GB |
| Storage | 20GB SSD | 40GB SSD |
| Network | 10 Mbps | 25 Mbps |

**Note:** Ollama requires ~4GB RAM for Mistral 7B. For lighter setups, use `llama2:7b` (~3GB) or `phi` (~2GB).

## Documentation

- [Technical Design](docs/TECHNICAL_DESIGN.md) - Full architecture and API details
- [Discord API Notes](docs/DISCORD_API.md) - Discord-specific implementation
- [Agent Instructions](AGENTS.md) - Development guidelines for contributors

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'feat: add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

[MIT](LICENSE)

## Related Projects

- [Living Lands Reloaded](https://github.com/yourusername/hytale-livinglands) - The Hytale mod this bot accompanies

---

**Status:** MVP Complete ✅  
**Version:** 0.1.0-MVP  
**Last Updated:** 2026-02-01  
**Features Implemented:** All critical features for MVP release
