package services

import (
	"strings"
)

// SanitizePromptInput sanitizes user input for inclusion in LLM prompts.
// This prevents prompt injection attacks by escaping or removing prompt delimiters
// and other potentially problematic characters.
func SanitizePromptInput(input string) string {
	if input == "" {
		return ""
	}

	// Replace common prompt delimiters and injection patterns
	// This is not meant to be bulletproof, but to prevent common attacks
	replacements := map[string]string{
		// Prompt template injections
		"User:":      "[User]",
		"user:":      "[user]",
		"Assistant:": "[Assistant]",
		"assistant:": "[assistant]",
		"System:":    "[System]",
		"system:":    "[system]",

		// Common control sequences
		"\x00": "", // null byte
		"\x1a": "", // EOF
		"\x1b": "", // ESC

		// Excessive newlines (more than 2 in a row are suspicious)
		"\n\n\n":   "\n\n",
		"\n\n\n\n": "\n\n",
	}

	result := input
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Trim excessive whitespace at start/end
	result = strings.TrimSpace(result)

	// Limit length to prevent token exhaustion attacks
	// Most questions should be under 500 chars
	maxLen := 2000 // Allow reasonably long questions, but prevent abuse
	if len(result) > maxLen {
		result = result[:maxLen]
	}

	return result
}

// ValidatePromptInput checks if input is valid for LLM processing.
// Returns true if input is valid, false otherwise.
func ValidatePromptInput(input string) bool {
	if input == "" {
		return false
	}

	// Check for suspiciously long inputs
	if len(input) > 2000 {
		return false
	}

	// Check for excessive control characters
	controlCharCount := 0
	for _, r := range input {
		if r < 32 && r != '\n' && r != '\t' && r != '\r' {
			controlCharCount++
		}
	}
	if controlCharCount > 5 {
		return false
	}

	return true
}
