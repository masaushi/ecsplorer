package view

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type ClusterDetail struct {
	cluster        *types.Cluster
	currentTab     int
	tabs           []*ui.Tab
	reloadAction   func(currentTab int)
	prevPageAction func()
	insightsAction func()
}

func NewClusterDetail(cluster *types.Cluster, currentTab int) *ClusterDetail {
	return &ClusterDetail{
		cluster:        cluster,
		currentTab:     currentTab,
		tabs:           make([]*ui.Tab, 0),
		reloadAction:   func(_ int) {},
		prevPageAction: func() {},
		insightsAction: func() {},
	}
}

func (cd *ClusterDetail) AddTab(title string, page app.Page) *ClusterDetail {
	cd.tabs = append(cd.tabs, &ui.Tab{
		Title:   title,
		Content: page.Render(),
	})
	return cd
}

func (cd *ClusterDetail) SetReloadAction(action func(currentTab int)) *ClusterDetail {
	cd.reloadAction = action
	return cd
}

func (cd *ClusterDetail) SetPrevPageAction(action func()) *ClusterDetail {
	cd.prevPageAction = action
	return cd
}

func (cd *ClusterDetail) SetInsightsAction(action func()) *ClusterDetail {
	cd.insightsAction = action
	return cd
}

func (cd *ClusterDetail) Render() tview.Primitive {
	tabPage := ui.CreateTabPage(cd.tabs, cd.currentTab)

	body := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cd.header(), 3, 1, false).
		AddItem(cd.description(), 3, 1, false).
		AddItem(tabPage.Page, 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//nolint:exhaustive
		switch {
		case event.Key() == tcell.KeyTab:
			cd.currentTab = tabPage.Next()
		case event.Key() == tcell.KeyBacktab:
			cd.currentTab = tabPage.Prev()
		case event.Key() == tcell.KeyESC:
			cd.prevPageAction()
		case event.Rune() == 'r':
			cd.reloadAction(cd.currentTab)
		case event.Rune() == 'i':
			cd.insightsAction()
		default:
		}
		return event
	})

	return ui.CreateLayout(body, ui.WithAdditionalCommands([]string{"i: cluster insights"}))
}

func (cd *ClusterDetail) header() *tview.Flex {
	return ui.CreateHeader(
		"CLUSTER: "+aws.ToString(cd.cluster.ClusterName),
		aws.ToString(cd.cluster.ClusterArn),
	)
}

func (cd *ClusterDetail) description() *tview.Flex {
	return tview.NewFlex().
		AddItem(ui.CreateDescription("Status", aws.ToString(cd.cluster.Status)), 0, 1, false).
		AddItem(ui.CreateDescription("Active Services", fmt.Sprintf("%d services", cd.cluster.ActiveServicesCount)), 0, 1, false).
		AddItem(ui.CreateDescription("Running Tasks", fmt.Sprintf("%d tasks", cd.cluster.RunningTasksCount)), 0, 1, false).
		AddItem(ui.CreateDescription("Pending Tasks", fmt.Sprintf("%d tasks", cd.cluster.PendingTasksCount)), 0, 1, false)
}
