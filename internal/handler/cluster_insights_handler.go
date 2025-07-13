package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ClusterInsightsHandler(ctx context.Context, _ ...any) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)

	// Get cluster insights
	insights, err := api.GetClusterInsights(ctx, cluster)
	if err != nil {
		return nil, err
	}

	return view.NewClusterInsights(cluster, insights).
		SetReloadAction(func() {
			app.Goto(ctx, ClusterInsightsHandler)
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterDetailHandler)
		}), nil
}