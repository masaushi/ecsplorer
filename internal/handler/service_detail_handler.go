package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

// ServiceDetailHandler displays detailed information about an ECS service with tabs for tasks, deployments, and events.
func ServiceDetailHandler(ctx context.Context, options ...any) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)
	service := valueFromContext[*types.Service](ctx)

	tasks, err := api.GetTasks(ctx, cluster, service)
	if err != nil {
		return nil, err
	}

	taskList := view.NewTaskList(tasks).
		SetSelectAction(func(t *types.Task) {
			ctx := contextWithValue[*types.Task](ctx, t)
			app.Goto(ctx, TaskDetailHandler)
		})
	deploymentList := view.NewDeploymentList(service)
	eventList := view.NewEventList(service)

	selectedTab := parseTabOption(options)

	return view.NewServiceDetail(service, selectedTab).
		AddTab("Tasks", taskList).
		AddTab("Deployments", deploymentList).
		AddTab("Events", eventList).
		SetReloadAction(func(currentTab int) {
			app.Goto(ctx, ServiceDetailHandler, &TabOption{SelectedTabIndex: currentTab})
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterDetailHandler)
		}), nil
}
