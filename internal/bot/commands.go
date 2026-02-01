package bot

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/bwmarrin/discordgo"

	"living-lands-bot/internal/services"
)

type CommandHandlers struct {
	account *services.AccountService
	rag     *services.RAGService
	llm     *services.LLMService
	limiter *services.RateLimiter
	logger  *slog.Logger
}

func NewCommandHandlers(account *services.AccountService, rag *services.RAGService, llm *services.LLMService, limiter *services.RateLimiter, logger *slog.Logger) *CommandHandlers {
	return &CommandHandlers{
		account: account,
		rag:     rag,
		llm:     llm,
		limiter: limiter,
		logger:  logger,
	}
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "link",
		Description: "Link your Hytale account to Discord",
	},
	{
		Name:        "guide",
		Description: "Get directions to important channels",
	},
	{
		Name:        "ask",
		Description: "Ask a question about Living Lands",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "question",
				Description: "Your question about the mod",
				Required:    true,
			},
		},
	},
}

func (h *CommandHandlers) RegisterCommands(s *discordgo.Session, guildID string) error {
	h.logger.Info("registering commands", "guild_id", guildID, "bot_id", s.State.User.ID)

	// Try guild commands first (instant), fallback to global if guild fails
	for _, cmd := range commands {
		created, err := s.ApplicationCommandCreate(s.State.User.ID, guildID, cmd)
		if err != nil {
			h.logger.Error("guild command failed, trying global",
				"command", cmd.Name,
				"error", err,
				"guild_id", guildID,
			)

			// Fallback: try global command (takes up to 1hr to propagate)
			created, err = s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
			if err != nil {
				return fmt.Errorf("failed to create command %s (guild and global): %w", cmd.Name, err)
			}
			h.logger.Info("registered global command", "name", cmd.Name, "id", created.ID)
		} else {
			h.logger.Info("registered guild command", "name", cmd.Name, "id", created.ID, "guild", guildID)
		}
	}
	return nil
}

func (h *CommandHandlers) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		h.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		h.handleComponent(s, i)
	}
}

func (h *CommandHandlers) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	switch data.Name {
	case "link":
		h.handleLinkCommand(s, i)
	case "guide":
		h.handleGuideCommand(s, i)
	case "ask":
		h.handleAskCommand(s, i)
	}
}

func (h *CommandHandlers) handleLinkCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var discordID string
	var username string

	// Safely extract user information with nil checks
	if i.Member != nil && i.Member.User != nil {
		// Discord IDs are strings in discordgo v0.29+, use directly
		discordID = i.Member.User.ID
		username = i.Member.User.Username
	} else if i.User != nil {
		discordID = i.User.ID
		username = i.User.Username
	}

	// Fail gracefully if we couldn't get user information
	if username == "" || discordID == "" {
		h.logger.Warn("failed to extract user information from interaction",
			"has_member", i.Member != nil,
			"has_user", i.User != nil,
		)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Unable to identify your Discord account. Please try again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	code, err := h.account.GenerateVerificationCode(discordID, username)
	if err != nil {
		h.logger.Error("failed to generate code", "error", err)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Failed to generate verification code. Please try again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Your verification code is: `%s`\n\n"+
				"Run `/verify %s` in Hytale to link your account.\n"+
				"This code expires in 10 minutes.", code, code),
			Flags: discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *CommandHandlers) handleGuideCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Build embed with channel guide
	embed := &discordgo.MessageEmbed{
		Title:       "📍 Channel Guide",
		Description: "Select a category below to navigate to the right channel:",
		Color:       0x2D6A4F, // Forest green from brand palette
	}

	// Build action row with buttons
	components := []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "🐛 Bug Reports",
					Style:    discordgo.PrimaryButton,
					CustomID: "guide_bugs",
				},
				discordgo.Button{
					Label:    "📋 Changelog",
					Style:    discordgo.PrimaryButton,
					CustomID: "guide_changelog",
				},
				discordgo.Button{
					Label:    "📚 Wiki",
					Style:    discordgo.PrimaryButton,
					CustomID: "guide_wiki",
				},
				discordgo.Button{
					Label:    "💬 Support",
					Style:    discordgo.PrimaryButton,
					CustomID: "guide_support",
				},
			},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			Flags:      discordgo.MessageFlagsEphemeral,
		},
	})
}

func (h *CommandHandlers) handleAskCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	startTime := time.Now()

	// Get user ID for rate limiting
	var userID string
	var username string
	if i.Member != nil && i.Member.User != nil {
		userID = i.Member.User.ID
		username = i.Member.User.Username
	} else if i.User != nil {
		userID = i.User.ID
		username = i.User.Username
	}

	// Check rate limit before processing
	if userID != "" && h.limiter != nil {
		allowed, remaining, retryAfter, err := h.limiter.IsAllowed(context.Background(), userID)
		if err != nil {
			h.logger.Error("rate limit check failed", "error", err, "user_id", userID)
			// Log but continue (don't block on rate limit failure)
		} else if !allowed {
			h.logger.Warn("rate limit exceeded", "user_id", userID, "username", username, "retry_after", retryAfter.Seconds())
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("The archives are experiencing many seekers at once. Please try again in %.0f seconds.", retryAfter.Seconds()),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
			return
		}
		h.logger.Debug("rate limit allowed", "user_id", userID, "remaining", remaining)
	}

	// Get question from command options
	data := i.ApplicationCommandData()
	if len(data.Options) == 0 {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Please provide a question.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	question := data.Options[0].StringValue()

	// Classify the query intent
	intent := services.ClassifyIntent(question)
	h.logger.Debug("query intent classified", "question", question, "intent", intent.String())

	// Handle navigation and account intents with shortcuts (no LLM needed)
	switch intent {
	case services.IntentNavigation:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "For channel navigation, use the `/guide` command - it will help you find the right place!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	case services.IntentAccountHelp:
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "For account linking, use the `/link` command - it will generate a verification code for you!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	// Defer the interaction response (RAG+LLM takes >3 seconds)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Determine response mode based on intent
	mode := services.DetermineMode(intent, false) // will be updated after RAG query

	// Set timeout based on mode - faster modes get shorter timeouts
	// Discord allows 15 minutes for follow-ups, but we want fast responses
	// Keep buffer of ~5s for final response send
	var timeout time.Duration
	switch mode {
	case services.ModeFast:
		timeout = 30 * time.Second
	case services.ModeStandard:
		timeout = 60 * time.Second
	case services.ModeDeep:
		timeout = 90 * time.Second // Increased for RAG-heavy queries
	default:
		timeout = 60 * time.Second
	}

	// Create root context with total timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 1. Only query RAG if the intent requires it
	var ragContext []string
	if intent.NeedsRAG() {
		// Use a sub-context with shorter timeout for RAG
		// Ensure RAG timeout doesn't exceed parent context timeout
		ragTimeout := 5 * time.Second
		if timeout < ragTimeout {
			// If parent timeout is shorter, use 80% of parent timeout for RAG
			ragTimeout = time.Duration(float64(timeout) * 0.8)
		}
		ragCtx, ragCancel := context.WithTimeout(ctx, ragTimeout)
		defer ragCancel()

		var err error
		ragContext, err = h.rag.Query(ragCtx, question, 5)
		if err != nil {
			ragTimeoutReached := ragCtx.Err() == context.DeadlineExceeded
			h.logger.Warn("rag query failed, continuing without context",
				"error", err,
				"question", question,
				"timeout_reached", ragTimeoutReached,
				"rag_timeout_ms", ragTimeout.Milliseconds(),
			)
			// Continue without context if RAG fails
			ragContext = []string{}
		} else {
			h.logger.Debug("rag context retrieved", "count", len(ragContext), "intent", intent.String())
		}
	} else {
		h.logger.Debug("skipping rag for conversational query", "question", question)
	}

	// Update mode now that we know if we have RAG context
	mode = services.DetermineMode(intent, len(ragContext) > 0)

	// 2. Generate LLM response with intent-aware mode
	answer, err := h.llm.GenerateResponseWithIntent(ctx, question, ragContext, intent)
	if err != nil {
		h.logger.Error("llm generation failed",
			"error", err,
			"question", question,
			"intent", intent.String(),
			"mode", mode.String(),
			"elapsed_ms", time.Since(startTime).Milliseconds(),
			"timeout_reached", ctx.Err() != nil,
		)
		// Provide a graceful fallback message
		if ctx.Err() != nil {
			answer = "The archives are being consulted by many travelers at this moment, causing some delay. Please try again shortly, seeker."
		} else {
			answer = "I apologize, traveler. The mists cloud my vision at this moment. Please try again."
		}
	}

	// 3. Send follow-up response
	_, sendErr := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: answer,
	})

	// username already obtained from rate limit check above, no need to redeclare

	elapsedMs := time.Since(startTime).Milliseconds()

	if sendErr != nil {
		h.logger.Error("failed to send followup",
			"error", sendErr,
			"username", username,
			"elapsed_ms", elapsedMs,
		)
	} else {
		h.logger.Info("ask command completed",
			"user", username,
			"question", question,
			"intent", intent.String(),
			"mode", mode.String(),
			"rag_contexts", len(ragContext),
			"elapsed_ms", elapsedMs,
			"success", err == nil,
		)
	}
}

func (h *CommandHandlers) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	// Handle guide button clicks
	if len(customID) > 6 && customID[:6] == "guide_" {
		keyword := customID[6:] // Extract keyword (e.g., "bugs" from "guide_bugs")

		// For now, just acknowledge the click
		// TODO: Look up channel_id from database and provide link
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("Navigation to **%s** coming soon! (Channel mapping needs configuration)", keyword),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	h.logger.Info("unknown component interaction", "custom_id", customID)
}
