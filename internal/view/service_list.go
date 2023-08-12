package view

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type ServiceList struct {
	services     []types.Service
	selectAction func(service *types.Service)
}

func NewServiceList(services []types.Service) *ServiceList {
	return &ServiceList{
		services:     services,
		selectAction: func(*types.Service) {},
	}
}

func (sl *ServiceList) SetSelectAction(action func(*types.Service)) *ServiceList {
	sl.selectAction = action
	return sl
}

func (sl *ServiceList) Render() tview.Primitive {
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(sl.table(), 0, 1, true)

	return page
}

func (sl *ServiceList) table() *tview.Table {
	header := []string{"NAME", "STATUS", "RUNNING TASKS"}

	rows := make([][]string, len(sl.services))
	for i, svc := range sl.services {
		rows[i] = make([]string, 0, len(header))
		rows[i] = append(rows[i],
			aws.ToString(svc.ServiceName),
			aws.ToString(svc.Status),
			fmt.Sprintf("%d tasks running", svc.RunningCount),
		)
	}

	return ui.CreateTable(header, rows, func(row, column int) {
		sl.selectAction(&sl.services[row-1])
	})
}
