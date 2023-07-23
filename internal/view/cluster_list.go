package view

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type ClusterList struct {
	clusters            []types.Cluster
	clusterSelectAction func(cluster types.Cluster)
}

func NewClusterList(clusters []types.Cluster) *ClusterList {
	return &ClusterList{
		clusters:            clusters,
		clusterSelectAction: func(cluster types.Cluster) {},
	}
}

func (cl *ClusterList) SetClusterSelectAction(action func(cluster types.Cluster)) *ClusterList {
	cl.clusterSelectAction = action
	return cl
}

func (cl *ClusterList) Render() tview.Primitive {
	body := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(cl.title(), 2, 1, false).
		AddItem(cl.table(), 0, 1, true)

	return ui.CreateLayout(body)
}

func (cl *ClusterList) title() *tview.TextView {
	return tview.NewTextView().
		SetText("SELECT CLUSTER").
		SetTextAlign(tview.AlignCenter)
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
		cl.clusterSelectAction(cl.clusters[row-1])
	})
}
