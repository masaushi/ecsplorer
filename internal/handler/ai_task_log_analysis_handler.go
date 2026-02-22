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

// AITaskLogAnalysisHandler handles AI-powered log analysis for an ECS task.
func AITaskLogAnalysisHandler(ctx context.Context, _ ...any) (app.Page, error) {
	provider := app.AIProvider()
	if provider == nil {
		return nil, errAIDisabled
	}

	task := valueFromContext[*types.Task](ctx)
	cluster := valueFromContext[*types.Cluster](ctx)

	// Get task insights for log configuration
	insights, err := api.GetTaskInsights(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to get task insights: %w", err)
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
	}

	// Extract log group from task definition
	logGroup := extractTaskLogGroup(insights)
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

	taskID := aws.ToString(task.TaskArn)
	if parts := strings.Split(taskID, "/"); len(parts) > 0 {
		taskID = parts[len(parts)-1]
	}

	// Build prompt and analyze
	prompt := ai.BuildLogAnalysisPrompt(
		taskID,
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
		"AI Task Log Analysis",
		taskID,
		resp.Content,
		modelInfo,
	).
		SetReloadAction(func() {
			app.GotoAsync(ctx, AITaskLogAnalysisHandler, "Analyzing task logs...")
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, TaskDetailHandler)
		}), nil
}

func extractTaskLogGroup(insights *api.TaskInsights) string {
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
