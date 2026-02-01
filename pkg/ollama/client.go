package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type GenerateRequest struct {
	Model   string  `json:"model"`
	Prompt  string  `json:"prompt"`
	System  string  `json:"system,omitempty"`
	Stream  bool    `json:"stream"`
	Options Options `json:"options,omitempty"`
}

type Options struct {
	Temperature   float64 `json:"temperature,omitempty"`
	NumPredict    int     `json:"num_predict,omitempty"`
	TopK          int     `json:"top_k,omitempty"`
	TopP          float64 `json:"top_p,omitempty"`
	RepeatPenalty float64 `json:"repeat_penalty,omitempty"`
	NumCtx        int     `json:"num_ctx,omitempty"`
}

type GenerateResponse struct {
	Response           string `json:"response"`
	Done               bool   `json:"done"`
	TotalDuration      int64  `json:"total_duration,omitempty"`       // nanoseconds
	LoadDuration       int64  `json:"load_duration,omitempty"`        // nanoseconds
	PromptEvalCount    int    `json:"prompt_eval_count,omitempty"`    // tokens
	PromptEvalDuration int64  `json:"prompt_eval_duration,omitempty"` // nanoseconds
	EvalCount          int    `json:"eval_count,omitempty"`           // tokens generated
	EvalDuration       int64  `json:"eval_duration,omitempty"`        // nanoseconds
}

type EmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type EmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// NewClient creates a new Ollama client with default timeout.
func NewClient(baseURL string) *Client {
	return NewClientWithTimeout(baseURL, 60*time.Second)
}

// NewClientWithTimeout creates a new Ollama client with custom timeout.
// The timeout should be longer than the expected generation time to allow
// for context deadline propagation from callers.
func NewClientWithTimeout(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	req.Stream = false

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for error details
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		// Truncate very long error responses to prevent log spam
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "... (truncated)"
		}
		return nil, fmt.Errorf("ollama generate request failed with status %d: %s", resp.StatusCode, bodyStr)
	}

	var genResp GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, err
	}

	return &genResp, nil
}

func (c *Client) Embed(ctx context.Context, model, text string) ([]float32, error) {
	req := EmbedRequest{
		Model: model,
		Input: text,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response body for error details
		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		// Truncate very long error responses to prevent log spam
		if len(bodyStr) > 500 {
			bodyStr = bodyStr[:500] + "... (truncated)"
		}
		return nil, fmt.Errorf("ollama embed request failed with status %d: %s", resp.StatusCode, bodyStr)
	}

	var embedResp EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, err
	}

	if len(embedResp.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return embedResp.Embeddings[0], nil
}
