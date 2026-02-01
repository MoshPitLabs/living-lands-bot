package services

import (
	"regexp"
	"strings"
)

// QueryIntent represents the type of user query.
type QueryIntent int

const (
	// IntentKnowledge - User is asking about mod/game knowledge (use RAG)
	IntentKnowledge QueryIntent = iota
	// IntentConversational - Simple greeting/chat (skip RAG)
	IntentConversational
	// IntentNavigation - Asking for directions/channels (use /guide)
	IntentNavigation
	// IntentAccountHelp - Asking about account linking (use /link)
	IntentAccountHelp
	// IntentIdentity - User asking about identity/location (needs persona response)
	IntentIdentity
)

// String returns the string representation of the intent.
func (i QueryIntent) String() string {
	switch i {
	case IntentKnowledge:
		return "knowledge"
	case IntentConversational:
		return "conversational"
	case IntentNavigation:
		return "navigation"
	case IntentAccountHelp:
		return "account_help"
	case IntentIdentity:
		return "identity"
	default:
		return "unknown"
	}
}

// NeedsRAG returns true if this intent requires RAG context.
func (i QueryIntent) NeedsRAG() bool {
	return i == IntentKnowledge
}

// conversationalPatterns are phrases that indicate simple chat.
// These are checked with exact boundaries to avoid false matches.
var conversationalPatterns = []string{
	// Greetings
	"hello", "hi", "hey", "howdy", "greetings", "good morning", "good evening",
	"good afternoon", "good night", "what's up", "sup", "yo",
	// Status questions
	"how are you", "how's it going", "how do you do", "what's new",
	// Simple responses
	"thanks", "thank you", "ok", "okay", "cool", "nice", "great",
	"bye", "goodbye", "see you", "later", "cya",
	// Testing
	"test", "testing", "ping", "pong",
}

// exactConversationalMatches are queries that must match exactly (after trimming/lowercase)
var exactConversationalMatches = []string{
	"what is this",
	"what's this",
	"hi",
	"hey",
	"hello",
	"yo",
	"sup",
	"test",
	"ping",
}

// navigationKeywords trigger navigation intent.
// These are specific to Discord channel navigation.
var navigationKeywords = []string{
	"channel", "help channel", "support channel",
	"bug report channel", "changelog channel", "wiki channel",
	"rules channel", "announcements channel", "general channel",
	"announcements",
}

// accountKeywords trigger account help intent.
var accountKeywords = []string{
	"link", "account", "verify", "verification", "connect",
	"hytale account", "discord account", "linking",
}

// identityKeywords trigger identity/location intent (needs persona response).
var identityKeywords = []string{
	"where am i", "who are you", "what are you", "what's this place",
	"who am i", "what is this place", "tell me about yourself",
}

// knowledgeKeywords strongly indicate knowledge queries.
var knowledgeKeywords = []string{
	// Mod-specific terms
	"mod", "living lands", "metabolism", "creature", "biome", "worldgen",
	"procedural", "generation", "ecs", "component", "entity",
	"hytale", "plugin", "api", "server", "config", "configuration",
	"architecture",
	// Question patterns that need knowledge
	"how does", "how do", "how to", "what is the", "explain",
	"tell me about", "describe", "difference between", "why does",
	"when does", "where does", "can i", "is there", "feature",
	"install", "download", "setup", "curseforge",
}

// ClassifyIntent analyzes a user query and determines its intent.
func ClassifyIntent(query string) QueryIntent {
	normalized := strings.ToLower(strings.TrimSpace(query))

	// Strip trailing punctuation for exact matching
	normalizedNoPunc := strings.TrimRight(normalized, "?!.,")

	// Check for exact conversational matches first
	for _, exact := range exactConversationalMatches {
		if normalizedNoPunc == exact {
			return IntentConversational
		}
	}

	// Check for short queries (likely conversational)
	wordCount := len(strings.Fields(normalized))
	if wordCount <= 2 {
		// Very short queries are usually conversational unless they contain knowledge keywords
		if !containsAnyKeyword(normalized, knowledgeKeywords) {
			return IntentConversational
		}
	}

	// Check for account-related queries first (high priority)
	if containsAnyKeyword(normalized, accountKeywords) {
		return IntentAccountHelp
	}

	// Check for identity/location queries (needs persona response)
	if containsAnyKeyword(normalized, identityKeywords) {
		return IntentIdentity
	}

	// Check for navigation keywords BEFORE conversational patterns
	// This ensures "where is the support channel" is navigation, not conversational
	if containsAnyKeyword(normalized, navigationKeywords) {
		return IntentNavigation
	}

	// Check for knowledge keywords BEFORE conversational
	if containsAnyKeyword(normalized, knowledgeKeywords) {
		return IntentKnowledge
	}

	// Check for conversational patterns (partial match) - lower priority
	for _, pattern := range conversationalPatterns {
		if strings.Contains(normalized, pattern) {
			return IntentConversational
		}
	}

	// Default: if it looks like a question, treat as knowledge query
	if isQuestion(normalized) {
		return IntentKnowledge
	}

	// Otherwise, treat as conversational
	return IntentConversational
}

// containsAnyKeyword checks if text contains any of the given keywords.
func containsAnyKeyword(text string, keywords []string) bool {
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

// questionPatterns for detecting questions.
var questionPattern = regexp.MustCompile(`(?i)^(what|how|why|when|where|who|which|can|does|is|are|do|will|would|could|should)\b`)

// isQuestion checks if the text appears to be a question.
func isQuestion(text string) bool {
	// Check for question mark
	if strings.Contains(text, "?") {
		return true
	}

	// Check for question word patterns
	return questionPattern.MatchString(text)
}
