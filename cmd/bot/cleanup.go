package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
)

// CleanupCommands removes all registered Discord commands (global and guild)
// This is useful for fixing duplicate commands
func CleanupCommands(token, guildID string) error {
	// Create session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return fmt.Errorf("error creating Discord session: %w", err)
	}

	if err := dg.Open(); err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}
	defer dg.Close()

	fmt.Printf("Connected to Discord as %s\n\n", dg.State.User.Username)

	// Delete global commands
	fmt.Println("Fetching global commands...")
	globalCommands, err := dg.ApplicationCommands(dg.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("error fetching global commands: %w", err)
	}

	if len(globalCommands) == 0 {
		fmt.Println("✓ No global commands found.")
	} else {
		fmt.Printf("Found %d global command(s):\n", len(globalCommands))
		for _, cmd := range globalCommands {
			fmt.Printf("  - %s (ID: %s) ... ", cmd.Name, cmd.ID)
			if err := dg.ApplicationCommandDelete(dg.State.User.ID, "", cmd.ID); err != nil {
				fmt.Printf("ERROR: %v\n", err)
			} else {
				fmt.Printf("✓ Deleted\n")
			}
		}
	}

	// Delete guild commands if guild ID provided
	if guildID != "" {
		fmt.Printf("\nFetching guild commands for guild %s...\n", guildID)
		guildCommands, err := dg.ApplicationCommands(dg.State.User.ID, guildID)
		if err != nil {
			return fmt.Errorf("error fetching guild commands: %w", err)
		}

		if len(guildCommands) == 0 {
			fmt.Println("✓ No guild commands found.")
		} else {
			fmt.Printf("Found %d guild command(s):\n", len(guildCommands))
			for _, cmd := range guildCommands {
				fmt.Printf("  - %s (ID: %s) ... ", cmd.Name, cmd.ID)
				if err := dg.ApplicationCommandDelete(dg.State.User.ID, guildID, cmd.ID); err != nil {
					fmt.Printf("ERROR: %v\n", err)
				} else {
					fmt.Printf("✓ Deleted\n")
				}
			}
		}
	}

	fmt.Println("\n✓ Cleanup complete!")
	fmt.Println("Commands will be re-registered when the bot starts normally.")
	return nil
}

func handleCleanup() {
	token := os.Getenv("DISCORD_TOKEN")
	guildID := os.Getenv("DISCORD_GUILD_ID")

	if token == "" {
		log.Fatal("DISCORD_TOKEN environment variable is required")
	}

	fmt.Println("=== Discord Command Cleanup ===")
	fmt.Println("This will delete ALL registered commands (global and guild).")
	fmt.Println("Commands will be re-registered when you start the bot normally.")
	fmt.Println()

	if err := CleanupCommands(token, guildID); err != nil {
		log.Fatal(err)
	}
}
