package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ServiceInsightsHandler(ctx context.Context, _ ...any) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)
	service := valueFromContext[*types.Service](ctx)

	// Get service insights
	insights, err := api.GetServiceInsights(ctx, service)
	if err != nil {
		return nil, err
	}

	return view.NewServiceInsights(cluster, service, insights).
		SetReloadAction(func() {
			app.Goto(ctx, ServiceInsightsHandler)
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ServiceDetailHandler)
		}), nil
}