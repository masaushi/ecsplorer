package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func DeploymentListHandler(ctx context.Context, ecsAPI *api.ECS) view.Page {
	service := valueFromContext[types.Service](ctx)

	return view.NewDeploymentList(service).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterListHandler)
		})
}
