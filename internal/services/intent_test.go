package services

import (
	"testing"
)

func TestClassifyIntent_Conversational(t *testing.T) {
	tests := []struct {
		query    string
		expected QueryIntent
	}{
		// Greetings
		{"hello", IntentConversational},
		{"hi", IntentConversational},
		{"hey there", IntentConversational},
		{"Hello!", IntentConversational},
		{"good morning", IntentConversational},

		// Location/identity questions
		{"where am i?", IntentConversational},
		{"who am i", IntentConversational},
		{"who are you", IntentConversational},
		{"what is this", IntentConversational},
		{"what's this place", IntentConversational},

		// Status questions
		{"how are you", IntentConversational},
		{"how's it going", IntentConversational},

		// Simple responses
		{"thanks", IntentConversational},
		{"thank you", IntentConversational},
		{"ok", IntentConversational},
		{"cool", IntentConversational},
		{"bye", IntentConversational},

		// Testing
		{"test", IntentConversational},
		{"ping", IntentConversational},

		// Short queries without knowledge keywords
		{"yo", IntentConversational},
		{"sup", IntentConversational},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := ClassifyIntent(tt.query)
			if result != tt.expected {
				t.Errorf("ClassifyIntent(%q) = %v, want %v", tt.query, result.String(), tt.expected.String())
			}
		})
	}
}

func TestClassifyIntent_Knowledge(t *testing.T) {
	tests := []struct {
		query    string
		expected QueryIntent
	}{
		// Mod-specific questions
		{"how does the metabolism system work?", IntentKnowledge},
		{"tell me about creatures in living lands", IntentKnowledge},
		{"what is the ECS architecture?", IntentKnowledge},
		{"how do I install the mod?", IntentKnowledge},
		{"where can I download the mod?", IntentKnowledge},
		{"explain worldgen v2", IntentKnowledge},
		{"what biomes are available?", IntentKnowledge},

		// API/technical questions
		{"how do I use the hytale plugin api?", IntentKnowledge},
		{"what is procedural generation?", IntentKnowledge},
		{"how to configure the server?", IntentKnowledge},

		// Feature questions
		{"is there a hunger system?", IntentKnowledge},
		{"can i add custom creatures?", IntentKnowledge},

		// General questions that look like they need documentation
		{"how does the metabolism feature work?", IntentKnowledge},
		{"what is the difference between v1 and v2?", IntentKnowledge},

		// Questions that could seem like navigation but are really knowledge
		{"where can i report bugs?", IntentKnowledge},
		{"how do i find the wiki?", IntentKnowledge},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := ClassifyIntent(tt.query)
			if result != tt.expected {
				t.Errorf("ClassifyIntent(%q) = %v, want %v", tt.query, result.String(), tt.expected.String())
			}
		})
	}
}

func TestClassifyIntent_Navigation(t *testing.T) {
	tests := []struct {
		query    string
		expected QueryIntent
	}{
		// Explicitly asking about channels
		{"where is the bug report channel?", IntentNavigation},
		{"where is the changelog channel?", IntentNavigation},
		{"where is the support channel?", IntentNavigation},
		{"find the help channel", IntentNavigation},
		{"which channel for announcements?", IntentNavigation},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := ClassifyIntent(tt.query)
			if result != tt.expected {
				t.Errorf("ClassifyIntent(%q) = %v, want %v", tt.query, result.String(), tt.expected.String())
			}
		})
	}
}

func TestClassifyIntent_AccountHelp(t *testing.T) {
	tests := []struct {
		query    string
		expected QueryIntent
	}{
		{"how do i link my account?", IntentAccountHelp},
		{"how to verify my account", IntentAccountHelp},
		{"connect discord to hytale account", IntentAccountHelp},
		{"account linking help", IntentAccountHelp},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			result := ClassifyIntent(tt.query)
			if result != tt.expected {
				t.Errorf("ClassifyIntent(%q) = %v, want %v", tt.query, result.String(), tt.expected.String())
			}
		})
	}
}

func TestQueryIntent_NeedsRAG(t *testing.T) {
	tests := []struct {
		intent   QueryIntent
		needsRAG bool
	}{
		{IntentKnowledge, true},
		{IntentConversational, false},
		{IntentNavigation, false},
		{IntentAccountHelp, false},
	}

	for _, tt := range tests {
		t.Run(tt.intent.String(), func(t *testing.T) {
			result := tt.intent.NeedsRAG()
			if result != tt.needsRAG {
				t.Errorf("%v.NeedsRAG() = %v, want %v", tt.intent.String(), result, tt.needsRAG)
			}
		})
	}
}

func TestQueryIntent_String(t *testing.T) {
	tests := []struct {
		intent   QueryIntent
		expected string
	}{
		{IntentKnowledge, "knowledge"},
		{IntentConversational, "conversational"},
		{IntentNavigation, "navigation"},
		{IntentAccountHelp, "account_help"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.intent.String()
			if result != tt.expected {
				t.Errorf("%v.String() = %v, want %v", tt.intent, result, tt.expected)
			}
		})
	}
}
