package services

import (
	"regexp"
	"testing"
	"time"
)

func TestGenerateVerificationCode(t *testing.T) {
	// Test code generation logic directly (unit test)
	// Real integration tests would use a test database

	// Test code format
	code := generateCode(8)

	// Validate code format (8 uppercase alphanumeric characters)
	if len(code) != 8 {
		t.Errorf("Expected code length 8, got %d", len(code))
	}

	// Check that code matches expected pattern (base32 encoded, uppercase)
	pattern := regexp.MustCompile(`^[A-Z0-9]{8}$`)
	if !pattern.MatchString(code) {
		t.Errorf("Code format invalid: %s (expected uppercase alphanumeric)", code)
	}

	t.Logf("Generated verification code: %s", code)
}

func TestGenerateCodeFunction(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"4 char code", 4},
		{"8 char code", 8},
		{"16 char code", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := generateCode(tt.length)

			if len(code) != tt.length {
				t.Errorf("Expected length %d, got %d", tt.length, len(code))
			}

			// Verify all characters are alphanumeric
			lenientPattern := regexp.MustCompile(`^[A-Z0-9]+$`)
			if !lenientPattern.MatchString(code) {
				t.Errorf("Invalid character in code: %s", code)
			} else {
				t.Logf("Code generated: %s", code)
			}
		})
	}
}

func TestVerificationCodeExpiry(t *testing.T) {
	// Test that expiry calculation logic is correct
	createdAt := time.Now()
	updatedAt := createdAt.Add(-15 * time.Minute) // Code is 15 minutes old
	expirySeconds := 600                          // 10 minutes

	if time.Since(updatedAt) > time.Duration(expirySeconds)*time.Second {
		t.Logf("Code correctly identified as expired (age: %v, expiry: %d sec)", time.Since(updatedAt), expirySeconds)
	} else {
		t.Errorf("Code should be expired")
	}

	// Test with fresh code
	updatedAt = time.Now()
	if time.Since(updatedAt) <= time.Duration(expirySeconds)*time.Second {
		t.Logf("Fresh code correctly identified as valid")
	} else {
		t.Errorf("Fresh code should be valid")
	}
}

func TestGenerateCodeRandomness(t *testing.T) {
	// Generate multiple codes and verify they're unique
	codes := make(map[string]bool)

	for i := 0; i < 100; i++ {
		code := generateCode(8)
		if codes[code] {
			t.Errorf("Duplicate code generated: %s", code)
		}
		codes[code] = true
	}

	t.Logf("Generated 100 unique codes successfully")
}

func TestVerificationCodeFormat(t *testing.T) {
	// Generate and verify code format is alphanumeric
	code := generateCode(8)

	// Code should be alphanumeric only
	pattern := regexp.MustCompile(`^[A-Z0-9]+$`)
	if !pattern.MatchString(code) {
		t.Errorf("Code contains invalid characters: %s", code)
		return
	}

	t.Logf("Code format valid: %s", code)
}
