# Living Lands Bot - Documentation Indexing Guide

## Problem Solved

The bot was giving generic Hytale responses instead of Living Lands-specific answers because **no Living Lands documentation was indexed in ChromaDB**.

## Solution

We created comprehensive Living Lands documentation and indexed it into the RAG system.

## Documentation Structure

All Living Lands documentation is in `docs/livinglands/`:

1. **overview.md** - General introduction and core features
2. **metabolism.md** - Hunger, thirst, and energy systems
3. **professions.md** - Combat, Mining, Logging, Building, Gathering
4. **commands.md** - Complete command reference
5. **installation.md** - Setup guide for servers and players
6. **faq.md** - Common questions and troubleshooting

## Indexing Process

### Prerequisites

1. **Start required services**:
   ```bash
   docker compose up -d postgres redis ollama chromadb
   ```

2. **Wait for services to be ready**:
   ```bash
   docker compose ps
   # All services should show "Up" status
   ```

3. **Build the bot**:
   ```bash
   go build -o bin/bot cmd/bot/main.go
   ```

### Index the Documentation

**Method 1: Using environment file (recommended)**

```bash
# Create indexing environment file
cat > .env.indexing <<'EOF'
DISCORD_TOKEN=dummy
DISCORD_GUILD_ID=dummy
DB_HOST=localhost
DB_PORT=5432
DB_USER=bot
DB_PASSWORD=$(grep "DB_PASSWORD" .env | cut -d'=' -f2)
DB_NAME=livinglands
DB_SSLMODE=disable
REDIS_URL=redis://localhost:6379
CHROMA_URL=http://localhost:8001
OLLAMA_URL=http://localhost:11434
LLM_MODEL=mistral:7b-instruct
EMBEDDING_MODEL=nomic-embed-text
OLLAMA_TIMEOUT=60
LLM_FAST_MAX_TOKENS=60
LLM_FAST_TEMPERATURE=0.5
LLM_STANDARD_MAX_TOKENS=120
LLM_STANDARD_TEMPERATURE=0.6
LLM_DEEP_MAX_TOKENS=180
LLM_DEEP_TEMPERATURE=0.7
HYTALE_API_SECRET=dummy
VERIFY_CODE_EXPIRY=600
RATE_LIMIT_PER_MINUTE=5
MAX_CONTEXT_MESSAGES=10
LOG_LEVEL=info
PERSONALITY_FILE=configs/personality.yaml
EOF

# Run indexing
set -a && source .env.indexing && set +a && ./bin/bot index-docs --path ./docs/livinglands
```

**Method 2: Using Docker (production)**

```bash
docker compose exec bot ./bot index-docs --path /app/livinglands-docs
```

### Expected Output

You should see output like:
```
{"level":"INFO","msg":"starting document indexing","path":"./docs/livinglands"}
{"level":"INFO","msg":"file processed","path":"docs/livinglands/overview.md","chunks":4}
{"level":"INFO","msg":"file processed","path":"docs/livinglands/metabolism.md","chunks":5}
{"level":"INFO","msg":"file processed","path":"docs/livinglands/professions.md","chunks":7}
{"level":"INFO","msg":"file processed","path":"docs/livinglands/commands.md","chunks":8}
{"level":"INFO","msg":"file processed","path":"docs/livinglands/installation.md","chunks":10}
{"level":"INFO","msg":"file processed","path":"docs/livinglands/faq.md","chunks":14}
{"level":"INFO","msg":"collection created","collection":"livinglands_docs"}
{"level":"INFO","msg":"documents added to rag collection","count":48}
{"level":"INFO","msg":"document indexing complete","processed_files":6,"total_chunks":48}
```

## Testing RAG Retrieval

After indexing, verify the system works:

```bash
# Test RAG retrieval
go run test_rag.go
```

Expected results:
- Queries about Living Lands features return relevant documentation
- Distances (similarity scores) are low (< 0.5 is good)
- All queries return 3+ relevant document chunks

## Verifying in Discord

1. **Start the bot**:
   ```bash
   docker compose up -d bot
   ```

2. **Test with Discord commands**:
   ```
   /ask What features does Living Lands have?
   /ask How do professions work?
   /ask What is the metabolism system?
   ```

3. **Expected behavior**:
   - Bot responds with Living Lands-specific information
   - Mentions features like hunger/thirst/energy, professions, etc.
   - References commands like `/ll stats`, `/ll professions`

## Re-indexing (When Documentation Changes)

When you update or add documentation:

1. **Stop the bot**:
   ```bash
   docker compose stop bot
   ```

2. **Re-run indexing**:
   ```bash
   set -a && source .env.indexing && set +a && ./bin/bot index-docs --path ./docs/livinglands
   ```

3. **Restart the bot**:
   ```bash
   docker compose up -d bot
   ```

**Note**: Re-indexing creates new document IDs (based on content hash), so old versions are automatically replaced.

## Troubleshooting

### "Collection already exists" error
This is normal and harmless. The indexer will use the existing collection.

### "Failed to generate embedding" errors
Check that:
- Ollama is running: `docker compose ps ollama`
- Embedding model is downloaded: `docker compose exec ollama ollama list`
- If missing: `docker compose exec ollama ollama pull nomic-embed-text`

### RAG returns no results
Check:
- Documents were indexed successfully (check logs)
- ChromaDB is running: `curl http://localhost:8001/api/v1/heartbeat`
- Collection exists and has documents

### Bot still gives generic answers
Verify:
1. Indexing completed successfully (48 chunks for current docs)
2. Bot is using correct ChromaDB URL in .env: `CHROMA_URL=http://chromadb:8000`
3. RAG query logs show "document accepted" messages
4. Intent classification is detecting "IntentKnowledge" for your questions

## Adding More Documentation

To add new Living Lands documentation:

1. Create new `.md` files in `docs/livinglands/`
2. Follow existing structure (clear headings, organized sections)
3. Re-run indexing command
4. Test with Discord to verify new content is accessible

### Documentation Best Practices

- **Keep chunks meaningful**: Aim for 300-700 characters per logical section
- **Use clear headings**: Helps with semantic search
- **Include keywords**: Use terms players will search for
- **Cross-reference**: Link related topics
- **Stay focused**: One topic per file

## Monitoring

Check RAG performance in logs:

```bash
docker compose logs bot | grep "rag query\|document accepted"
```

Look for:
- **distance** values (lower = more relevant, < 0.5 is excellent)
- **results** count (should be 3-5 per query)
- **filtered** count (high values mean relevance threshold is too strict)

## Advanced: Tuning Relevance Threshold

Default threshold is `1.0` (permissive). To adjust:

**In code** (`internal/services/rag.go`):
```go
const DefaultRelevanceThreshold = 0.8  // More strict
```

**Or dynamically**:
```go
ragService.SetRelevanceThreshold(0.7)  // Higher precision
```

Lower values = fewer but more relevant results.
Higher values = more results but potentially less relevant.

---

## Summary

1. ✅ Created comprehensive Living Lands documentation (6 files, 48 chunks)
2. ✅ Indexed documentation into ChromaDB
3. ✅ Verified RAG retrieval works correctly
4. ✅ Bot now gives Living Lands-specific answers

The bot is ready to answer questions about Living Lands Reloaded!
