package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func TaskInsightsHandler(ctx context.Context, _ ...any) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)
	task := valueFromContext[*types.Task](ctx)

	// Get task insights
	insights, err := api.GetTaskInsights(ctx, task)
	if err != nil {
		return nil, err
	}

	return view.NewTaskInsights(cluster, task, insights).
		SetReloadAction(func() {
			app.Goto(ctx, TaskInsightsHandler)
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, TaskDetailHandler)
		}), nil
}
