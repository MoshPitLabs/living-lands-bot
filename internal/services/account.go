package services

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"gorm.io/gorm"

	"living-lands-bot/internal/database/models"
)

type AccountService struct {
	db     *gorm.DB
	expiry time.Duration
	logger *slog.Logger
}

func NewAccountService(db *gorm.DB, expirySeconds int, logger *slog.Logger) *AccountService {
	return &AccountService{
		db:     db,
		expiry: time.Duration(expirySeconds) * time.Second,
		logger: logger,
	}
}

// GenerateVerificationCode creates a new 8-char code for Discord user
// discordID should be the Discord user ID as a string (from discordgo)
func (s *AccountService) GenerateVerificationCode(discordID string, discordUsername string) (string, error) {
	if discordID == "" {
		return "", fmt.Errorf("discord_id cannot be empty")
	}

	code := generateCode(8)

	user := &models.User{
		DiscordID:        discordID,
		DiscordUsername:  discordUsername,
		VerificationCode: code,
	}

	err := s.db.Where("discord_id = ?", discordID).
		Assign(user).
		FirstOrCreate(user).Error

	if err != nil {
		return "", fmt.Errorf("failed to generate verification code for user %s: %w", discordID, err)
	}

	s.logger.Info("verification code generated", "discord_id", discordID, "code", code)
	return code, nil
}

// VerifyLink validates code from Hytale and links accounts
func (s *AccountService) VerifyLink(code, hytaleUsername, hytaleUUID string) error {
	var user models.User

	err := s.db.Where("verification_code = ?", code).First(&user).Error
	if err != nil {
		return fmt.Errorf("invalid verification code")
	}

	// Check expiry (code valid for configured duration)
	if time.Since(user.UpdatedAt) > s.expiry {
		return fmt.Errorf("verification code expired")
	}

	// Update with Hytale info
	now := time.Now()
	user.HytaleUsername = hytaleUsername
	user.HytaleUUID = hytaleUUID
	user.VerifiedAt = &now
	user.VerificationCode = "" // Clear code

	if err := s.db.Save(&user).Error; err != nil {
		return fmt.Errorf("failed to save verified user: %w", err)
	}

	s.logger.Info("account linked",
		"discord_id", user.DiscordID,
		"hytale_username", hytaleUsername,
		"hytale_uuid", hytaleUUID,
	)

	return nil
}

func generateCode(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	code := base32.StdEncoding.EncodeToString(b)[:length]
	return strings.ToUpper(code)
}
