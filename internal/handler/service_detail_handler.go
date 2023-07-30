package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ServiceDetailHandler(ctx context.Context, ecsAPI *api.ECS) app.Page {
	cluster := valueFromContext[types.Cluster](ctx)
	service := valueFromContext[types.Service](ctx)

	tasks, err := ecsAPI.GetTasks(ctx, cluster, service)

	taskList := view.NewTaskList(tasks).
		SetTaskSelectAction(func(t types.Task) {
			ctx := contextWithValue[types.Task](ctx, t)
			app.Goto(ctx, TaskDetailHandler)
		})
	deploymentList := view.NewDeploymentList(service)
	eventList := view.NewEventList(service)

	return view.NewServiceDetail(service, err).
		AddTab("Tasks", taskList).
		AddTab("Deployments", deploymentList).
		AddTab("Events", eventList).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterDetailHandler)
		})
}
