package handler

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ClusterListHandler(ctx context.Context, ecsAPI *api.ECS) view.Page {
	clusters, err := ecsAPI.GetClusters(ctx)
	if err != nil {
		log.Fatal(err)
	}

	return view.NewClusterList(clusters).
		SetClusterSelectAction(func(cluster types.Cluster) {
			ctx := contextWithValue[types.Cluster](ctx, cluster)
			app.Goto(ctx, ClusterDetailHandler)
		})
}
