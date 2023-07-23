package view

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type ClusterDetail struct {
	cluster        types.Cluster
	tabs           []*ui.Tab
	prevPageAction func()
}

func NewClusterDetail(cluster types.Cluster) *ClusterDetail {
	return &ClusterDetail{
		cluster:        cluster,
		tabs:           make([]*ui.Tab, 0),
		prevPageAction: func() {},
	}
}

func (cd *ClusterDetail) AddTab(title string, page Page) *ClusterDetail {
	cd.tabs = append(cd.tabs, &ui.Tab{
		Title:   title,
		Content: page.Render(),
	})
	return cd
}

func (cd *ClusterDetail) SetPrevPageAction(action func()) *ClusterDetail {
	cd.prevPageAction = action
	return cd
}

func (cd *ClusterDetail) Render() tview.Primitive {
	tab, nextTab, prevTab := ui.CreateTabPage(cd.tabs)

	body := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cd.header(), 3, 1, false).
		AddItem(cd.description(), 3, 1, false).
		AddItem(tab, 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			nextTab()
		case tcell.KeyBacktab:
			prevTab()
		}
		return event
	})

	return ui.CreateLayout(body)
}

func (cd *ClusterDetail) header() *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetTextAlign(tview.AlignCenter).SetDynamicColors(true).SetText("[::b]"+aws.ToString(cd.cluster.ClusterName)), 0, 1, false).
		AddItem(tview.NewTextView().SetTextAlign(tview.AlignCenter).SetDynamicColors(true).SetText("("+aws.ToString(cd.cluster.ClusterArn)+")"), 0, 1, false)
}

func (cd *ClusterDetail) description() *tview.Flex {
	return tview.NewFlex().
		AddItem(ui.CreateDescription("Status", aws.ToString(cd.cluster.Status)), 0, 1, false).
		AddItem(ui.CreateDescription("Active Services", fmt.Sprintf("%d services", cd.cluster.ActiveServicesCount)), 0, 1, false).
		AddItem(ui.CreateDescription("Running Tasks", fmt.Sprintf("%d tasks", cd.cluster.RunningTasksCount)), 0, 1, false).
		AddItem(ui.CreateDescription("Pending Tasks", fmt.Sprintf("%d tasks", cd.cluster.PendingTasksCount)), 0, 1, false)
}
