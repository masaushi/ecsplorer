package handler

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/masaushi/ecsplorer/internal/ai"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

// AIClusterConfigReviewHandler handles AI-powered configuration review for an ECS cluster.
func AIClusterConfigReviewHandler(ctx context.Context, _ ...any) (app.Page, error) {
	provider := app.AIProvider()
	if provider == nil {
		return nil, errAIDisabled
	}

	cluster := valueFromContext[*types.Cluster](ctx)
	clusterName := aws.ToString(cluster.ClusterName)

	// Get cluster insights
	insights, err := api.GetClusterInsights(ctx, cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster insights: %w", err)
	}

	// Format configuration data
	configData := formatClusterConfig(cluster, insights)

	// Build prompt and analyze
	prompt := ai.BuildConfigReviewPrompt(clusterName, "Cluster", configData)

	resp, err := provider.Analyze(ctx, ai.AnalysisRequest{
		Feature: ai.FeatureConfigReview,
		Prompt:  prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	modelInfo := fmt.Sprintf("%s (input: %d, output: %d tokens)",
		resp.Model, resp.TokensUsed.InputTokens, resp.TokensUsed.OutputTokens)

	return view.NewAIAnalysis(
		"AI Cluster Config Review",
		clusterName,
		resp.Content,
		modelInfo,
	).
		SetReloadAction(func() {
			app.GotoAsync(ctx, AIClusterConfigReviewHandler, "Reviewing cluster configuration...")
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterDetailHandler)
		}), nil
}

func formatClusterConfig(cluster *types.Cluster, insights *api.ClusterInsights) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Cluster: %s\n", aws.ToString(cluster.ClusterName))
	fmt.Fprintf(&b, "Status: %s\n", aws.ToString(cluster.Status))
	fmt.Fprintf(&b, "Active Services: %d\n", cluster.ActiveServicesCount)
	fmt.Fprintf(&b, "Running Tasks: %d\n", cluster.RunningTasksCount)
	fmt.Fprintf(&b, "Pending Tasks: %d\n", cluster.PendingTasksCount)
	fmt.Fprintf(&b, "Registered Container Instances: %d\n", cluster.RegisteredContainerInstancesCount)
	fmt.Fprintf(&b, "Container Insights: %s\n", insights.ContainerInsights)

	// Capacity Providers
	if len(insights.CapacityProviders) > 0 {
		fmt.Fprintf(&b, "\nCapacity Providers:\n")
		for _, cp := range insights.CapacityProviders {
			fmt.Fprintf(&b, "  Name: %s\n", aws.ToString(cp.Name))
			fmt.Fprintf(&b, "  Status: %s\n", string(cp.Status))
			if cp.AutoScalingGroupProvider != nil {
				fmt.Fprintf(&b, "  ASG ARN: %s\n", aws.ToString(cp.AutoScalingGroupProvider.AutoScalingGroupArn))
				if cp.AutoScalingGroupProvider.ManagedScaling != nil {
					ms := cp.AutoScalingGroupProvider.ManagedScaling
					fmt.Fprintf(&b, "  Managed Scaling: %s\n", string(ms.Status))
					if ms.TargetCapacity != nil {
						fmt.Fprintf(&b, "  Target Capacity: %d%%\n", aws.ToInt32(ms.TargetCapacity))
					}
				}
			}
		}
	}

	// Configuration
	if insights.Configuration != nil && insights.Configuration.ExecuteCommandConfiguration != nil {
		ecc := insights.Configuration.ExecuteCommandConfiguration
		fmt.Fprintf(&b, "\nExecute Command Configuration:\n")
		fmt.Fprintf(&b, "  Logging: %s\n", string(ecc.Logging))
		if ecc.KmsKeyId != nil {
			fmt.Fprintf(&b, "  KMS Key: %s\n", aws.ToString(ecc.KmsKeyId))
		}
	}

	// Statistics
	if len(insights.Statistics) > 0 {
		fmt.Fprintf(&b, "\nStatistics:\n")
		for _, stat := range insights.Statistics {
			fmt.Fprintf(&b, "  %s: %s\n", aws.ToString(stat.Name), aws.ToString(stat.Value))
		}
	}

	// Tags
	if len(insights.Tags) > 0 {
		fmt.Fprintf(&b, "\nTags:\n")
		for _, tag := range insights.Tags {
			fmt.Fprintf(&b, "  %s: %s\n", aws.ToString(tag.Key), aws.ToString(tag.Value))
		}
	}

	return b.String()
}
