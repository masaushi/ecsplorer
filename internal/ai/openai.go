package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	openaiDefaultModel = "gpt-4o-mini"
	openaiBaseURL      = "https://api.openai.com/v1"
)

var errOpenAIAPI = errors.New("openai API error")

// OpenAIProvider implements Provider using the OpenAI Chat Completions API.
// It also supports OpenAI-compatible endpoints (Ollama, vLLM, LiteLLM, etc.)
// by setting a custom base URL.
type OpenAIProvider struct {
	apiKey  string
	modelID string
	baseURL string
	client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI-backed AI provider.
// If baseURL is empty, the official OpenAI API endpoint is used.
func NewOpenAIProvider(apiKey string, modelID string, baseURL string) *OpenAIProvider {
	if modelID == "" {
		modelID = openaiDefaultModel
	}
	if baseURL == "" {
		baseURL = openaiBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &OpenAIProvider{
		apiKey:  apiKey,
		modelID: modelID,
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (o *OpenAIProvider) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	body := openaiRequest{
		Model:     o.modelID,
		MaxTokens: maxTokens,
		Messages: []openaiMessage{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling openai request: %w", err)
	}

	url := o.baseURL + "/chat/completions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating openai request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if o.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+o.apiKey)
	}

	httpResp, err := o.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai API request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading openai response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s", errOpenAIAPI, httpResp.StatusCode, string(respBody))
	}

	var result openaiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing openai response: %w", err)
	}

	var content string
	if len(result.Choices) > 0 {
		content = result.Choices[0].Message.Content
	}

	resp := &AnalysisResponse{
		Content: content,
		Model:   result.Model,
	}

	if result.Usage.PromptTokens > 0 || result.Usage.CompletionTokens > 0 {
		resp.TokensUsed = &TokenUsage{
			InputTokens:  result.Usage.PromptTokens,
			OutputTokens: result.Usage.CompletionTokens,
		}
	}

	return resp, nil
}

func (o *OpenAIProvider) Name() string {
	if o.baseURL != openaiBaseURL {
		return "openai-compatible"
	}
	return "openai"
}

// OpenAI API request/response types.

type openaiRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	Messages  []openaiMessage `json:"messages"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiResponse struct {
	Choices []openaiChoice `json:"choices"`
	Model   string         `json:"model"`
	Usage   openaiUsage    `json:"usage"`
}

type openaiChoice struct {
	Message openaiMessage `json:"message"`
}

type openaiUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}
