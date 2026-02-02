#!/bin/bash
set -e

# Living Lands Discord Bot - Remote Management Script
# Usage: ./remote-manage.sh <command> [remote-host]

COMMAND="${1}"
REMOTE_HOST="${2:-ubuntu@192.168.178.50}"
DEPLOY_PATH="/opt/living-lands-bot"
LOCAL_TRAINING_DATA="/mnt/ugreen-nas/Coding/Hytale/DiscordBot/training-data"
LOCAL_LIVINGLANDS_DOCS="/home/moshpitcodes/Development/living-lands-reloaded/docs"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print usage
usage() {
	echo "Living Lands Bot - Remote Management"
	echo ""
	echo "Usage: $0 <command> [remote-host]"
	echo ""
	echo "Commands:"
	echo "  update          - Sync code, rebuild, and restart bot (excludes data)"
	echo "  update-full     - Full update: code + training data + docs + rebuild"
	echo "  rebuild         - Rebuild Docker images and restart"
	echo "  restart         - Restart all services"
	echo "  stop            - Stop all services"
	echo "  start           - Start all services"
	echo "  logs            - Show bot logs (follow mode)"
	echo "  status          - Show service status"
	echo "  health          - Check health endpoint"
	echo "  backup-db       - Backup PostgreSQL database"
	echo "  shell           - Open shell in bot container"
	echo "  migrate         - Run database migrations"
	echo "  index-docs      - Reindex documentation"
	echo "  sync-training   - Sync training data from local NAS to remote"
	echo "  sync-docs       - Sync living-lands-reloaded docs to remote"
	echo "  sync-all        - Sync both training data and docs"
	echo ""
	echo "Default remote host: $REMOTE_HOST"
	echo "Deploy path: $DEPLOY_PATH"
	echo ""
	echo "Examples:"
	echo "  $0 update 192.168.178.50"
	echo "  $0 logs"
	echo "  $0 status"
}

# Check if command is provided
if [ -z "$COMMAND" ]; then
	usage
	exit 1
fi

echo -e "${BLUE}=== Living Lands Bot - Remote Management ===${NC}"
echo -e "${YELLOW}Remote Host:${NC} $REMOTE_HOST"
echo -e "${YELLOW}Deploy Path:${NC} $DEPLOY_PATH"
echo ""

# Execute commands
case "$COMMAND" in
update)
	echo -e "${GREEN}Updating bot on remote host...${NC}"
	echo ""

	echo -e "${YELLOW}Step 1/7:${NC} Syncing code to remote..."
	rsync -avz --progress --delete \
		--exclude '.git' \
		--exclude '.env' \
		--exclude 'training-data' \
		--exclude 'livinglands-docs' \
		--exclude 'backups' \
		./ "$REMOTE_HOST:$DEPLOY_PATH/"

	echo -e "${YELLOW}Step 2/7:${NC} Syncing configs..."
	rsync -avz --progress ./configs/ "$REMOTE_HOST:$DEPLOY_PATH/configs/"

	echo -e "${YELLOW}Step 3/7:${NC} Rebuilding Docker images..."
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose build --no-cache bot"

	echo -e "${YELLOW}Step 4/7:${NC} Stopping bot service..."
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose stop bot"

	echo -e "${YELLOW}Step 5/7:${NC} Starting bot service..."
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose up -d bot"

	echo -e "${YELLOW}Step 6/7:${NC} Running migrations..."
	sleep 5
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose exec -T bot ./bot migrate" || echo "Migrations completed or already up to date"

	echo -e "${YELLOW}Step 7/7:${NC} Checking health..."
	sleep 3
	ssh "$REMOTE_HOST" "curl -f http://localhost:8000/health" || echo "Health check endpoint not responding yet"

	echo ""
	echo -e "${GREEN}✓ Update complete!${NC}"
	echo -e "${YELLOW}Note: Training data and docs not synced. Use 'sync-all' or 'update-full' if needed.${NC}"
	;;

update-full)
	echo -e "${GREEN}Full update: code, data, and rebuild...${NC}"
	echo ""

	echo -e "${YELLOW}Step 1/10:${NC} Syncing code to remote..."
	rsync -avz --progress --delete \
		--exclude '.git' \
		--exclude '.env' \
		--exclude 'training-data' \
		--exclude 'livinglands-docs' \
		--exclude 'backups' \
		./ "$REMOTE_HOST:$DEPLOY_PATH/"

	echo -e "${YELLOW}Step 2/10:${NC} Syncing configs..."
	rsync -avz --progress ./configs/ "$REMOTE_HOST:$DEPLOY_PATH/configs/"

	# Sync training data
	if [ -d "$LOCAL_TRAINING_DATA" ]; then
		echo -e "${YELLOW}Step 3/10:${NC} Syncing training data..."
		ssh "$REMOTE_HOST" "mkdir -p $DEPLOY_PATH/training-data"
		rsync -avz --progress --delete "$LOCAL_TRAINING_DATA/" "$REMOTE_HOST:$DEPLOY_PATH/training-data/"
		echo -e "${GREEN}✓ Training data synced${NC}"
	else
		echo -e "${YELLOW}Step 3/10: Skipping training data (not found at $LOCAL_TRAINING_DATA)${NC}"
	fi

	# Sync living lands docs
	if [ -d "$LOCAL_LIVINGLANDS_DOCS" ]; then
		echo -e "${YELLOW}Step 4/10:${NC} Syncing living-lands-reloaded docs..."
		ssh "$REMOTE_HOST" "mkdir -p $DEPLOY_PATH/livinglands-docs"
		rsync -avz --progress --delete "$LOCAL_LIVINGLANDS_DOCS/" "$REMOTE_HOST:$DEPLOY_PATH/livinglands-docs/"
		echo -e "${GREEN}✓ Living Lands docs synced${NC}"
	else
		echo -e "${YELLOW}Step 4/10: Skipping living lands docs (not found at $LOCAL_LIVINGLANDS_DOCS)${NC}"
	fi

	echo -e "${YELLOW}Step 5/10:${NC} Rebuilding Docker images..."
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose build --no-cache bot"

	echo -e "${YELLOW}Step 6/10:${NC} Stopping bot service..."
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose stop bot"

	echo -e "${YELLOW}Step 7/10:${NC} Starting bot service..."
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose up -d bot"

	echo -e "${YELLOW}Step 8/10:${NC} Running migrations..."
	sleep 5
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose exec -T bot ./bot migrate" || echo "Migrations completed or already up to date"

	echo -e "${YELLOW}Step 9/10:${NC} Reindexing documentation..."
	sleep 2
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose exec -T bot ./bot index-docs --path /app/training-data" || echo "Indexing completed"

	echo -e "${YELLOW}Step 10/10:${NC} Checking health..."
	sleep 3
	ssh "$REMOTE_HOST" "curl -f http://localhost:8000/health" || echo "Health check endpoint not responding yet"

	echo ""
	echo -e "${GREEN}✓ Full update complete!${NC}"
	;;

rebuild)
	echo -e "${GREEN}Rebuilding bot...${NC}"
	echo ""

	echo -e "${YELLOW}Step 1/3:${NC} Rebuilding Docker images..."
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose build --no-cache bot"

	echo -e "${YELLOW}Step 2/3:${NC} Restarting services..."
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose up -d bot"

	echo -e "${YELLOW}Step 3/3:${NC} Checking health..."
	sleep 3
	ssh "$REMOTE_HOST" "curl -f http://localhost:8000/health" || echo "Health check endpoint not responding yet"

	echo ""
	echo -e "${GREEN}✓ Rebuild complete!${NC}"
	;;

restart)
	echo -e "${GREEN}Restarting services...${NC}"
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose restart"
	echo -e "${GREEN}✓ Services restarted${NC}"
	;;

stop)
	echo -e "${GREEN}Stopping services...${NC}"
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose stop"
	echo -e "${GREEN}✓ Services stopped${NC}"
	;;

start)
	echo -e "${GREEN}Starting services...${NC}"
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose up -d"
	echo -e "${GREEN}✓ Services started${NC}"
	;;

logs)
	echo -e "${GREEN}Showing bot logs (Ctrl+C to exit)...${NC}"
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose logs -f bot"
	;;

status)
	echo -e "${GREEN}Service Status:${NC}"
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose ps"
	;;

health)
	echo -e "${GREEN}Checking health endpoint...${NC}"
	ssh "$REMOTE_HOST" "curl -s http://localhost:8000/health | jq '.' || curl -s http://localhost:8000/health"
	;;

backup-db)
	echo -e "${GREEN}Backing up database...${NC}"
	BACKUP_FILE="backup-$(date +%Y%m%d-%H%M%S).sql"
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose exec -T postgres pg_dump -U bot livinglands > backups/$BACKUP_FILE"
	echo -e "${GREEN}✓ Database backed up to: $DEPLOY_PATH/backups/$BACKUP_FILE${NC}"
	;;

shell)
	echo -e "${GREEN}Opening shell in bot container...${NC}"
	ssh -t "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose exec bot /bin/sh"
	;;

migrate)
	echo -e "${GREEN}Running database migrations...${NC}"
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose exec -T bot ./bot migrate"
	echo -e "${GREEN}✓ Migrations complete${NC}"
	;;

index-docs)
	echo -e "${GREEN}Reindexing documentation...${NC}"
	ssh "$REMOTE_HOST" "cd $DEPLOY_PATH && docker compose exec -T bot ./bot index-docs --path /app/training-data"
	echo -e "${GREEN}✓ Documentation indexed${NC}"
	;;

sync-training)
	echo -e "${GREEN}Syncing training data to remote...${NC}"

	if [ ! -d "$LOCAL_TRAINING_DATA" ]; then
		echo -e "${RED}Error: Local training data not found at $LOCAL_TRAINING_DATA${NC}"
		exit 1
	fi

	echo -e "${YELLOW}Source:${NC} $LOCAL_TRAINING_DATA"
	echo -e "${YELLOW}Target:${NC} $REMOTE_HOST:$DEPLOY_PATH/training-data/"
	echo ""

	ssh "$REMOTE_HOST" "mkdir -p $DEPLOY_PATH/training-data"
	rsync -avz --progress --delete "$LOCAL_TRAINING_DATA/" "$REMOTE_HOST:$DEPLOY_PATH/training-data/"

	echo ""
	echo -e "${GREEN}✓ Training data synced${NC}"
	;;

sync-docs)
	echo -e "${GREEN}Syncing living-lands-reloaded docs to remote...${NC}"

	if [ ! -d "$LOCAL_LIVINGLANDS_DOCS" ]; then
		echo -e "${RED}Error: Local docs not found at $LOCAL_LIVINGLANDS_DOCS${NC}"
		exit 1
	fi

	echo -e "${YELLOW}Source:${NC} $LOCAL_LIVINGLANDS_DOCS"
	echo -e "${YELLOW}Target:${NC} $REMOTE_HOST:$DEPLOY_PATH/livinglands-docs/"
	echo ""

	ssh "$REMOTE_HOST" "mkdir -p $DEPLOY_PATH/livinglands-docs"
	rsync -avz --progress --delete "$LOCAL_LIVINGLANDS_DOCS/" "$REMOTE_HOST:$DEPLOY_PATH/livinglands-docs/"

	echo ""
	echo -e "${GREEN}✓ Living Lands docs synced${NC}"
	;;

sync-all)
	echo -e "${GREEN}Syncing all data to remote...${NC}"
	echo ""

	# Sync training data
	if [ -d "$LOCAL_TRAINING_DATA" ]; then
		echo -e "${YELLOW}Step 1/2:${NC} Syncing training data..."
		ssh "$REMOTE_HOST" "mkdir -p $DEPLOY_PATH/training-data"
		rsync -avz --progress --delete "$LOCAL_TRAINING_DATA/" "$REMOTE_HOST:$DEPLOY_PATH/training-data/"
		echo -e "${GREEN}✓ Training data synced${NC}"
		echo ""
	else
		echo -e "${YELLOW}Skipping training data (not found at $LOCAL_TRAINING_DATA)${NC}"
		echo ""
	fi

	# Sync living lands docs
	if [ -d "$LOCAL_LIVINGLANDS_DOCS" ]; then
		echo -e "${YELLOW}Step 2/2:${NC} Syncing living-lands-reloaded docs..."
		ssh "$REMOTE_HOST" "mkdir -p $DEPLOY_PATH/livinglands-docs"
		rsync -avz --progress --delete "$LOCAL_LIVINGLANDS_DOCS/" "$REMOTE_HOST:$DEPLOY_PATH/livinglands-docs/"
		echo -e "${GREEN}✓ Living Lands docs synced${NC}"
		echo ""
	else
		echo -e "${YELLOW}Skipping living lands docs (not found at $LOCAL_LIVINGLANDS_DOCS)${NC}"
		echo ""
	fi

	echo -e "${GREEN}✓ All data synced${NC}"
	;;

*)
	echo -e "${RED}Error: Unknown command '$COMMAND'${NC}"
	echo ""
	usage
	exit 1
	;;
esac

echo ""
