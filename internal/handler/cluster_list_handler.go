package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ClusterListHandler(ctx context.Context, operator app.Operator) (app.Page, error) {
	clusters, err := operator.ECS().GetClusters(ctx)
	if err != nil {
		return nil, err
	}

	return view.NewClusterList(operator.Region(), clusters).
		SetClusterSelectAction(func(cluster *types.Cluster) {
			ctx := contextWithValue[*types.Cluster](ctx, cluster)
			operator.Goto(ctx, ClusterDetailHandler)
		}), nil
}
