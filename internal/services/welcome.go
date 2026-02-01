package services

import (
	"fmt"
	"log/slog"
	"math/rand"
	"strings"

	"gorm.io/gorm"

	"living-lands-bot/internal/database/models"
)

type WelcomeService struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewWelcomeService(db *gorm.DB, logger *slog.Logger) *WelcomeService {
	return &WelcomeService{
		db:     db,
		logger: logger,
	}
}

// GetRandomTemplate returns a weighted random welcome message
func (s *WelcomeService) GetRandomTemplate(username string) (string, error) {
	var templates []models.WelcomeTemplate

	err := s.db.Where("active = ?", true).Find(&templates).Error
	if err != nil {
		return "", fmt.Errorf("failed to fetch templates: %w", err)
	}

	if len(templates) == 0 {
		// Fallback if no templates exist
		return fmt.Sprintf("Welcome, %s!", username), nil
	}

	// Calculate total weight
	totalWeight := 0
	for _, t := range templates {
		totalWeight += t.Weight
	}

	// Weighted random selection
	r := rand.Intn(totalWeight)
	cumWeight := 0

	for _, t := range templates {
		cumWeight += t.Weight
		if r < cumWeight {
			// Replace {username} placeholder
			message := strings.ReplaceAll(t.Message, "{username}", username)
			return message, nil
		}
	}

	// Fallback (shouldn't reach here)
	return fmt.Sprintf("Welcome, %s!", username), nil
}
