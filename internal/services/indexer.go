package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DocumentIndexer handles document processing and RAG indexing.
type DocumentIndexer struct {
	ragService *RAGService
	logger     *slog.Logger
	chunkSize  int // Size of document chunks (characters)
	overlap    int // Overlap between chunks (characters)
}

// NewDocumentIndexer creates a new document indexer.
func NewDocumentIndexer(ragService *RAGService, logger *slog.Logger) *DocumentIndexer {
	return &DocumentIndexer{
		ragService: ragService,
		logger:     logger,
		chunkSize:  500, // 500 character chunks
		overlap:    50,  // 50 character overlap
	}
}

// IndexDirectory recursively indexes all Markdown and TXT files in a directory.
func (d *DocumentIndexer) IndexDirectory(ctx context.Context, dirPath string) error {
	d.logger.Info("starting document indexing", "path", dirPath)

	if _, err := os.Stat(dirPath); err != nil {
		return fmt.Errorf("directory does not exist: %w", err)
	}

	var documents []Document
	var processedCount int
	var skippedCount int

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			d.logger.Error("walk error", "path", path, "error", err)
			return nil // Continue walking
		}

		if info.IsDir() {
			return nil
		}

		// Only process markdown, MDX, and text files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".mdx" && ext != ".txt" {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			d.logger.Error("failed to read file", "path", path, "error", err)
			return nil
		}

		if len(content) == 0 {
			d.logger.Debug("skipping empty file", "path", path)
			skippedCount++
			return nil
		}

		// Calculate checksum for duplicate detection
		hash := sha256.Sum256(content)
		checksum := fmt.Sprintf("%x", hash)
		docID := fmt.Sprintf("%s:%s", path, checksum)

		// Chunk the document
		chunks := d.chunkDocument(string(content), path)
		if len(chunks) == 0 {
			d.logger.Debug("no chunks generated", "path", path)
			skippedCount++
			return nil
		}

		for i, chunk := range chunks {
			chunkID := fmt.Sprintf("%s:chunk_%d", docID, i)
			doc := Document{
				ID:   chunkID,
				Text: chunk,
				Metadata: map[string]interface{}{
					"source":   path,
					"checksum": checksum,
					"chunk":    i,
					"indexed":  time.Now().Unix(),
				},
			}
			documents = append(documents, doc)
		}

		processedCount++
		d.logger.Info("file processed", "path", path, "chunks", len(chunks))
		return nil
	})

	if err != nil {
		return fmt.Errorf("directory walk failed: %w", err)
	}

	if len(documents) == 0 {
		d.logger.Warn("no documents found to index", "path", dirPath)
		return nil
	}

	// Add documents to RAG service in batches to avoid context timeouts
	const batchSize = 25
	totalBatches := (len(documents) + batchSize - 1) / batchSize

	d.logger.Info("adding documents to RAG collection", "total_chunks", len(documents), "batches", totalBatches)

	for i := 0; i < len(documents); i += batchSize {
		end := i + batchSize
		if end > len(documents) {
			end = len(documents)
		}

		batch := documents[i:end]
		batchNum := (i / batchSize) + 1

		d.logger.Info("processing batch", "batch", batchNum, "total_batches", totalBatches, "batch_size", len(batch))

		if err := d.ragService.AddDocuments(ctx, batch); err != nil {
			return fmt.Errorf("failed to add batch %d/%d to RAG: %w", batchNum, totalBatches, err)
		}
	}

	d.logger.Info("document indexing complete",
		"processed_files", processedCount,
		"skipped_files", skippedCount,
		"total_chunks", len(documents),
	)

	return nil
}

// IndexFile indexes a single file.
func (d *DocumentIndexer) IndexFile(ctx context.Context, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("file does not exist: %w", err)
	}

	if info.IsDir() {
		return d.IndexDirectory(ctx, filePath)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".md" && ext != ".mdx" && ext != ".txt" {
		return fmt.Errorf("unsupported file type: %s (only .md, .mdx, and .txt are supported)", ext)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if len(content) == 0 {
		return fmt.Errorf("file is empty")
	}

	// Calculate checksum
	hash := sha256.Sum256(content)
	checksum := fmt.Sprintf("%x", hash)
	docID := fmt.Sprintf("%s:%s", filePath, checksum)

	// Chunk the document
	chunks := d.chunkDocument(string(content), filePath)
	if len(chunks) == 0 {
		return fmt.Errorf("no chunks generated from file")
	}

	var documents []Document
	for i, chunk := range chunks {
		chunkID := fmt.Sprintf("%s:chunk_%d", docID, i)
		doc := Document{
			ID:   chunkID,
			Text: chunk,
			Metadata: map[string]interface{}{
				"source":   filePath,
				"checksum": checksum,
				"chunk":    i,
				"indexed":  time.Now().Unix(),
			},
		}
		documents = append(documents, doc)
	}

	if err := d.ragService.AddDocuments(ctx, documents); err != nil {
		return fmt.Errorf("failed to add documents to RAG: %w", err)
	}

	d.logger.Info("file indexed successfully",
		"path", filePath,
		"chunks", len(chunks),
		"total_chars", len(content),
	)

	return nil
}

// chunkDocument splits a document into overlapping chunks.
func (d *DocumentIndexer) chunkDocument(content, source string) []string {
	if len(content) < d.chunkSize {
		return []string{content}
	}

	var chunks []string
	runeContent := []rune(content)

	for i := 0; i < len(runeContent); i += d.chunkSize - d.overlap {
		end := i + d.chunkSize
		if end > len(runeContent) {
			end = len(runeContent)
		}

		chunk := string(runeContent[i:end])
		if strings.TrimSpace(chunk) == "" {
			continue
		}

		chunks = append(chunks, chunk)

		// Stop if we've reached the end
		if end == len(runeContent) {
			break
		}
	}

	return chunks
}

// GetIndexingStats returns information about the current RAG collection.
func (d *DocumentIndexer) GetIndexingStats(ctx context.Context) (map[string]interface{}, error) {
	count, err := d.ragService.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection count: %w", err)
	}

	stats := map[string]interface{}{
		"total_documents": count,
		"chunk_size":      d.chunkSize,
		"overlap":         d.overlap,
		"timestamp":       time.Now().Unix(),
	}

	return stats, nil
}

// CalculateFileChecksum calculates SHA-256 checksum of a file.
func CalculateFileChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
