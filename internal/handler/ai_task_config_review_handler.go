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

// AITaskConfigReviewHandler handles AI-powered configuration review for an ECS task.
func AITaskConfigReviewHandler(ctx context.Context, _ ...any) (app.Page, error) {
	provider := app.AIProvider()
	if provider == nil {
		return nil, errAIDisabled
	}

	task := valueFromContext[*types.Task](ctx)

	// Get task insights
	insights, err := api.GetTaskInsights(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to get task insights: %w", err)
	}

	// Format task configuration
	configData := formatTaskConfig(task, insights)

	taskID := aws.ToString(task.TaskArn)
	// Use just the task ID portion for display
	if parts := strings.Split(taskID, "/"); len(parts) > 0 {
		taskID = parts[len(parts)-1]
	}

	// Build prompt and analyze
	prompt := ai.BuildConfigReviewPrompt(taskID, "Task", configData)

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
		"AI Task Config Review",
		taskID,
		resp.Content,
		modelInfo,
	).
		SetReloadAction(func() {
			app.GotoAsync(ctx, AITaskConfigReviewHandler, "Reviewing task configuration...")
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, TaskDetailHandler)
		}), nil
}

func formatTaskConfig(task *types.Task, insights *api.TaskInsights) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Task ARN: %s\n", aws.ToString(task.TaskArn))
	fmt.Fprintf(&b, "Last Status: %s\n", aws.ToString(task.LastStatus))
	fmt.Fprintf(&b, "Desired Status: %s\n", aws.ToString(task.DesiredStatus))
	fmt.Fprintf(&b, "Health Status: %s\n", string(task.HealthStatus))
	fmt.Fprintf(&b, "Launch Type: %s\n", string(task.LaunchType))
	if task.Cpu != nil {
		fmt.Fprintf(&b, "CPU: %s\n", aws.ToString(task.Cpu))
	}
	if task.Memory != nil {
		fmt.Fprintf(&b, "Memory: %s\n", aws.ToString(task.Memory))
	}
	if task.PlatformVersion != nil {
		fmt.Fprintf(&b, "Platform Version: %s\n", aws.ToString(task.PlatformVersion))
	}

	// Task Definition
	if insights.TaskDefinition != nil {
		td := insights.TaskDefinition
		fmt.Fprintf(&b, "\nTask Definition:\n")
		fmt.Fprintf(&b, "  Family: %s\n", aws.ToString(td.Family))
		fmt.Fprintf(&b, "  Revision: %d\n", td.Revision)
		fmt.Fprintf(&b, "  Network Mode: %s\n", string(td.NetworkMode))
		if td.ExecutionRoleArn != nil {
			fmt.Fprintf(&b, "  Execution Role: %s\n", aws.ToString(td.ExecutionRoleArn))
		}
		if td.TaskRoleArn != nil {
			fmt.Fprintf(&b, "  Task Role: %s\n", aws.ToString(td.TaskRoleArn))
		}
	}

	// Container Details
	for _, cd := range insights.ContainerDetails {
		fmt.Fprintf(&b, "\nContainer: %s\n", aws.ToString(cd.Container.Name))
		fmt.Fprintf(&b, "  Status: %s\n", aws.ToString(cd.Container.LastStatus))
		fmt.Fprintf(&b, "  Health: %s\n", string(cd.Container.HealthStatus))
		if cd.Definition != nil {
			fmt.Fprintf(&b, "  Image: %s\n", aws.ToString(cd.Definition.Image))
			fmt.Fprintf(&b, "  Essential: %v\n", aws.ToBool(cd.Definition.Essential))
			if cd.Definition.Cpu != 0 {
				fmt.Fprintf(&b, "  CPU: %d\n", cd.Definition.Cpu)
			}
			if cd.Definition.Memory != nil {
				fmt.Fprintf(&b, "  Memory: %d\n", aws.ToInt32(cd.Definition.Memory))
			}
			if cd.Definition.HealthCheck != nil {
				fmt.Fprintf(&b, "  Health Check: %s\n", strings.Join(cd.Definition.HealthCheck.Command, " "))
			}
			if cd.Definition.LogConfiguration != nil {
				fmt.Fprintf(&b, "  Log Driver: %s\n", string(cd.Definition.LogConfiguration.LogDriver))
			}
		}
	}

	// Attachments
	if len(insights.Attachments) > 0 {
		fmt.Fprintf(&b, "\nAttachments:\n")
		for _, att := range insights.Attachments {
			fmt.Fprintf(&b, "  Type: %s, Status: %s\n", aws.ToString(att.Type), aws.ToString(att.Status))
			for _, detail := range att.Details {
				fmt.Fprintf(&b, "    %s: %s\n", aws.ToString(detail.Name), aws.ToString(detail.Value))
			}
		}
	}

	return b.String()
}
