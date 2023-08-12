package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ClusterListHandler(ctx context.Context) (app.Page, error) {
	clusters, err := app.ECS().GetClusters(ctx)
	if err != nil {
		return nil, err
	}

	return view.NewClusterList(app.Region(), clusters).
		SetClusterSelectAction(func(cluster *types.Cluster) {
			ctx := contextWithValue[*types.Cluster](ctx, cluster)
			app.Goto(ctx, ClusterDetailHandler)
		}), nil
}
