package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"living-lands-bot/pkg/ollama"
)

// DefaultRelevanceThreshold is the maximum cosine distance for a document to be considered relevant.
// ChromaDB uses cosine distance (0 = identical, 2 = opposite). Lower values are more similar.
// A threshold of 1.0 is permissive - it allows moderately relevant documents.
// For high precision, use 0.5-0.7. For higher recall (more results), use 0.8-1.2.
const DefaultRelevanceThreshold = 1.0

// RAGService handles retrieval-augmented generation queries against ChromaDB.
// Thread-safe: all operations on collectionID are protected by mu mutex.
type RAGService struct {
	chromaURL          string
	ollamaClient       *ollama.Client
	httpClient         *http.Client
	embedModel         string
	logger             *slog.Logger
	collectionID       string       // Cached collection ID for v2 API (protected by mu)
	collectionName     string       // Collection name for retrieval
	relevanceThreshold float32      // Maximum distance for relevant documents
	mu                 sync.RWMutex // Protects collectionID field
}

// Document represents a document to be indexed in the RAG system.
type Document struct {
	ID       string
	Text     string
	Metadata map[string]interface{}
}

// ChromaQueryRequest represents the request body for ChromaDB query endpoint
type ChromaQueryRequest struct {
	QueryEmbeddings [][]float32 `json:"query_embeddings,omitempty"`
	QueryTexts      []string    `json:"query_texts,omitempty"`
	NResults        int         `json:"n_results"`
	Include         []string    `json:"include,omitempty"`
}

// ChromaQueryResponse represents the response from ChromaDB query endpoint
type ChromaQueryResponse struct {
	IDs        [][]string                 `json:"ids"`
	Documents  [][]string                 `json:"documents"`
	Embeddings [][][]float32              `json:"embeddings"`
	Distances  [][]float32                `json:"distances"`
	Metadatas  [][]map[string]interface{} `json:"metadatas"`
}

// ChromaAddRequest represents the request body for ChromaDB add endpoint
type ChromaAddRequest struct {
	IDs        []string                 `json:"ids"`
	Embeddings [][]float32              `json:"embeddings"`
	Documents  []string                 `json:"documents"`
	Metadatas  []map[string]interface{} `json:"metadatas,omitempty"`
}

// NewRAGService initializes a RAG service with ChromaDB and Ollama clients.
func NewRAGService(chromaURL string, ollamaClient *ollama.Client, embedModel string, logger *slog.Logger) (*RAGService, error) {
	s := &RAGService{
		chromaURL:          chromaURL,
		ollamaClient:       ollamaClient,
		embedModel:         embedModel,
		logger:             logger,
		collectionName:     "livinglands_docs",
		relevanceThreshold: DefaultRelevanceThreshold,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	logger.Info("chromadb service initialized", "url", chromaURL, "embedding_model", embedModel, "relevance_threshold", s.relevanceThreshold)
	return s, nil
}

// SetRelevanceThreshold sets the maximum distance for documents to be considered relevant.
// Values range from 0 (exact match) to 2 (completely opposite).
// Thread-safe.
func (s *RAGService) SetRelevanceThreshold(threshold float32) {
	s.relevanceThreshold = threshold
	s.logger.Info("relevance threshold updated", "threshold", threshold)
}

// Query retrieves the top-N most relevant documents for a given question.
func (s *RAGService) Query(ctx context.Context, question string, nResults int) ([]string, error) {
	// 0. Ensure collection exists and get its ID
	if err := s.ensureCollection(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure collection exists: %w", err)
	}

	// 1. Generate embedding for the question using Ollama
	embedding, err := s.ollamaClient.Embed(ctx, s.embedModel, question)
	if err != nil {
		return nil, fmt.Errorf("failed to generate question embedding: %w", err)
	}

	if len(embedding) == 0 {
		return nil, fmt.Errorf("empty embedding received from ollama")
	}

	s.logger.Debug("question embedded", "length", len(embedding))

	// 2. Query ChromaDB with the embedding using v2 API
	queryReq := ChromaQueryRequest{
		QueryEmbeddings: [][]float32{embedding},
		NResults:        nResults,
		Include:         []string{"documents", "distances"},
	}

	body, err := json.Marshal(queryReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query request: %w", err)
	}

	// Use v2 API endpoint with collection ID
	url := fmt.Sprintf("%s/api/v2/tenants/default_tenant/databases/default_database/collections/%s/query",
		s.chromaURL, s.collectionID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("chromadb query request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Collection doesn't exist yet, return empty results
		s.logger.Debug("collection not found, returning empty results")
		return []string{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("chromadb returned %d: %s", resp.StatusCode, string(respBody))
	}

	var queryResp ChromaQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&queryResp); err != nil {
		return nil, fmt.Errorf("failed to decode chromadb response: %w", err)
	}

	// 3. Extract document texts from the query result, filtering by relevance threshold
	var contexts []string
	var filteredCount int

	for i, docs := range queryResp.Documents {
		for j, doc := range docs {
			if doc == "" {
				continue
			}

			// Check if we have distance information for this document
			var distance float32 = 0
			if i < len(queryResp.Distances) && j < len(queryResp.Distances[i]) {
				distance = queryResp.Distances[i][j]
			}

			// Get metadata for better debugging
			var metadata map[string]interface{}
			if i < len(queryResp.Metadatas) && j < len(queryResp.Metadatas[i]) {
				metadata = queryResp.Metadatas[i][j]
			}

			// Filter out documents that exceed the relevance threshold
			if distance > s.relevanceThreshold {
				filteredCount++
				s.logger.Debug("document filtered due to low relevance",
					"distance", distance,
					"threshold", s.relevanceThreshold,
					"source", getMetadataSource(metadata),
					"doc_preview", truncateString(doc, 80),
				)
				continue
			}

			contexts = append(contexts, doc)
			s.logger.Info("document accepted for RAG context",
				"distance", distance,
				"threshold", s.relevanceThreshold,
				"source", getMetadataSource(metadata),
				"doc_preview", truncateString(doc, 100),
			)
		}
	}

	s.logger.Info("rag query complete",
		"question", question,
		"results", len(contexts),
		"filtered", filteredCount,
		"threshold", s.relevanceThreshold,
	)
	return contexts, nil
}

// truncateString truncates a string to maxLen characters, adding ellipsis if needed.
// Uses runes to properly handle multi-byte UTF-8 characters without corruption.
func truncateString(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	// Truncate to maxLen runes and add ellipsis
	return string(runes[:maxLen]) + "..."
}

// getMetadataSource extracts the source path from metadata if available.
func getMetadataSource(metadata map[string]interface{}) string {
	if metadata == nil {
		return "unknown"
	}
	if source, ok := metadata["source"].(string); ok {
		return source
	}
	return "unknown"
}

// AddDocuments adds multiple documents to the RAG collection with generated embeddings.
func (s *RAGService) AddDocuments(ctx context.Context, docs []Document) error {
	if len(docs) == 0 {
		return nil
	}

	// First, ensure the collection exists
	if err := s.ensureCollection(ctx); err != nil {
		return fmt.Errorf("failed to ensure collection exists: %w", err)
	}

	// Generate embeddings for all documents
	var embeddings [][]float32
	var ids []string
	var documents []string
	var metadatas []map[string]interface{}

	for _, doc := range docs {
		// Generate embedding
		embedding, err := s.ollamaClient.Embed(ctx, s.embedModel, doc.Text)
		if err != nil {
			s.logger.Error("failed to generate embedding", "doc_id", doc.ID, "error", err)
			continue
		}

		embeddings = append(embeddings, embedding)
		ids = append(ids, doc.ID)
		documents = append(documents, doc.Text)
		metadatas = append(metadatas, doc.Metadata)
	}

	if len(embeddings) == 0 {
		return fmt.Errorf("failed to generate embeddings for any documents")
	}

	// Add to ChromaDB collection using v2 API
	addReq := ChromaAddRequest{
		IDs:        ids,
		Embeddings: embeddings,
		Documents:  documents,
		Metadatas:  metadatas,
	}

	body, err := json.Marshal(addReq)
	if err != nil {
		return fmt.Errorf("failed to marshal add request: %w", err)
	}

	// Use v2 API endpoint with collection ID
	url := fmt.Sprintf("%s/api/v2/tenants/default_tenant/databases/default_database/collections/%s/add",
		s.chromaURL, s.collectionID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("chromadb add request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb returned %d: %s", resp.StatusCode, string(respBody))
	}

	s.logger.Info("documents added to rag collection", "count", len(ids))
	return nil
}

// DeleteDocument removes a document from the RAG collection.
func (s *RAGService) DeleteDocument(ctx context.Context, id string) error {
	// Ensure collection exists
	if err := s.ensureCollection(ctx); err != nil {
		return fmt.Errorf("failed to ensure collection exists: %w", err)
	}

	// ChromaDB delete endpoint
	reqBody := map[string]interface{}{
		"ids": []string{id},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal delete request: %w", err)
	}

	// Use v2 API endpoint with collection ID
	url := fmt.Sprintf("%s/api/v2/tenants/default_tenant/databases/default_database/collections/%s/delete",
		s.chromaURL, s.collectionID)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("chromadb delete request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb returned %d: %s", resp.StatusCode, string(respBody))
	}

	s.logger.Info("document deleted from rag collection", "id", id)
	return nil
}

// Count returns the number of documents in the collection.
func (s *RAGService) Count(ctx context.Context) (int, error) {
	// Ensure collection exists
	if err := s.ensureCollection(ctx); err != nil {
		return 0, fmt.Errorf("failed to ensure collection exists: %w", err)
	}

	// Use v2 API endpoint with collection ID for count
	url := fmt.Sprintf("%s/api/v2/tenants/default_tenant/databases/default_database/collections/%s/count",
		s.chromaURL, s.collectionID)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create http request: %w", err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("chromadb count request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("chromadb returned %d: %s", resp.StatusCode, string(respBody))
	}

	// ChromaDB v2 returns just an integer for count
	var count int
	if err := json.NewDecoder(resp.Body).Decode(&count); err != nil {
		return 0, fmt.Errorf("failed to decode count response: %w", err)
	}

	return count, nil
}

// ensureCollection creates the collection if it doesn't exist using v2 API.
// It caches the collection ID for subsequent operations.
// Thread-safe: uses mutex to prevent concurrent initialization.
func (s *RAGService) ensureCollection(ctx context.Context) error {
	// Check if collection ID is already cached (read lock)
	s.mu.RLock()
	if s.collectionID != "" {
		s.mu.RUnlock()
		return nil
	}
	s.mu.RUnlock()

	// First, try to get the collection by name (GET is safe and fast)
	collURL := fmt.Sprintf("%s/api/v2/tenants/default_tenant/databases/default_database/collections/%s",
		s.chromaURL, s.collectionName)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", collURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create get collection request: %w", err)
	}

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("chromadb get collection request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		// Collection exists, extract the ID
		var collResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&collResp); err != nil {
			return fmt.Errorf("failed to decode collection response: %w", err)
		}

		collID, ok := collResp["id"].(string)
		if !ok {
			return fmt.Errorf("collection response missing or invalid id field")
		}

		// Write lock to update cached collection ID
		s.mu.Lock()
		s.collectionID = collID
		s.mu.Unlock()
		s.logger.Debug("collection retrieved", "collection", s.collectionName, "id", collID)
		return nil
	}

	if resp.StatusCode != http.StatusNotFound {
		// Some other error occurred
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("chromadb get collection returned %d: %s", resp.StatusCode, string(respBody))
	}

	// Collection doesn't exist, create it
	return s.createCollection(ctx)
}

// createCollection creates a new collection in ChromaDB using v2 API.
func (s *RAGService) createCollection(ctx context.Context) error {
	reqBody := map[string]interface{}{
		"name":     s.collectionName,
		"metadata": map[string]interface{}{"hnsw:space": "cosine"},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal create request: %w", err)
	}

	createURL := fmt.Sprintf("%s/api/v2/tenants/default_tenant/databases/default_database/collections",
		s.chromaURL)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", createURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create collection request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("chromadb create collection request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle both success (201) and already-exists (error) cases
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		respStr := string(respBody)

		// If collection already exists, try to get it by name again
		if resp.StatusCode == http.StatusConflict || strings.Contains(respStr, "already exists") {
			return s.ensureCollection(ctx)
		}

		return fmt.Errorf("chromadb create collection returned %d: %s", resp.StatusCode, respStr)
	}

	// Extract the collection ID from the response
	var collResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&collResp); err != nil {
		return fmt.Errorf("failed to decode create collection response: %w", err)
	}

	collID, ok := collResp["id"].(string)
	if !ok {
		return fmt.Errorf("collection response missing or invalid id field")
	}

	// Write lock to update cached collection ID
	s.mu.Lock()
	s.collectionID = collID
	s.mu.Unlock()
	s.logger.Info("collection created", "collection", s.collectionName, "id", collID)
	return nil
}

// EnsureCollectionPublic is a public wrapper for testing ensureCollection.
// It ensures the collection exists and caches the collection ID.
func (s *RAGService) EnsureCollectionPublic(ctx context.Context) error {
	return s.ensureCollection(ctx)
}
