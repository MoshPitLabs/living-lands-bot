# Living Lands Discord Bot - Proxmox Deployment Guide

## Overview

This guide walks you through deploying the Living Lands Discord Bot to a Proxmox Docker host using Docker Compose.

## Prerequisites

### On Your Local Machine
- SSH access to Proxmox host
- `rsync` installed
- `.env` file configured with production values
- Training data available (or ready to transfer)

### On Proxmox Host
- Docker and Docker Compose installed
- At least 8GB RAM available
- 20GB disk space minimum (50GB+ recommended for Ollama models)
- Open ports: 8000 (bot HTTP API)

## Quick Deployment

### 1. Prepare Environment File

Create/update `.env` file with production values:

```bash
# Copy example and edit
cp .env.example .env
nano .env
```

Required variables:
```env
# Discord
DISCORD_TOKEN=your_discord_bot_token_here
DISCORD_GUILD_ID=your_guild_id_here

# Database
DB_PASSWORD=strong_random_password_here

# Ollama
OLLAMA_MODEL=mistral:7b-instruct
OLLAMA_EMBED_MODEL=nomic-embed-text

# Optional: ChromaDB auth
CHROMA_AUTH_TOKEN=random_token_for_chroma
CHROMA_AUTH_PROVIDER=chromadb.auth.token.TokenAuthServerProvider
```

### 2. Run Deployment Script

```bash
./deploy-to-proxmox.sh user@proxmox.local
```

Or specify custom deploy path:
```bash
./deploy-to-proxmox.sh user@proxmox.local /opt/my-bot
```

The script will:
1. Create deployment directory on Proxmox
2. Copy all project files
3. Transfer environment configuration
4. Transfer training data (~3GB)
5. Build Docker images
6. Start all services
7. Download LLM models
8. Run database migrations
9. Index documentation

**Estimated time:** 15-30 minutes (depending on network speed and hardware)

## Manual Deployment

If the automated script doesn't work, follow these steps:

### 1. Transfer Files to Proxmox

```bash
# SSH into Proxmox
ssh user@proxmox.local

# Create deployment directory
sudo mkdir -p /opt/living-lands-bot
cd /opt/living-lands-bot

# From your local machine, sync files
rsync -avz --exclude '.git' \
  ~/Development/living-lands-bot/ \
  user@proxmox.local:/opt/living-lands-bot/

# Copy environment file
scp .env user@proxmox.local:/opt/living-lands-bot/.env

# Transfer training data
rsync -avz /mnt/ugreen-nas/Coding/Hytale/DiscordBot/training-data/ \
  user@proxmox.local:/opt/living-lands-bot/training-data/
```

### 2. Start Services

```bash
# SSH into Proxmox
ssh user@proxmox.local
cd /opt/living-lands-bot

# Build and start services
docker-compose -f docker-compose.prod.yml up -d
```

### 3. Download LLM Models

```bash
# Download Mistral 7B (2-5 minutes)
docker-compose -f docker-compose.prod.yml exec ollama \
  ollama pull mistral:7b-instruct

# Download embedding model (1-2 minutes)
docker-compose -f docker-compose.prod.yml exec ollama \
  ollama pull nomic-embed-text
```

### 4. Initialize Database

```bash
# Run migrations
docker-compose -f docker-compose.prod.yml exec bot ./bot migrate

# Index documentation
docker-compose -f docker-compose.prod.yml exec bot \
  ./bot index-docs --path /app/training-data
```

## Verify Deployment

### Check Service Status

```bash
docker-compose -f docker-compose.prod.yml ps
```

All services should show "Up" and "healthy":
- `bot` - Discord bot application
- `postgres` - PostgreSQL database
- `redis` - Redis cache
- `ollama` - LLM inference server
- `chromadb` - Vector database

### Check Health Endpoint

```bash
curl http://localhost:8000/health
```

Should return: `{"status":"healthy"}`

### View Logs

```bash
# All services
docker-compose -f docker-compose.prod.yml logs -f

# Bot only
docker-compose -f docker-compose.prod.yml logs -f bot

# Last 100 lines
docker-compose -f docker-compose.prod.yml logs --tail 100 bot
```

Look for:
```json
{"level":"INFO","msg":"discord connected","user":"The Chronicler"}
{"level":"INFO","msg":"registered global command","name":"ask"}
```

## Common Operations

### Restart Bot (After Code Changes)

```bash
# Rebuild and restart
docker-compose -f docker-compose.prod.yml build bot
docker-compose -f docker-compose.prod.yml restart bot
```

### Update Documentation

```bash
# Transfer new training data
rsync -avz /path/to/training-data/ \
  user@proxmox.local:/opt/living-lands-bot/training-data/

# Reindex
docker-compose -f docker-compose.prod.yml exec bot \
  ./bot index-docs --path /app/training-data
```

### Backup Database

```bash
# Create backup
docker-compose -f docker-compose.prod.yml exec postgres \
  pg_dump -U bot livinglands > backup_$(date +%Y%m%d).sql

# Or use the backup script
docker-compose -f docker-compose.prod.yml exec bot ./bot backup
```

### Update to Latest Version

```bash
# Pull latest code (on local machine)
cd ~/Development/living-lands-bot
git pull

# Redeploy
./deploy-to-proxmox.sh user@proxmox.local
```

## Monitoring

### Resource Usage

```bash
# Check container stats
docker stats

# Check disk usage
docker system df
```

### Logs

View real-time logs:
```bash
docker-compose -f docker-compose.prod.yml logs -f bot
```

Filter for errors:
```bash
docker-compose -f docker-compose.prod.yml logs bot | grep ERROR
```

### Metrics

The bot logs structured JSON with metrics:
- `llm_response_generated` - LLM generation time, tokens/sec
- `rag_query_complete` - RAG retrieval results
- `ask_command_completed` - End-to-end request duration

## Troubleshooting

### Bot Not Connecting to Discord

1. Check Discord token:
   ```bash
   docker-compose -f docker-compose.prod.yml exec bot env | grep DISCORD_TOKEN
   ```

2. Verify network connectivity:
   ```bash
   docker-compose -f docker-compose.prod.yml exec bot ping -c 3 discord.com
   ```

3. Check logs for connection errors:
   ```bash
   docker-compose -f docker-compose.prod.yml logs bot | grep -i "discord\|error"
   ```

### LLM Timeouts

If seeing frequent timeouts:

1. Check Ollama memory usage:
   ```bash
   docker stats living-lands-bot-ollama-1
   ```

2. Increase timeout (already set to 90s in latest version)

3. Consider using smaller model:
   ```bash
   # In .env, change to:
   OLLAMA_MODEL=mistral:7b-instruct-q4_0
   ```

### ChromaDB Empty (No RAG Results)

Reindex documents:
```bash
docker-compose -f docker-compose.prod.yml exec bot \
  ./bot index-docs --path /app/training-data
```

Check logs:
```bash
docker-compose -f docker-compose.prod.yml logs bot | grep "document indexing"
```

### Database Connection Errors

1. Check PostgreSQL health:
   ```bash
   docker-compose -f docker-compose.prod.yml exec postgres \
     pg_isready -U bot -d livinglands
   ```

2. Verify password in `.env` matches

3. Restart PostgreSQL:
   ```bash
   docker-compose -f docker-compose.prod.yml restart postgres
   ```

### Out of Disk Space

Clean up Docker resources:
```bash
# Remove unused images
docker image prune -a

# Remove unused volumes (CAUTION: only if you have backups!)
docker volume prune

# Check space
docker system df
```

## Security Recommendations

1. **Firewall Rules:**
   ```bash
   # Only allow SSH and bot HTTP API
   sudo ufw allow 22/tcp
   sudo ufw allow 8000/tcp
   sudo ufw enable
   ```

2. **Use Strong Passwords:**
   - Generate DB password: `openssl rand -base64 32`
   - Generate Chroma token: `openssl rand -hex 32`

3. **Regular Backups:**
   ```bash
   # Add to crontab
   0 2 * * * cd /opt/living-lands-bot && docker-compose exec bot ./bot backup
   ```

4. **Keep Updated:**
   ```bash
   # Update Docker images monthly
   docker-compose -f docker-compose.prod.yml pull
   docker-compose -f docker-compose.prod.yml up -d
   ```

## Performance Tuning

### For Low-Resource Servers (4-8GB RAM)

Edit `docker-compose.prod.yml`:

```yaml
redis:
  command: redis-server --maxmemory 128mb --maxmemory-policy allkeys-lru

ollama:
  environment:
    OLLAMA_NUM_PARALLEL: 1
    OLLAMA_MAX_LOADED_MODELS: 1
```

Use quantized model:
```env
OLLAMA_MODEL=mistral:7b-instruct-q4_0
```

### For High-Performance Servers (16GB+ RAM)

Enable parallel requests:
```yaml
ollama:
  environment:
    OLLAMA_NUM_PARALLEL: 4
    OLLAMA_MAX_LOADED_MODELS: 2
```

Increase Redis memory:
```yaml
redis:
  command: redis-server --maxmemory 512mb
```

## Uninstall

```bash
# Stop all services
docker-compose -f docker-compose.prod.yml down

# Remove volumes (WARNING: deletes all data!)
docker-compose -f docker-compose.prod.yml down -v

# Remove deployment directory
sudo rm -rf /opt/living-lands-bot
```

## Support

For issues:
1. Check logs: `docker-compose -f docker-compose.prod.yml logs -f`
2. Verify health: `curl http://localhost:8000/health`
3. Review Linear issues: [LLB Project](https://linear.app/moshpitcodes)
4. Check documentation: `docs/TECHNICAL_DESIGN.md`

---

**Version:** 0.1.0  
**Last Updated:** 2026-02-01  
**Author:** Living Lands Discord Bot Team
