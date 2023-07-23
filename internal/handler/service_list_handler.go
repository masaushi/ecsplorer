package handler

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ServiceListHandler(ctx context.Context, ecsAPI *api.ECS) view.Page {
	cluster := valueFromContext[types.Cluster](ctx)

	services, err := ecsAPI.GetServices(ctx, cluster)
	if err != nil {
		log.Fatal(err)
	}

	return view.NewServiceList(services).
		SetServiceSelectAction(func(service types.Service) {
			ctx := contextWithValue[types.Service](ctx, service)
			app.Goto(ctx, ServiceDetailHandler)
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterListHandler)
		})
}
