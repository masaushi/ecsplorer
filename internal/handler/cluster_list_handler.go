package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ClusterListHandler(ctx context.Context, _ ...any) (app.Page, error) {
	clusters, err := api.GetClusters(ctx)
	if err != nil {
		return nil, err
	}

	return view.NewClusterList(clusters).
		SetSelectAction(func(cluster *types.Cluster) {
			ctx := contextWithValue[*types.Cluster](ctx, cluster)
			app.Goto(ctx, ClusterDetailHandler)
		}).
		SetReloadAction(func() {
			app.Goto(ctx, ClusterListHandler)
		}), nil
}
