package handler

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/masaushi/ecsplorer/internal/ai"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

// AIMetricsAnalysisHandler handles AI-powered metrics analysis for an ECS service.
func AIMetricsAnalysisHandler(ctx context.Context, _ ...any) (app.Page, error) {
	provider := app.AIProvider()
	if provider == nil {
		return nil, errAIDisabled
	}

	cluster := valueFromContext[*types.Cluster](ctx)
	service := valueFromContext[*types.Service](ctx)

	clusterName := aws.ToString(cluster.ClusterName)
	serviceName := aws.ToString(service.ServiceName)

	// Get metrics (6 hours, 5 minute intervals)
	metrics, err := api.GetECSMetrics(ctx, clusterName, serviceName, 6*time.Hour, 300)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	// Format metrics data
	var metricsData strings.Builder
	for _, m := range metrics {
		fmt.Fprintf(&metricsData, "\n%s (%s):\n", m.MetricName, m.Unit)

		// Sort data points by timestamp
		sort.Slice(m.DataPoints, func(i, j int) bool {
			return m.DataPoints[i].Timestamp.Before(m.DataPoints[j].Timestamp)
		})

		for _, dp := range m.DataPoints {
			fmt.Fprintf(&metricsData, "  %s: avg=%.2f%%, max=%.2f%%, min=%.2f%%\n",
				dp.Timestamp.Format("15:04"), dp.Average, dp.Maximum, dp.Minimum)
		}
	}

	// Build service config summary
	var serviceConfig strings.Builder
	fmt.Fprintf(&serviceConfig, "Desired Count: %d\n", service.DesiredCount)
	fmt.Fprintf(&serviceConfig, "Running Count: %d\n", service.RunningCount)
	fmt.Fprintf(&serviceConfig, "Launch Type: %s\n", string(service.LaunchType))
	if service.TaskDefinition != nil {
		fmt.Fprintf(&serviceConfig, "Task Definition: %s\n", aws.ToString(service.TaskDefinition))
	}

	// Build prompt and analyze
	prompt := ai.BuildMetricsAnalysisPrompt(
		serviceName,
		clusterName,
		metricsData.String(),
		serviceConfig.String(),
	)

	resp, err := provider.Analyze(ctx, ai.AnalysisRequest{
		Feature: ai.FeatureMetrics,
		Prompt:  prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	modelInfo := fmt.Sprintf("%s (input: %d, output: %d tokens)",
		resp.Model, resp.TokensUsed.InputTokens, resp.TokensUsed.OutputTokens)

	return view.NewAIAnalysis(
		"AI Metrics Analysis",
		serviceName+" @ "+clusterName,
		resp.Content,
		modelInfo,
	).
		SetReloadAction(func() {
			app.GotoAsync(ctx, AIMetricsAnalysisHandler, "Analyzing metrics...")
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ServiceDetailHandler)
		}), nil
}
