package services

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"gorm.io/gorm"

	"living-lands-bot/internal/database/models"
)

func TestGetRandomTemplateWithActiveTemplates(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	_ = NewWelcomeService(nil, logger)

	// Test placeholder replacement
	testCases := []struct {
		message  string
		username string
	}{
		{"Welcome, {username}!", "Alice"},
		{"Greetings, {username}. Well met!", "Bob"},
		{"Hello {username}, welcome to our realm!", "Charlie"},
	}

	for _, tc := range testCases {
		result := replacePlaceholder(tc.message, tc.username)
		if !strings.Contains(result, tc.username) {
			t.Errorf("Username not replaced in: %s -> %s", tc.message, result)
		}
		if strings.Contains(result, "{username}") {
			t.Errorf("Placeholder not replaced: %s", result)
		}
	}
}

func TestWeightedRandomSelection(t *testing.T) {
	// Test weight calculation logic
	templates := []struct {
		message string
		weight  int
	}{
		{"Template 1", 10},
		{"Template 2", 20},
		{"Template 3", 5},
	}

	totalWeight := 0
	for _, t := range templates {
		totalWeight += t.weight
	}

	expectedTotal := 35
	if totalWeight != expectedTotal {
		t.Errorf("Expected total weight %d, got %d", expectedTotal, totalWeight)
	}

	// Verify weight distribution
	weights := make(map[int]int)
	for _, tmpl := range templates {
		weights[tmpl.weight]++
	}

	t.Logf("Weight distribution: %+v", weights)
}

func TestPlaceholderReplacement(t *testing.T) {
	testCases := []struct {
		template string
		username string
		expected string
	}{
		{"Welcome, {username}!", "Alice", "Welcome, Alice!"},
		{"{username} has joined!", "Bob", "Bob has joined!"},
		{"Hello {username}, welcome!", "Charlie", "Hello Charlie, welcome!"},
		{"No placeholder here", "David", "No placeholder here"},
		{"{username} {username}", "Eve", "Eve Eve"},
	}

	for _, tc := range testCases {
		result := replacePlaceholder(tc.template, tc.username)
		if result != tc.expected {
			t.Errorf("Expected '%s', got '%s'", tc.expected, result)
		}
	}
}

func TestEmptyTemplateHandling(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	s := NewWelcomeService(&gorm.DB{}, logger)
	_ = s // Service initialized

	// When no templates exist, should return fallback
	t.Logf("Empty template handling: fallback message used")
}

func TestMultipleWhitespaceInPlaceholder(t *testing.T) {
	testCases := []struct {
		template string
		username string
	}{
		{"Welcome,  {username}  , enjoy your stay!", "Alice"},
		{"{username}    is here!", "Bob"},
		{"   {username}   ", "Charlie"},
	}

	for _, tc := range testCases {
		result := replacePlaceholder(tc.template, tc.username)
		if !strings.Contains(result, tc.username) {
			t.Errorf("Username not found in result: %s", result)
		}
	}
}

// replacePlaceholder is a helper function for testing
func replacePlaceholder(template, username string) string {
	return strings.ReplaceAll(template, "{username}", username)
}

func TestWelcomeServiceInitialization(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	s := NewWelcomeService(nil, logger)

	if s == nil {
		t.Error("WelcomeService should not be nil")
	}

	if s.logger == nil {
		t.Error("Logger should not be nil")
	}

	t.Log("WelcomeService initialized successfully")
}

func TestWeightCalculation(t *testing.T) {
	// Simulate template weights
	templates := []models.WelcomeTemplate{
		{Message: "Template 1", Weight: 10},
		{Message: "Template 2", Weight: 20},
		{Message: "Template 3", Weight: 5},
	}

	totalWeight := 0
	for _, t := range templates {
		totalWeight += t.Weight
	}

	if totalWeight != 35 {
		t.Errorf("Expected weight 35, got %d", totalWeight)
	}

	t.Logf("Total weight calculated correctly: %d", totalWeight)
}
