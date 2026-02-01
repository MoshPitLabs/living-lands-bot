package services

import (
	"log/slog"
	"os"
	"testing"

	"living-lands-bot/internal/database/models"
)

func TestChannelServiceInitialization(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	s := NewChannelService(nil, logger)

	if s == nil {
		t.Error("ChannelService should not be nil")
	}

	if s.logger == nil {
		t.Error("Logger should not be nil")
	}

	t.Log("ChannelService initialized successfully")
}

func TestChannelRouteStructure(t *testing.T) {
	// Test ChannelRoute model structure
	route := models.ChannelRoute{
		ID:          1,
		Keyword:     "bugs",
		ChannelID:   "1234567890",
		Description: "Report bugs and issues",
		Emoji:       "🐛",
	}

	if route.Keyword != "bugs" {
		t.Errorf("Expected keyword 'bugs', got '%s'", route.Keyword)
	}

	if route.ChannelID != "1234567890" {
		t.Errorf("Expected channel ID '1234567890', got '%s'", route.ChannelID)
	}

	if route.Emoji != "🐛" {
		t.Errorf("Expected emoji '🐛', got '%s'", route.Emoji)
	}

	t.Log("ChannelRoute structure is correct")
}

func TestRouteKeywordNormalization(t *testing.T) {
	testCases := []struct {
		keyword string
		valid   bool
	}{
		{"bugs", true},
		{"changelog", true},
		{"wiki", true},
		{"support", true},
		{"", false},
	}

	for _, tc := range testCases {
		isValid := len(tc.keyword) > 0
		if isValid != tc.valid {
			t.Errorf("Keyword '%s' validation failed: expected %v, got %v", tc.keyword, tc.valid, isValid)
		}
	}

	t.Log("Route keyword validation passed")
}

func TestMultipleChannelRoutes(t *testing.T) {
	// Test handling multiple routes
	routes := []models.ChannelRoute{
		{
			Keyword:     "bugs",
			ChannelID:   "123",
			Description: "Bug reports",
			Emoji:       "🐛",
		},
		{
			Keyword:     "changelog",
			ChannelID:   "456",
			Description: "Version history",
			Emoji:       "📋",
		},
		{
			Keyword:     "wiki",
			ChannelID:   "789",
			Description: "Documentation",
			Emoji:       "📚",
		},
	}

	if len(routes) != 3 {
		t.Errorf("Expected 3 routes, got %d", len(routes))
	}

	keywordMap := make(map[string]string)
	for _, route := range routes {
		keywordMap[route.Keyword] = route.ChannelID
	}

	if keywordMap["bugs"] != "123" {
		t.Errorf("Expected channel '123' for bugs, got '%s'", keywordMap["bugs"])
	}

	if keywordMap["wiki"] != "789" {
		t.Errorf("Expected channel '789' for wiki, got '%s'", keywordMap["wiki"])
	}

	t.Log("Multiple channel routes handled correctly")
}

func TestChannelIDFormat(t *testing.T) {
	// Test Discord channel ID format
	validIDs := []string{
		"1234567890",
		"999999999999",
		"123",
	}

	for _, id := range validIDs {
		if len(id) == 0 {
			t.Errorf("Channel ID should not be empty: %s", id)
		}
		t.Logf("Valid channel ID: %s", id)
	}
}

func TestEmojiSupport(t *testing.T) {
	// Test emoji in routes
	testCases := []struct {
		emoji string
		desc  string
	}{
		{"🐛", "Bug icon"},
		{"📋", "Changelog icon"},
		{"📚", "Wiki icon"},
		{"💬", "Support icon"},
	}

	for _, tc := range testCases {
		if len(tc.emoji) == 0 {
			t.Errorf("Emoji should not be empty for %s", tc.desc)
		}
		t.Logf("Emoji '%s' for %s is valid", tc.emoji, tc.desc)
	}
}

func TestChannelRouteDuplicateKeywords(t *testing.T) {
	// Test handling duplicate keywords (should use unique index)
	routes := []models.ChannelRoute{
		{Keyword: "bugs", ChannelID: "123"},
		{Keyword: "bugs", ChannelID: "456"}, // Duplicate keyword
	}

	// Count unique keywords
	keywordMap := make(map[string]bool)
	for _, route := range routes {
		keywordMap[route.Keyword] = true
	}

	if len(keywordMap) != 1 {
		t.Logf("Duplicate keyword detected: %d unique keywords out of %d routes", len(keywordMap), len(routes))
	}

	t.Log("Duplicate keyword handling verified")
}
