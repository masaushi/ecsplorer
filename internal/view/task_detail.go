package view

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type TaskDetail struct {
	task                  types.Task
	containerSelectAction func(types.Container)
	prevPageAction        func()
}

func NewTaskDetail(task types.Task) *TaskDetail {
	return &TaskDetail{
		task:                  task,
		containerSelectAction: func(types.Container) {},
		prevPageAction:        func() {},
	}
}

func (td *TaskDetail) SetContainerSelectAction(action func(types.Container)) *TaskDetail {
	td.containerSelectAction = action
	return td
}

func (td *TaskDetail) SetPrevPageAction(action func()) *TaskDetail {
	td.prevPageAction = action
	return td
}

func (td *TaskDetail) Render() tview.Primitive {
	body := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(td.header(), 2, 1, false).
		AddItem(td.description(), 3, 1, false).
		AddItem(td.table(), 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			td.prevPageAction()
		}
		return event
	})

	return ui.CreateLayout(body)
}

func (td *TaskDetail) header() *tview.Flex {
	text := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("task"), 0, 1, false).
		AddItem(tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(aws.ToString(td.task.TaskArn)), 0, 1, false)

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(text, 0, 1, false)
}

func (td *TaskDetail) description() *tview.Flex {
	return tview.NewFlex().
		AddItem(ui.CreateDescription("Last Status", aws.ToString(td.task.LastStatus)), 0, 1, false).
		AddItem(ui.CreateDescription("Desired Status", aws.ToString(td.task.DesiredStatus)), 0, 1, false).
		AddItem(ui.CreateDescription("Health Status", string(td.task.HealthStatus)), 0, 1, false).
		AddItem(ui.CreateDescription("Started At", td.task.StartedAt.Format(time.RFC3339)), 0, 1, false)
}

func (td *TaskDetail) table() *tview.Table {
	header := []string{"NAME", "STATUS", "HEALTH STATUS", "CPU", "MEMORY"}

	rows := make([][]string, len(td.task.Containers))
	for i, container := range td.task.Containers {
		rows[i] = make([]string, 0, len(header))
		rows[i] = append(rows[i],
			aws.ToString(container.Name),
			aws.ToString(container.LastStatus),
			string(container.HealthStatus),
			aws.ToString(container.Cpu),
			aws.ToString(container.Memory),
		)
	}

	return ui.CreateTable(header, rows, func(row, column int) {
		td.containerSelectAction(td.task.Containers[row-1])
	})
}
