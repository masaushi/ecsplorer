package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ClusterDetailHandler(ctx context.Context, ecsAPI *api.ECS) view.Page {
	cluster := valueFromContext[types.Cluster](ctx)
	return view.NewClusterDetail(cluster).
		AddTab("Services", ServiceListHandler(ctx, ecsAPI)).
		AddTab("Tasks", TaskListHandler(ctx, ecsAPI))
}
