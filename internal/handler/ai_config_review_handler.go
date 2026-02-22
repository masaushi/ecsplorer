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

// AIConfigReviewHandler handles AI-powered configuration review for an ECS service.
func AIConfigReviewHandler(ctx context.Context, _ ...any) (app.Page, error) {
	provider := app.AIProvider()
	if provider == nil {
		return nil, errAIDisabled
	}

	cluster := valueFromContext[*types.Cluster](ctx)
	service := valueFromContext[*types.Service](ctx)

	serviceName := aws.ToString(service.ServiceName)
	clusterName := aws.ToString(cluster.ClusterName)

	// Get service insights (reuses existing function, no new AWS API calls needed)
	insights, err := api.GetServiceInsights(ctx, service)
	if err != nil {
		return nil, fmt.Errorf("failed to get service insights: %w", err)
	}

	// Format configuration data
	configData := formatServiceConfig(service, insights, clusterName)

	// Build prompt and analyze
	prompt := ai.BuildConfigReviewPrompt(serviceName, "Service", configData)

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
		"AI Config Review",
		serviceName+" @ "+clusterName,
		resp.Content,
		modelInfo,
	).
		SetReloadAction(func() {
			app.GotoAsync(ctx, AIConfigReviewHandler, "Reviewing configuration...")
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ServiceDetailHandler)
		}), nil
}

func formatServiceConfig(service *types.Service, insights *api.ServiceInsights, clusterName string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Cluster: %s\n", clusterName)
	fmt.Fprintf(&b, "Service: %s\n", aws.ToString(service.ServiceName))
	fmt.Fprintf(&b, "Status: %s\n", aws.ToString(service.Status))
	fmt.Fprintf(&b, "Launch Type: %s\n", string(service.LaunchType))
	fmt.Fprintf(&b, "Desired Count: %d\n", service.DesiredCount)
	fmt.Fprintf(&b, "Running Count: %d\n", service.RunningCount)
	fmt.Fprintf(&b, "Pending Count: %d\n", service.PendingCount)

	if service.PlatformVersion != nil {
		fmt.Fprintf(&b, "Platform Version: %s\n", aws.ToString(service.PlatformVersion))
	}
	if service.HealthCheckGracePeriodSeconds != nil {
		fmt.Fprintf(&b, "Health Check Grace Period: %d seconds\n", aws.ToInt32(service.HealthCheckGracePeriodSeconds))
	}

	// Task Definition
	if insights.TaskDefinition != nil {
		td := insights.TaskDefinition
		fmt.Fprintf(&b, "\nTask Definition:\n")
		fmt.Fprintf(&b, "  Family: %s\n", aws.ToString(td.Family))
		fmt.Fprintf(&b, "  Revision: %d\n", td.Revision)
		fmt.Fprintf(&b, "  Network Mode: %s\n", string(td.NetworkMode))
		if td.Cpu != nil {
			fmt.Fprintf(&b, "  CPU: %s\n", aws.ToString(td.Cpu))
		}
		if td.Memory != nil {
			fmt.Fprintf(&b, "  Memory: %s\n", aws.ToString(td.Memory))
		}
		if td.ExecutionRoleArn != nil {
			fmt.Fprintf(&b, "  Execution Role: %s\n", aws.ToString(td.ExecutionRoleArn))
		}
		if td.TaskRoleArn != nil {
			fmt.Fprintf(&b, "  Task Role: %s\n", aws.ToString(td.TaskRoleArn))
		}

		for _, cd := range td.ContainerDefinitions {
			fmt.Fprintf(&b, "\n  Container: %s\n", aws.ToString(cd.Name))
			fmt.Fprintf(&b, "    Image: %s\n", aws.ToString(cd.Image))
			fmt.Fprintf(&b, "    Essential: %v\n", aws.ToBool(cd.Essential))
			if cd.Cpu != 0 {
				fmt.Fprintf(&b, "    CPU: %d\n", cd.Cpu)
			}
			if cd.Memory != nil {
				fmt.Fprintf(&b, "    Memory: %d\n", aws.ToInt32(cd.Memory))
			}
			if cd.MemoryReservation != nil {
				fmt.Fprintf(&b, "    Memory Reservation: %d\n", aws.ToInt32(cd.MemoryReservation))
			}
			if cd.HealthCheck != nil {
				fmt.Fprintf(&b, "    Health Check: %s\n", strings.Join(cd.HealthCheck.Command, " "))
			}
			if cd.LogConfiguration != nil {
				fmt.Fprintf(&b, "    Log Driver: %s\n", string(cd.LogConfiguration.LogDriver))
			}
		}
	}

	// Network Configuration
	if insights.NetworkConfig != nil && insights.NetworkConfig.AwsvpcConfiguration != nil {
		nc := insights.NetworkConfig.AwsvpcConfiguration
		fmt.Fprintf(&b, "\nNetwork Configuration:\n")
		fmt.Fprintf(&b, "  Assign Public IP: %s\n", string(nc.AssignPublicIp))
		fmt.Fprintf(&b, "  Subnets: %s\n", strings.Join(nc.Subnets, ", "))
		fmt.Fprintf(&b, "  Security Groups: %s\n", strings.Join(nc.SecurityGroups, ", "))
	}

	// Load Balancers
	if len(insights.LoadBalancers) > 0 {
		fmt.Fprintf(&b, "\nLoad Balancers:\n")
		for _, lb := range insights.LoadBalancers {
			fmt.Fprintf(&b, "  Target Group: %s\n", aws.ToString(lb.TargetGroupArn))
			fmt.Fprintf(&b, "  Container: %s:%d\n", aws.ToString(lb.ContainerName), lb.ContainerPort)
		}
	}

	// Placement
	if len(insights.PlacementStrategy) > 0 {
		fmt.Fprintf(&b, "\nPlacement Strategy:\n")
		for _, ps := range insights.PlacementStrategy {
			fmt.Fprintf(&b, "  Type: %s, Field: %s\n", string(ps.Type), aws.ToString(ps.Field))
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
