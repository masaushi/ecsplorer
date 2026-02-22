package handler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/rivo/tview"

	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

type serviceDetailHandlerOption struct {
	selectedTabIndex int
}

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

	var selectedTab int
	if len(options) > 0 {
		if option, ok := options[0].(*serviceDetailHandlerOption); ok {
			selectedTab = option.selectedTabIndex
		}
	}

	return view.NewServiceDetail(service, selectedTab).
		AddTab("Tasks", taskList).
		AddTab("Deployments", deploymentList).
		AddTab("Events", eventList).
		SetReloadAction(func(currentTab int) {
			app.Goto(ctx, ServiceDetailHandler, &serviceDetailHandlerOption{selectedTabIndex: currentTab})
		}).
		SetScaleAction(func() {
			app.Goto(ctx, ServiceScaleHandler)
		}).
		SetInsightsAction(func() {
			app.Goto(ctx, ServiceInsightsHandler)
		}).
		SetAIAction(func() {
			showServiceAIMenu(ctx)
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterDetailHandler)
		}), nil
}

func showServiceAIMenu(ctx context.Context) {
	if app.AIProvider() == nil {
		app.InfoModal("AI Disabled", "AI features are disabled. Use --ai=true to enable.")
		return
	}

	list := tview.NewList().
		AddItem("Log Analysis", "Analyze recent CloudWatch logs", 'l', func() {
			app.GotoAsync(ctx, AILogAnalysisHandler, "Analyzing logs...")
		}).
		AddItem("Metrics Analysis", "Analyze CPU/Memory metrics", 'm', func() {
			app.GotoAsync(ctx, AIMetricsAnalysisHandler, "Analyzing metrics...")
		}).
		AddItem("Config Review", "Review service configuration", 'c', func() {
			app.GotoAsync(ctx, AIConfigReviewHandler, "Reviewing configuration...")
		}).
		AddItem("Troubleshoot", "Comprehensive troubleshooting", 't', func() {
			app.GotoAsync(ctx, AITroubleshootHandler, "Troubleshooting...")
		})

	app.ShowAIMenu(list)
}
