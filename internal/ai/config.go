package ai

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
)

var (
	errUnknownProvider = errors.New("unknown AI provider")
	errAPIKeyRequired  = errors.New("API key is required. Use --ai-api-key or set the corresponding env var (GEMINI_API_KEY, ANTHROPIC_API_KEY, OPENAI_API_KEY)")
)

// Config holds AI feature configuration.
type Config struct {
	Enabled       bool
	ProviderName  string // "bedrock" (default), "gemini", "anthropic", "openai", "openai-compatible"
	ModelID       string
	BedrockRegion string // Override region for Bedrock (empty uses default)
	APIKey        string // API key for providers that require one
	BaseURL       string // Custom base URL for OpenAI-compatible providers
}

// CreateProvider creates an AI provider based on the configuration.
func CreateProvider(cfg Config, awsCfg aws.Config) (Provider, error) {
	if !cfg.Enabled {
		return nil, nil //nolint:nilnil
	}

	switch cfg.ProviderName {
	case "bedrock", "":
		bedrockCfg := awsCfg.Copy()
		if cfg.BedrockRegion != "" {
			bedrockCfg.Region = cfg.BedrockRegion
		}
		return NewBedrockProvider(bedrockCfg, cfg.ModelID), nil
	case "gemini":
		if cfg.APIKey == "" {
			return nil, errAPIKeyRequired
		}
		return NewGeminiProvider(cfg.APIKey, cfg.ModelID), nil
	case "anthropic":
		if cfg.APIKey == "" {
			return nil, errAPIKeyRequired
		}
		return NewAnthropicProvider(cfg.APIKey, cfg.ModelID), nil
	case "openai":
		if cfg.APIKey == "" {
			return nil, errAPIKeyRequired
		}
		return NewOpenAIProvider(cfg.APIKey, cfg.ModelID, ""), nil
	case "openai-compatible":
		return NewOpenAIProvider(cfg.APIKey, cfg.ModelID, cfg.BaseURL), nil
	default:
		return nil, fmt.Errorf("%w: %s", errUnknownProvider, cfg.ProviderName)
	}
}
