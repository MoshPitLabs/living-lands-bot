//go:build integration

package services

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"living-lands-bot/pkg/ollama"
)

// TestRAGServiceV2APIIntegration tests the v2 API with a real ChromaDB instance.
// Run with: go test -tags=integration ./internal/services/...
func TestRAGServiceV2APIIntegration(t *testing.T) {
	// Skip if ChromaDB not available
	if os.Getenv("CHROMA_URL") == "" {
		t.Skip("Skipping integration test: CHROMA_URL not set")
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	chromaURL := os.Getenv("CHROMA_URL")
	if chromaURL == "" {
		chromaURL = "http://localhost:8001"
	}

	ollamaClient := ollama.NewClient("http://localhost:11434")
	ragSvc, err := NewRAGService(chromaURL, ollamaClient, "nomic-embed-text", logger)
	if err != nil {
		t.Fatalf("Failed to create RAG service: %v", err)
	}

	ctx := context.Background()

	// Test 1: ensureCollection should create the collection
	t.Run("EnsureCollection", func(t *testing.T) {
		if err := ragSvc.EnsureCollectionPublic(ctx); err != nil {
			t.Fatalf("ensureCollection failed: %v", err)
		}

		if ragSvc.collectionID == "" {
			t.Error("collectionID should be set after ensureCollection")
		}
		t.Logf("Collection ID: %s", ragSvc.collectionID)
	})

	// Test 2: Add documents
	t.Run("AddDocuments", func(t *testing.T) {
		docs := []Document{
			{
				ID:   "integration_test_doc1",
				Text: "Living Lands is a Hytale mod",
				Metadata: map[string]interface{}{
					"source": "test.md",
					"chunk":  0,
				},
			},
			{
				ID:   "integration_test_doc2",
				Text: "The mod adds new features to Hytale",
				Metadata: map[string]interface{}{
					"source": "test.md",
					"chunk":  1,
				},
			},
		}

		if err := ragSvc.AddDocuments(ctx, docs); err != nil {
			t.Fatalf("AddDocuments failed: %v", err)
		}
	})

	// Test 3: Count documents
	t.Run("Count", func(t *testing.T) {
		count, err := ragSvc.Count(ctx)
		if err != nil {
			t.Fatalf("Count failed: %v", err)
		}

		if count < 2 {
			t.Errorf("Expected at least 2 documents, got %d", count)
		}
		t.Logf("Document count: %d", count)
	})

	// Test 4: Query documents
	t.Run("Query", func(t *testing.T) {
		results, err := ragSvc.Query(ctx, "Hytale mod features", 5)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}

		if len(results) == 0 {
			t.Error("Expected at least 1 result from query")
		}
		t.Logf("Query returned %d results", len(results))
		for i, result := range results {
			t.Logf("  Result %d: %s", i+1, result)
		}
	})

	// Test 5: Delete a document
	t.Run("DeleteDocument", func(t *testing.T) {
		if err := ragSvc.DeleteDocument(ctx, "integration_test_doc1"); err != nil {
			t.Fatalf("DeleteDocument failed: %v", err)
		}
	})

	// Test 6: Verify count after delete
	t.Run("CountAfterDelete", func(t *testing.T) {
		count, err := ragSvc.Count(ctx)
		if err != nil {
			t.Fatalf("Count after delete failed: %v", err)
		}

		if count < 1 {
			t.Errorf("Expected at least 1 document after delete, got %d", count)
		}
		t.Logf("Document count after delete: %d", count)
	})
}
