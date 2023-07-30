package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ServiceDetailHandler(ctx context.Context, operator app.Operator) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)
	service := valueFromContext[*types.Service](ctx)

	tasks, err := operator.ECS().GetTasks(ctx, cluster, service)
	if err != nil {
		return nil, err
	}

	taskList := view.NewTaskList(tasks).
		SetTaskSelectAction(func(t *types.Task) {
			ctx := contextWithValue[*types.Task](ctx, t)
			operator.Goto(ctx, TaskDetailHandler)
		})
	deploymentList := view.NewDeploymentList(service)
	eventList := view.NewEventList(service)

	return view.NewServiceDetail(service).
		AddTab("Tasks", taskList).
		AddTab("Deployments", deploymentList).
		AddTab("Events", eventList).
		SetPrevPageAction(func() {
			operator.Goto(ctx, ClusterDetailHandler)
		}), nil
}
