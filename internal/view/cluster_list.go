package view

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type ClusterList struct {
	region              string
	clusters            []types.Cluster
	clusterSelectAction func(cluster types.Cluster)
	err                 error
}

func NewClusterList(region string, clusters []types.Cluster, err error) *ClusterList {
	return &ClusterList{
		region:              region,
		clusters:            clusters,
		err:                 err,
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
		AddItem(ui.CreateHeader("SELECT CLUSTER", cl.region), 2, 1, false).
		AddItem(cl.table(), 0, 1, true)

	return ui.CreateLayout(body, cl.err)
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
