package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/masaushi/ecsplorer/internal/ai"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

// AILogAnalysisHandler handles AI-powered log analysis for an ECS service.
func AILogAnalysisHandler(ctx context.Context, _ ...any) (app.Page, error) {
	provider := app.AIProvider()
	if provider == nil {
		return nil, errAIDisabled
	}

	cluster := valueFromContext[*types.Cluster](ctx)
	service := valueFromContext[*types.Service](ctx)

	// Get task definition for log configuration
	insights, err := api.GetServiceInsights(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("failed to get service insights: %w", err)
	}

	// Build task definition summary
	var taskDefSummary strings.Builder
	if insights.TaskDefinition != nil {
		td := insights.TaskDefinition
		fmt.Fprintf(&taskDefSummary, "Family: %s, Revision: %d\n", aws.ToString(td.Family), td.Revision)
		if td.Cpu != nil {
			fmt.Fprintf(&taskDefSummary, "CPU: %s, ", aws.ToString(td.Cpu))
		}
		if td.Memory != nil {
			fmt.Fprintf(&taskDefSummary, "Memory: %s\n", aws.ToString(td.Memory))
		}
		fmt.Fprintf(&taskDefSummary, "Containers: %d\n", len(td.ContainerDefinitions))
	}

	// Extract log group from task definition
	logGroup := extractLogGroup(insights)
	if logGroup == "" {
		return nil, errNoLogGroup
	}

	// Get recent logs
	logs, err := api.GetRecentLogs(ctx, logGroup, "", 1*time.Hour, 500)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	if len(logs) == 0 {
		return nil, fmt.Errorf("%w: %s", errNoLogsFound, logGroup)
	}

	// Format log entries
	var logEntries strings.Builder
	for _, entry := range logs {
		fmt.Fprintf(&logEntries, "[%s] %s\n", entry.Timestamp.Format(time.RFC3339), entry.Message)
	}

	// Build prompt and analyze
	prompt := ai.BuildLogAnalysisPrompt(
		aws.ToString(service.ServiceName),
		aws.ToString(cluster.ClusterName),
		logEntries.String(),
		taskDefSummary.String(),
	)

	resp, err := provider.Analyze(ctx, ai.AnalysisRequest{
		Feature: ai.FeatureLogAnalysis,
		Prompt:  prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	modelInfo := fmt.Sprintf("%s (input: %d, output: %d tokens)",
		resp.Model, resp.TokensUsed.InputTokens, resp.TokensUsed.OutputTokens)

	return view.NewAIAnalysis(
		"AI Log Analysis",
		aws.ToString(service.ServiceName)+" @ "+aws.ToString(cluster.ClusterName),
		resp.Content,
		modelInfo,
	).
		SetReloadAction(func() {
			app.GotoAsync(ctx, AILogAnalysisHandler, "Analyzing logs...")
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ServiceDetailHandler)
		}), nil
}

func extractLogGroup(insights *api.ServiceInsights) string {
	if insights.TaskDefinition == nil {
		return ""
	}
	for _, cd := range insights.TaskDefinition.ContainerDefinitions {
		if cd.LogConfiguration != nil && cd.LogConfiguration.Options != nil {
			if group, ok := cd.LogConfiguration.Options["awslogs-group"]; ok {
				return group
			}
		}
	}
	return ""
}
