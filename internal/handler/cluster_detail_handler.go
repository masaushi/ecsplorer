package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ClusterDetailHandler(ctx context.Context, ecsAPI *api.ECS) app.Page {
	cluster := valueFromContext[types.Cluster](ctx)
	services, err := ecsAPI.GetServices(ctx, cluster)
	// TODO fix err overwritten
	tasks, err := ecsAPI.GetTasks(ctx, cluster, types.Service{})

	serviceListView := view.NewServiceList(services).
		SetServiceSelectAction(func(s types.Service) {
			ctx := contextWithValue[types.Service](ctx, s)
			app.Goto(ctx, ServiceDetailHandler)
		})

	taskListView := view.NewTaskList(tasks).
		SetTaskSelectAction(func(t types.Task) {
			ctx := contextWithValue[types.Task](ctx, t)
			app.Goto(ctx, TaskDetailHandler)
		})

	return view.NewClusterDetail(cluster, err).
		AddTab("Services", serviceListView).
		AddTab("Tasks", taskListView).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterListHandler)
		})
}
