package services

import (
	"log/slog"

	"gorm.io/gorm"

	"living-lands-bot/internal/database/models"
)

type ChannelService struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewChannelService(db *gorm.DB, logger *slog.Logger) *ChannelService {
	return &ChannelService{
		db:     db,
		logger: logger,
	}
}

func (s *ChannelService) GetAllRoutes() ([]models.ChannelRoute, error) {
	var routes []models.ChannelRoute
	if err := s.db.Find(&routes).Error; err != nil {
		return nil, err
	}
	return routes, nil
}

func (s *ChannelService) GetRouteByKeyword(keyword string) (*models.ChannelRoute, error) {
	var route models.ChannelRoute
	if err := s.db.Where("keyword = ?", keyword).First(&route).Error; err != nil {
		return nil, err
	}
	return &route, nil
}
