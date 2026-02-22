package ai

import "context"

// AnalysisFeature represents the type of AI analysis to perform.
type AnalysisFeature string

const (
	FeatureLogAnalysis  AnalysisFeature = "log_analysis"
	FeatureMetrics      AnalysisFeature = "metrics_analysis"
	FeatureConfigReview AnalysisFeature = "config_review"
	FeatureTroubleshoot AnalysisFeature = "troubleshoot"
)

// AnalysisRequest holds the parameters for an AI analysis request.
type AnalysisRequest struct {
	Feature   AnalysisFeature
	Prompt    string
	MaxTokens int
}

// AnalysisResponse holds the result of an AI analysis.
type AnalysisResponse struct {
	Content    string
	Model      string
	TokensUsed *TokenUsage
}

// TokenUsage tracks token consumption for an AI request.
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

// Provider is the abstraction interface for AI backends.
// Users can implement this interface to use any AI provider.
type Provider interface {
	Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error)
	Name() string
}
