package ai

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	brtypes "github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	smithy "github.com/aws/smithy-go"
)

const defaultModel = "anthropic.claude-3-5-haiku-20241022-v1:0"

// BedrockProvider implements Provider using Amazon Bedrock.
type BedrockProvider struct {
	client  *bedrockruntime.Client
	modelID string
}

// NewBedrockProvider creates a new Bedrock-backed AI provider.
func NewBedrockProvider(cfg aws.Config, modelID string) *BedrockProvider {
	if modelID == "" {
		modelID = defaultModel
	}
	return &BedrockProvider{
		client:  bedrockruntime.NewFromConfig(cfg),
		modelID: modelID,
	}
}

func (b *BedrockProvider) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	maxTokens := req.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 4096
	}

	output, err := b.client.Converse(ctx, &bedrockruntime.ConverseInput{
		ModelId: aws.String(b.modelID),
		Messages: []brtypes.Message{
			{
				Role: brtypes.ConversationRoleUser,
				Content: []brtypes.ContentBlock{
					&brtypes.ContentBlockMemberText{Value: req.Prompt},
				},
			},
		},
		InferenceConfig: &brtypes.InferenceConfiguration{
			MaxTokens: aws.Int32(int32(maxTokens)), //nolint:gosec
		},
	})
	if err != nil {
		return nil, bedrockError(err, b.modelID)
	}

	var content string
	if msg, ok := output.Output.(*brtypes.ConverseOutputMemberMessage); ok {
		for _, block := range msg.Value.Content {
			if textBlock, ok := block.(*brtypes.ContentBlockMemberText); ok {
				content += textBlock.Value
			}
		}
	}

	resp := &AnalysisResponse{
		Content: content,
		Model:   b.modelID,
	}

	if output.Usage != nil {
		resp.TokensUsed = &TokenUsage{
			InputTokens:  int(aws.ToInt32(output.Usage.InputTokens)),
			OutputTokens: int(aws.ToInt32(output.Usage.OutputTokens)),
		}
	}

	return resp, nil
}

func (b *BedrockProvider) Name() string {
	return "bedrock"
}

func bedrockError(err error, modelID string) error {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		msg := apiErr.ErrorMessage()
		if strings.Contains(msg, "use case details") || strings.Contains(msg, "not been submitted") {
			return fmt.Errorf("model access not enabled for %s. "+
				"Enable model access in the AWS Bedrock console (Model access > Modify model access), "+
				"or specify a different model with --ai-model: %w", modelID, err)
		}
		if strings.Contains(msg, "not authorized") || strings.Contains(msg, "AccessDeniedException") {
			return fmt.Errorf("access denied for model %s. "+
				"Check IAM permissions for bedrock:InvokeModel: %w", modelID, err)
		}
	}
	return fmt.Errorf("bedrock Converse failed (model: %s): %w", modelID, err)
}
