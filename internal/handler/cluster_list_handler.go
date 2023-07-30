package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ClusterListHandler(ctx context.Context, ecsAPI *api.ECS) app.Page {
	clusters, err := ecsAPI.GetClusters(ctx)

	return view.NewClusterList(app.Region(), clusters, err).
		SetClusterSelectAction(func(cluster types.Cluster) {
			ctx := contextWithValue[types.Cluster](ctx, cluster)
			app.Goto(ctx, ClusterDetailHandler)
		})
}
