package view

import (
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type DeploymentList struct {
	service *types.Service
}

func NewDeploymentList(service *types.Service) *DeploymentList {
	return &DeploymentList{
		service: service,
	}
}

func (dl *DeploymentList) Render() tview.Primitive {
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(dl.table(), 0, 1, true)

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		}
		return event
	})

	return page
}

func (dl *DeploymentList) table() *tview.Table {
	header := []string{"STARTED AT", "STATUS", "RUNNING TASKS", "FAILED TASKS", "TASK DEF"}

	rows := make([][]string, len(dl.service.Deployments))
	for i, deploy := range dl.service.Deployments {
		rows[i] = make([]string, 0, len(header))
		rows[i] = append(rows[i],
			deploy.CreatedAt.Format(time.RFC3339),
			aws.ToString(deploy.Status),
			strconv.Itoa(int(deploy.RunningCount)),
			strconv.Itoa(int(deploy.FailedTasks)),
			aws.ToString(deploy.TaskDefinition),
		)
	}

	return ui.CreateTable(header, rows, func(row, column int) {})
}
