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
	geminiDefaultModel = "gemini-2.5-flash"
	geminiBaseURL      = "https://generativelanguage.googleapis.com/v1beta/models"
)

var errGeminiAPI = errors.New("gemini API error")

// GeminiProvider implements Provider using the Google Gemini API.
type GeminiProvider struct {
	apiKey  string
	modelID string
	client  *http.Client
}

// NewGeminiProvider creates a new Gemini-backed AI provider.
func NewGeminiProvider(apiKey string, modelID string) *GeminiProvider {
	if modelID == "" {
		modelID = geminiDefaultModel
	}
	return &GeminiProvider{
		apiKey:  apiKey,
		modelID: modelID,
		client:  &http.Client{},
	}
}

func (g *GeminiProvider) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	body := geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: req.Prompt},
				},
			},
		},
		GenerationConfig: geminiGenerationConfig{
			MaxOutputTokens: maxTokens,
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling gemini request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", geminiBaseURL, g.modelID, g.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("creating gemini request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini API request failed: %w", err)
	}
	defer httpResp.Body.Close()

	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading gemini response: %w", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d: %s", errGeminiAPI, httpResp.StatusCode, string(respBody))
	}

	var result geminiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing gemini response: %w", err)
	}

	var content string
	if len(result.Candidates) > 0 {
		for _, part := range result.Candidates[0].Content.Parts {
			content += part.Text
		}
	}

	resp := &AnalysisResponse{
		Content: content,
		Model:   g.modelID,
	}

	if result.UsageMetadata.PromptTokenCount > 0 || result.UsageMetadata.CandidatesTokenCount > 0 {
		resp.TokensUsed = &TokenUsage{
			InputTokens:  result.UsageMetadata.PromptTokenCount,
			OutputTokens: result.UsageMetadata.CandidatesTokenCount,
		}
	}

	return resp, nil
}

func (g *GeminiProvider) Name() string {
	return "gemini"
}

// Gemini API request/response types.

type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenerationConfig struct {
	MaxOutputTokens int `json:"maxOutputTokens"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate    `json:"candidates"`
	UsageMetadata geminiUsageMetadata  `json:"usageMetadata"`
}

type geminiCandidate struct {
	Content geminiContent `json:"content"`
}

type geminiUsageMetadata struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
}
