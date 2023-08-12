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

type ClusterList struct {
	clusters     []types.Cluster
	selectAction func(cluster *types.Cluster)
	reloadAction func()
}

func NewClusterList(clusters []types.Cluster) *ClusterList {
	return &ClusterList{
		clusters:     clusters,
		selectAction: func(cluster *types.Cluster) {},
		reloadAction: func() {},
	}
}

func (cl *ClusterList) SetSelectAction(action func(cluster *types.Cluster)) *ClusterList {
	cl.selectAction = action
	return cl
}

func (cl *ClusterList) SetReloadAction(action func()) *ClusterList {
	cl.reloadAction = action
	return cl
}

func (cl *ClusterList) Render() tview.Primitive {
	body := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ui.CreateHeader("SELECT CLUSTER", app.Region()), 2, 1, false).
		AddItem(cl.table(), 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Rune() == 'r':
			cl.reloadAction()
		default:
		}

		return event
	})

	return ui.CreateLayout(body)
}

func (cl *ClusterList) table() *tview.Table {
	header := []string{"NAME", "STATUS", "ACTIVE SERVICES", "RUNNING TASKS"}

	rows := make([][]string, len(cl.clusters))
	for i, cluster := range cl.clusters {
		rows[i] = make([]string, 0, len(header))
		rows[i] = append(rows[i],
			aws.ToString(cluster.ClusterName),
			aws.ToString(cluster.Status),
			fmt.Sprintf("%d services running", cluster.ActiveServicesCount),
			fmt.Sprintf("%d tasks running", cluster.RunningTasksCount),
		)
	}

	return ui.CreateTable(header, rows, func(row, column int) {
		cl.selectAction(&cl.clusters[row-1])
	})
}
