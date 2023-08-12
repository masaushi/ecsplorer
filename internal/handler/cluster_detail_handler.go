package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ClusterDetailHandler(ctx context.Context) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)
	services, err := app.ECS().GetServices(ctx, cluster)
	if err != nil {
		return nil, err
	}

	tasks, err := app.ECS().GetTasks(ctx, cluster, nil)
	if err != nil {
		return nil, err
	}

	serviceListView := view.NewServiceList(services).
		SetServiceSelectAction(func(s *types.Service) {
			ctx := contextWithValue[*types.Service](ctx, s)
			app.Goto(ctx, ServiceDetailHandler)
		})

	taskListView := view.NewTaskList(tasks).
		SetTaskSelectAction(func(t *types.Task) {
			ctx := contextWithValue[*types.Task](ctx, t)
			app.Goto(ctx, TaskDetailHandler)
		})

	return view.NewClusterDetail(cluster).
		AddTab("Services", serviceListView).
		AddTab("Tasks", taskListView).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterListHandler)
		}), nil
}
