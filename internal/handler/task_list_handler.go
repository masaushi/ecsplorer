package handler

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func TaskListHandler(ctx context.Context, ecsAPI *api.ECS) view.Page {
	cluster := valueFromContext[types.Cluster](ctx)
	service := valueFromContext[types.Service](ctx)

	tasks, err := ecsAPI.GetTasks(ctx, cluster, service)
	if err != nil {
		log.Fatal(err)
	}

	return view.NewTaskList(tasks).
		SetTaskSelectAction(func(task types.Task) {
			ctx := contextWithValue[types.Task](ctx, task)
			app.Goto(ctx, TaskDetailHandler)
		}).
		SetPrevPageAction(func() {
			// TODO: refactor
			if service.ServiceName != nil {
				app.Goto(ctx, ServiceDetailHandler)
				return
			}
			app.Goto(ctx, ClusterListHandler)
		})
}
