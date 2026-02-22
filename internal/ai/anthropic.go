package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

const (
	anthropicDefaultModel = "claude-sonnet-4-20250514"
	anthropicBaseURL      = "https://api.anthropic.com/v1/messages"
	anthropicVersion      = "2023-06-01"
)

var errAnthropicAPI = errors.New("anthropic API error")

// AnthropicProvider implements Provider using the Anthropic Messages API.
type AnthropicProvider struct {
	apiKey  string
	modelID string
	client  *http.Client
}

// NewAnthropicProvider creates a new Anthropic-backed AI provider.
func NewAnthropicProvider(apiKey string, modelID string) *AnthropicProvider {
	if modelID == "" {
		modelID = anthropicDefaultModel
	}
	return &AnthropicProvider{
		apiKey:  apiKey,
		modelID: modelID,
		client:  &http.Client{},
	}
}

func (a *AnthropicProvider) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	body := anthropicRequest{
		Model:     a.modelID,
		MaxTokens: maxTokens,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling anthropic request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicBaseURL, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating anthropic request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", a.apiKey)
	httpReq.Header.Set("anthropic-version", anthropicVersion)

	httpResp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic API request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading anthropic response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s", errAnthropicAPI, httpResp.StatusCode, string(respBody))
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing anthropic response: %w", err)
	}

	var content string
	for _, block := range result.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	resp := &AnalysisResponse{
		Content: content,
		Model:   result.Model,
	}

	if result.Usage.InputTokens > 0 || result.Usage.OutputTokens > 0 {
		resp.TokensUsed = &TokenUsage{
			InputTokens:  result.Usage.InputTokens,
			OutputTokens: result.Usage.OutputTokens,
		}
	}

	return resp, nil
}

func (a *AnthropicProvider) Name() string {
	return "anthropic"
}

// Anthropic API request/response types.

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
	Model   string                 `json:"model"`
	Usage   anthropicUsage         `json:"usage"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}
