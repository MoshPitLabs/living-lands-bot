#!/bin/bash
# Script to clean up duplicate Discord commands

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Discord Command Cleanup ===${NC}"
echo ""

# Check if we're running locally or on remote
if [ -f "$PROJECT_ROOT/.env" ]; then
	echo -e "${YELLOW}Running locally...${NC}"
	source "$PROJECT_ROOT/.env"
	DOCKER_CMD="docker exec living-lands-bot-bot-1"
else
	echo -e "${RED}Error: .env file not found${NC}"
	exit 1
fi

# Check if bot container is running
if ! docker ps | grep -q "living-lands-bot-bot-1"; then
	echo -e "${RED}Error: Bot container is not running${NC}"
	exit 1
fi

echo -e "${YELLOW}This will delete ALL registered commands (global and guild) and re-register them.${NC}"
echo -e "${YELLOW}This fixes duplicate commands in Discord.${NC}"
echo ""
read -p "Continue? (y/n): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
	echo "Cancelled."
	exit 0
fi

echo ""
echo -e "${GREEN}Step 1: Stopping bot...${NC}"
docker compose stop bot

echo ""
echo -e "${GREEN}Step 2: Creating cleanup script...${NC}"
cat >/tmp/cleanup-commands.go <<'EOF'
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := os.Getenv("DISCORD_TOKEN")
	guildID := os.Getenv("DISCORD_GUILD_ID")

	if token == "" {
		log.Fatal("DISCORD_TOKEN is required")
	}

	// Create session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	if err := dg.Open(); err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer dg.Close()

	fmt.Println("Connected to Discord")
	fmt.Printf("Bot User: %s#%s\n\n", dg.State.User.Username, dg.State.User.Discriminator)

	// Delete global commands
	fmt.Println("Fetching global commands...")
	globalCommands, err := dg.ApplicationCommands(dg.State.User.ID, "")
	if err != nil {
		log.Fatal("Error fetching global commands:", err)
	}

	if len(globalCommands) == 0 {
		fmt.Println("No global commands found.")
	} else {
		fmt.Printf("Found %d global command(s):\n", len(globalCommands))
		for _, cmd := range globalCommands {
			fmt.Printf("  - %s (ID: %s)\n", cmd.Name, cmd.ID)
			if err := dg.ApplicationCommandDelete(dg.State.User.ID, "", cmd.ID); err != nil {
				fmt.Printf("    ERROR deleting: %v\n", err)
			} else {
				fmt.Printf("    ✓ Deleted\n")
			}
		}
	}

	// Delete guild commands if guild ID provided
	if guildID != "" {
		fmt.Printf("\nFetching guild commands for %s...\n", guildID)
		guildCommands, err := dg.ApplicationCommands(dg.State.User.ID, guildID)
		if err != nil {
			log.Fatal("Error fetching guild commands:", err)
		}

		if len(guildCommands) == 0 {
			fmt.Println("No guild commands found.")
		} else {
			fmt.Printf("Found %d guild command(s):\n", len(guildCommands))
			for _, cmd := range guildCommands {
				fmt.Printf("  - %s (ID: %s)\n", cmd.Name, cmd.ID)
				if err := dg.ApplicationCommandDelete(dg.State.User.ID, guildID, cmd.ID); err != nil {
					fmt.Printf("    ERROR deleting: %v\n", err)
				} else {
					fmt.Printf("    ✓ Deleted\n")
				}
			}
		}
	}

	fmt.Println("\n✓ Cleanup complete!")
	fmt.Println("Commands will be re-registered when the bot starts.")
}
EOF

echo ""
echo -e "${GREEN}Step 3: Running cleanup inside container...${NC}"
docker compose run --rm -e DISCORD_TOKEN="$DISCORD_TOKEN" -e DISCORD_GUILD_ID="$DISCORD_GUILD_ID" bot sh -c "
cat > /tmp/cleanup.go << 'INNEREOF'
$(cat /tmp/cleanup-commands.go)
INNEREOF
cd /tmp && go mod init cleanup && go get github.com/bwmarrin/discordgo && go run cleanup.go
"

echo ""
echo -e "${GREEN}Step 4: Starting bot...${NC}"
docker compose start bot

echo ""
echo -e "${GREEN}✓ Done! Commands have been cleaned up and will be re-registered.${NC}"
echo -e "${YELLOW}Note: Global command deletions can take up to 1 hour to propagate.${NC}"
echo ""

# Cleanup
rm -f /tmp/cleanup-commands.go
