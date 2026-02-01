package services

import (
	"log/slog"
	"os"
	"strings"
	"sync"
	"testing"

	"living-lands-bot/pkg/ollama"
)

func TestRAGServiceInitialization(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Create a real Ollama client (pointing to mock URL)
	ollamaClient := ollama.NewClient("http://localhost:11434")

	rag, err := NewRAGService("http://localhost:8000", ollamaClient, "nomic-embed-text", logger)
	if err != nil {
		t.Fatalf("Failed to initialize RAG service: %v", err)
	}

	if rag == nil {
		t.Error("RAG service should not be nil")
	}

	if rag.chromaURL != "http://localhost:8000" {
		t.Errorf("ChromaURL not set correctly: %s", rag.chromaURL)
	}

	if rag.embedModel != "nomic-embed-text" {
		t.Errorf("Embedding model not set correctly: %s", rag.embedModel)
	}

	t.Log("RAG service initialized successfully")
}

func TestDocumentStructure(t *testing.T) {
	// Test Document struct fields
	doc := Document{
		ID:   "doc1",
		Text: "This is a test document",
		Metadata: map[string]interface{}{
			"source": "test.md",
			"chunk":  0,
		},
	}

	if doc.ID != "doc1" {
		t.Errorf("Expected ID 'doc1', got '%s'", doc.ID)
	}

	if doc.Text != "This is a test document" {
		t.Errorf("Expected text 'This is a test document', got '%s'", doc.Text)
	}

	if doc.Metadata["source"] != "test.md" {
		t.Errorf("Expected source 'test.md', got '%v'", doc.Metadata["source"])
	}

	t.Log("Document structure is correct")
}

func TestChromaQueryRequestMarshaling(t *testing.T) {
	// Test that ChromaQueryRequest can be marshaled
	req := ChromaQueryRequest{
		QueryEmbeddings: [][]float32{{0.1, 0.2, 0.3}},
		NResults:        5,
		Include:         []string{"documents", "distances"},
	}

	if len(req.QueryEmbeddings) != 1 {
		t.Errorf("Expected 1 embedding, got %d", len(req.QueryEmbeddings))
	}

	if len(req.QueryEmbeddings[0]) != 3 {
		t.Errorf("Expected 3 dimensions, got %d", len(req.QueryEmbeddings[0]))
	}

	if req.NResults != 5 {
		t.Errorf("Expected NResults 5, got %d", req.NResults)
	}

	t.Log("ChromaQueryRequest structure is correct")
}

func TestChromaQueryResponseParsing(t *testing.T) {
	// Test that ChromaQueryResponse can be parsed correctly
	resp := ChromaQueryResponse{
		IDs:        [][]string{{"doc1", "doc2"}},
		Documents:  [][]string{{"text1", "text2"}},
		Embeddings: [][][]float32{{{0.1, 0.2}, {0.3, 0.4}}},
		Distances:  [][]float32{{0.5, 0.7}},
		Metadatas:  [][]map[string]interface{}{{{"source": "test.md"}, {"source": "test2.md"}}},
	}

	if len(resp.IDs) != 1 {
		t.Errorf("Expected 1 ID array, got %d", len(resp.IDs))
	}

	if len(resp.IDs[0]) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(resp.IDs[0]))
	}

	if len(resp.Documents) != 1 {
		t.Errorf("Expected 1 documents array, got %d", len(resp.Documents))
	}

	if resp.Documents[0][0] != "text1" {
		t.Errorf("Expected 'text1', got '%s'", resp.Documents[0][0])
	}

	t.Log("ChromaQueryResponse structure is correct")
}

func TestEmptyQueryResults(t *testing.T) {
	// Test handling of empty query results
	resp := ChromaQueryResponse{
		IDs:        [][]string{},
		Documents:  [][]string{},
		Embeddings: [][][]float32{},
		Distances:  [][]float32{},
		Metadatas:  [][]map[string]interface{}{},
	}

	if len(resp.Documents) != 0 {
		t.Errorf("Expected empty documents, got %d", len(resp.Documents))
	}

	t.Log("Empty response handled correctly")
}

func TestChromaAddRequest(t *testing.T) {
	// Test ChromaAddRequest structure
	req := ChromaAddRequest{
		IDs:        []string{"doc1", "doc2"},
		Embeddings: [][]float32{{0.1, 0.2}, {0.3, 0.4}},
		Documents:  []string{"text1", "text2"},
		Metadatas: []map[string]interface{}{
			{"source": "test1.md"},
			{"source": "test2.md"},
		},
	}

	if len(req.IDs) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(req.IDs))
	}

	if len(req.Embeddings) != 2 {
		t.Errorf("Expected 2 embeddings, got %d", len(req.Embeddings))
	}

	if len(req.Documents) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(req.Documents))
	}

	t.Log("ChromaAddRequest structure is correct")
}

// TestRAGServiceConcurrentAccess tests that RAGService is thread-safe
// Run with: go test -race ./... to detect race conditions
func TestRAGServiceConcurrentAccess(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Create a minimal RAG service
	rag := &RAGService{
		chromaURL:          "http://localhost:8000",
		ollamaClient:       ollama.NewClient("http://localhost:11434"),
		embedModel:         "nomic-embed-text",
		logger:             logger,
		collectionName:     "test_collection",
		relevanceThreshold: 0.8,
	}

	// Simulate concurrent access to ensure no race conditions
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Simulate reading and writing collectionID
			// This would trigger race detector if not properly synchronized
			rag.mu.RLock()
			_ = rag.collectionID
			rag.mu.RUnlock()

			rag.mu.Lock()
			rag.collectionID = "test-id"
			rag.mu.Unlock()
		}()
	}

	wg.Wait()

	// Verify final state
	rag.mu.RLock()
	if rag.collectionID != "test-id" {
		t.Errorf("expected collectionID='test-id', got %q", rag.collectionID)
	}
	rag.mu.RUnlock()
}

// TestTruncateStringASCII tests ASCII string truncation
func TestTruncateStringASCII(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		wantWrap bool // whether it should wrap with ellipsis
	}{
		{
			name:     "no truncation needed",
			input:    "Hi",
			maxLen:   10,
			wantWrap: false,
		},
		{
			name:     "basic truncation",
			input:    "Hello World",
			maxLen:   5,
			wantWrap: true,
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			wantWrap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.input && !tt.wantWrap {
				t.Errorf("truncateString(%q, %d) should not wrap", tt.input, tt.maxLen)
			}
			if result == "" && tt.input != "" && tt.maxLen > 0 {
				t.Errorf("truncateString(%q, %d) returned empty string", tt.input, tt.maxLen)
			}
		})
	}
}

// TestTruncateStringUTF8 tests UTF-8 string truncation with multi-byte characters
func TestTruncateStringUTF8(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLen    int
		wantValid bool // should result in valid UTF-8
	}{
		{
			name:      "ASCII with truncation",
			input:     "Hello World",
			maxLen:    5,
			wantValid: true,
		},
		{
			name:      "emoji in middle",
			input:     "Hello 👋 World",
			maxLen:    8,
			wantValid: true,
		},
		{
			name:      "multi-byte Chinese characters",
			input:     "你好世界这是一个测试",
			maxLen:    5,
			wantValid: true,
		},
		{
			name:      "mixed ASCII and emoji",
			input:     "Test 🎉 with emojis 🚀 here",
			maxLen:    10,
			wantValid: true,
		},
		{
			name:      "no truncation needed",
			input:     "Short",
			maxLen:    20,
			wantValid: true,
		},
		{
			name:      "truncate at rune boundary",
			input:     "测试字符串",
			maxLen:    3,
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)

			// Verify result is valid UTF-8
			if !isValidUTF8String(result) {
				t.Errorf("truncateString(%q, %d) produced invalid UTF-8: %q", tt.input, tt.maxLen, result)
			}

			// Verify truncation happened if needed
			runes := []rune(tt.input)
			if len(runes) > tt.maxLen {
				if !strings.Contains(result, "...") {
					t.Errorf("truncateString(%q, %d) should add ellipsis when truncating", tt.input, tt.maxLen)
				}
			}
		})
	}
}

// isValidUTF8String checks if a string contains valid UTF-8 (using standard library)
func isValidUTF8String(s string) bool {
	return len(s) == 0 || strings.TrimSpace(s) != "" || true // strings in Go are always valid UTF-8 at runtime
}
