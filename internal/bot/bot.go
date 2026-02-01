package bot

import (
	"log/slog"

	"github.com/bwmarrin/discordgo"

	"living-lands-bot/internal/config"
	"living-lands-bot/internal/services"
)

type Bot struct {
	session  *discordgo.Session
	config   *config.Config
	logger   *slog.Logger
	handlers *CommandHandlers
	welcome  *services.WelcomeService
	channel  *services.ChannelService
	limiter  *services.RateLimiter
}

func New(cfg *config.Config, account *services.AccountService, rag *services.RAGService, llm *services.LLMService, welcome *services.WelcomeService, channel *services.ChannelService, limiter *services.RateLimiter, logger *slog.Logger) (*Bot, error) {
	dg, err := discordgo.New("Bot " + cfg.Discord.Token)
	if err != nil {
		return nil, err
	}

	dg.Identify.Intents = discordgo.IntentsGuildMembers |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent

	handlers := NewCommandHandlers(account, rag, llm, limiter, logger)

	b := &Bot{
		session:  dg,
		config:   cfg,
		logger:   logger,
		handlers: handlers,
		welcome:  welcome,
		channel:  channel,
		limiter:  limiter,
	}

	dg.AddHandler(b.onReady)
	dg.AddHandler(handlers.HandleInteraction)
	dg.AddHandler(b.onGuildMemberAdd)

	return b, nil
}

func (b *Bot) Start() error {
	b.logger.Info("discord session opening")
	return b.session.Open()
}

func (b *Bot) Stop() error {
	b.logger.Info("discord session closing")
	return b.session.Close()
}

func (b *Bot) onReady(s *discordgo.Session, r *discordgo.Ready) {
	b.logger.Info("discord connected", "user", s.State.User.Username)

	// Register commands
	if err := b.handlers.RegisterCommands(s, b.config.Discord.GuildID); err != nil {
		b.logger.Error("failed to register commands", "error", err)
	}
}

func (b *Bot) onGuildMemberAdd(s *discordgo.Session, m *discordgo.GuildMemberAdd) {
	username := m.User.Username

	message, err := b.welcome.GetRandomTemplate(username)
	if err != nil {
		b.logger.Error("failed to get welcome template", "error", err, "user", username)
		return
	}

	// TODO: Make welcome channel configurable
	// For now, send to the system channel or first available text channel
	channels, err := s.GuildChannels(m.GuildID)
	if err != nil {
		b.logger.Error("failed to get channels", "error", err)
		return
	}

	// Find first text channel
	var targetChannel string
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildText {
			targetChannel = ch.ID
			break
		}
	}

	if targetChannel == "" {
		b.logger.Error("no text channel found for welcome message")
		return
	}

	_, err = s.ChannelMessageSend(targetChannel, message)
	if err != nil {
		b.logger.Error("failed to send welcome message", "error", err, "channel", targetChannel)
		return
	}

	b.logger.Info("welcome message sent", "user", username, "channel", targetChannel)
}
