#!/bin/bash
set -e

# Living Lands Discord Bot - Proxmox Deployment Script
# Usage: ./deploy-to-proxmox.sh <proxmox-host> [deploy-path]

PROXMOX_HOST="${1}"
DEPLOY_PATH="${2:-/opt/living-lands-bot}"
LOCAL_TRAINING_DATA="/mnt/ugreen-nas/Coding/Hytale/DiscordBot/training-data"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Living Lands Bot - Proxmox Deployment ===${NC}"
echo ""

# Validate arguments
if [ -z "$PROXMOX_HOST" ]; then
	echo -e "${RED}Error: Proxmox host not specified${NC}"
	echo "Usage: $0 <proxmox-host> [deploy-path]"
	echo "Example: $0 user@proxmox.local /opt/living-lands-bot"
	exit 1
fi

echo -e "${YELLOW}Configuration:${NC}"
echo "  Proxmox Host: $PROXMOX_HOST"
echo "  Deploy Path: $DEPLOY_PATH"
echo ""

# Check if .env file exists
if [ ! -f .env ]; then
	echo -e "${RED}Error: .env file not found!${NC}"
	echo "Please create .env file with required configuration."
	exit 1
fi

# Check if training data exists
if [ ! -d "$LOCAL_TRAINING_DATA" ]; then
	echo -e "${YELLOW}Warning: Training data not found at $LOCAL_TRAINING_DATA${NC}"
	echo "Training data will need to be transferred separately."
fi

echo -e "${GREEN}Step 1: Creating deployment directory on Proxmox...${NC}"
ssh "$PROXMOX_HOST" "mkdir -p $DEPLOY_PATH/{configs,training-data,backups}"

echo -e "${GREEN}Step 2: Copying project files...${NC}"
rsync -avz --progress \
	--exclude '.git' \
	--exclude '.env.example' \
	--exclude 'node_modules' \
	--exclude 'backups' \
	--exclude 'training-data' \
	./ "$PROXMOX_HOST:$DEPLOY_PATH/"

echo -e "${GREEN}Step 3: Copying environment file...${NC}"
scp .env "$PROXMOX_HOST:$DEPLOY_PATH/.env"

echo -e "${GREEN}Step 4: Copying configs...${NC}"
rsync -avz --progress configs/ "$PROXMOX_HOST:$DEPLOY_PATH/configs/"

echo -e "${GREEN}Step 5: Transferring training data (this may take a while)...${NC}"
if [ -d "$LOCAL_TRAINING_DATA" ]; then
	rsync -avz --progress "$LOCAL_TRAINING_DATA/" "$PROXMOX_HOST:$DEPLOY_PATH/training-data/"
	echo -e "${GREEN}Training data transferred successfully!${NC}"
else
	echo -e "${YELLOW}Skipping training data transfer (directory not found)${NC}"
	echo "You'll need to manually transfer training data to: $PROXMOX_HOST:$DEPLOY_PATH/training-data/"
fi

echo -e "${GREEN}Step 6: Building Docker images on Proxmox...${NC}"
ssh "$PROXMOX_HOST" "cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml build"

echo -e "${GREEN}Step 7: Pulling required Docker images...${NC}"
ssh "$PROXMOX_HOST" "cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml pull"

echo -e "${GREEN}Step 8: Starting services...${NC}"
ssh "$PROXMOX_HOST" "cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml up -d"

echo -e "${GREEN}Step 9: Downloading Ollama models...${NC}"
echo "Downloading Mistral 7B (this will take several minutes)..."
ssh "$PROXMOX_HOST" "cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml exec -T ollama ollama pull mistral:7b-instruct"
echo "Downloading embedding model..."
ssh "$PROXMOX_HOST" "cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml exec -T ollama ollama pull nomic-embed-text"

echo -e "${GREEN}Step 10: Running database migrations...${NC}"
ssh "$PROXMOX_HOST" "cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml exec -T bot ./bot migrate"

echo -e "${GREEN}Step 11: Indexing documentation...${NC}"
ssh "$PROXMOX_HOST" "cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml exec -T bot ./bot index-docs --path /app/training-data"

echo ""
echo -e "${GREEN}=== Deployment Complete! ===${NC}"
echo ""
echo "Services are now running on $PROXMOX_HOST"
echo ""
echo "Useful commands:"
echo "  View logs:    ssh $PROXMOX_HOST 'cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml logs -f bot'"
echo "  Stop:         ssh $PROXMOX_HOST 'cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml stop'"
echo "  Restart:      ssh $PROXMOX_HOST 'cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml restart'"
echo "  Status:       ssh $PROXMOX_HOST 'cd $DEPLOY_PATH && docker-compose -f docker-compose.prod.yml ps'"
echo ""
echo "Health check URL: http://$PROXMOX_HOST:8000/health"
echo ""
