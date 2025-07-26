package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

type clusterDetailHandlerOption struct {
	selectedTabIndex int
}

func ClusterDetailHandler(ctx context.Context, options ...any) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)
	services, err := api.GetServices(ctx, cluster)
	if err != nil {
		return nil, err
	}

	tasks, err := api.GetTasks(ctx, cluster, nil)
	if err != nil {
		return nil, err
	}

	serviceListView := view.NewServiceList(services).
		SetSelectAction(func(s *types.Service) {
			ctx := contextWithValue[*types.Service](ctx, s)
			app.Goto(ctx, ServiceDetailHandler)
		})

	taskListView := view.NewTaskList(tasks).
		SetSelectAction(func(t *types.Task) {
			ctx := contextWithValue[*types.Task](ctx, t)
			app.Goto(ctx, TaskDetailHandler)
		})

	var selectedTab int
	if len(options) > 0 {
		if option, ok := options[0].(*clusterDetailHandlerOption); ok {
			selectedTab = option.selectedTabIndex
		}
	}

	return view.NewClusterDetail(cluster, selectedTab).
		AddTab("Services", serviceListView).
		AddTab("Tasks", taskListView).
		SetReloadAction(func(currentTab int) {
			app.Goto(ctx, ClusterDetailHandler, &clusterDetailHandlerOption{selectedTabIndex: currentTab})
		}).
		SetInsightsAction(func() {
			app.Goto(ctx, ClusterInsightsHandler)
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterListHandler)
		}), nil
}
