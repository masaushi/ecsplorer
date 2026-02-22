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

// AITroubleshootHandler handles AI-powered troubleshooting for an ECS service.
func AITroubleshootHandler(ctx context.Context, _ ...any) (app.Page, error) {
	provider := app.AIProvider()
	if provider == nil {
		return nil, errAIDisabled
	}

	cluster := valueFromContext[*types.Cluster](ctx)
	service := valueFromContext[*types.Service](ctx)

	serviceName := aws.ToString(service.ServiceName)
	clusterName := aws.ToString(cluster.ClusterName)

	var diagnosticData strings.Builder

	// 1. Service configuration
	insights, err := api.GetServiceInsights(ctx, service)
	if err != nil {
		fmt.Fprintf(&diagnosticData, "Service Configuration: [error retrieving: %v]\n\n", err)
	} else {
		diagnosticData.WriteString("=== Service Configuration ===\n")
		diagnosticData.WriteString(formatServiceConfig(service, insights, clusterName))
		diagnosticData.WriteString("\n")
	}

	// 2. Recent logs (non-fatal error)
	logGroup := extractLogGroup(insights)
	if logGroup != "" {
		logs, err := api.GetRecentLogs(ctx, logGroup, "", 1*time.Hour, 200)
		if err != nil {
			fmt.Fprintf(&diagnosticData, "=== Recent Logs ===\n[error retrieving: %v]\n\n", err)
		} else {
			diagnosticData.WriteString("=== Recent Logs ===\n")
			for _, entry := range logs {
				fmt.Fprintf(&diagnosticData, "[%s] %s\n", entry.Timestamp.Format(time.RFC3339), entry.Message)
			}
			diagnosticData.WriteString("\n")
		}
	}

	// 3. Metrics (non-fatal error)
	metrics, err := api.GetECSMetrics(ctx, clusterName, serviceName, 3*time.Hour, 300)
	if err != nil {
		fmt.Fprintf(&diagnosticData, "=== Metrics ===\n[error retrieving: %v]\n\n", err)
	} else {
		diagnosticData.WriteString("=== Metrics (Last 3 Hours) ===\n")
		for _, m := range metrics {
			fmt.Fprintf(&diagnosticData, "%s:\n", m.MetricName)

			sort.Slice(m.DataPoints, func(i, j int) bool {
				return m.DataPoints[i].Timestamp.Before(m.DataPoints[j].Timestamp)
			})

			for _, dp := range m.DataPoints {
				fmt.Fprintf(&diagnosticData, "  %s: avg=%.2f%%, max=%.2f%%, min=%.2f%%\n",
					dp.Timestamp.Format("15:04"), dp.Average, dp.Maximum, dp.Minimum)
			}
		}
		diagnosticData.WriteString("\n")
	}

	// 4. Service events (already in service object)
	if len(service.Events) > 0 {
		diagnosticData.WriteString("=== Service Events (Recent) ===\n")
		eventCount := len(service.Events)
		if eventCount > 20 {
			eventCount = 20
		}
		for _, event := range service.Events[:eventCount] {
			fmt.Fprintf(&diagnosticData, "[%s] %s\n",
				aws.ToTime(event.CreatedAt).Format(time.RFC3339),
				aws.ToString(event.Message))
		}
		diagnosticData.WriteString("\n")
	}

	// Build prompt and analyze
	prompt := ai.BuildTroubleshootPrompt(serviceName, clusterName, diagnosticData.String())

	resp, err := provider.Analyze(ctx, ai.AnalysisRequest{
		Feature:   ai.FeatureTroubleshoot,
		Prompt:    prompt,
		MaxTokens: 8192,
	})
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	modelInfo := fmt.Sprintf("%s (input: %d, output: %d tokens)",
		resp.Model, resp.TokensUsed.InputTokens, resp.TokensUsed.OutputTokens)

	return view.NewAIAnalysis(
		"AI Troubleshoot",
		serviceName+" @ "+clusterName,
		resp.Content,
		modelInfo,
	).
		SetReloadAction(func() {
			app.GotoAsync(ctx, AITroubleshootHandler, "Troubleshooting...")
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ServiceDetailHandler)
		}), nil
}
