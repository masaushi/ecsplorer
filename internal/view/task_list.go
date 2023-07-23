package view

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type TaskList struct {
	tasks            []types.Task
	taskSelectAction func(types.Task)
	prevPageAction   func()
}

func NewTaskList(tasks []types.Task) *TaskList {
	return &TaskList{
		tasks:            tasks,
		taskSelectAction: func(types.Task) {},
		prevPageAction:   func() {},
	}
}

func (tl *TaskList) SetTaskSelectAction(action func(types.Task)) *TaskList {
	tl.taskSelectAction = action
	return tl
}

func (tl *TaskList) SetPrevPageAction(action func()) *TaskList {
	tl.prevPageAction = action
	return tl
}

func (tl *TaskList) Render() tview.Primitive {
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tl.table(), 0, 1, true)

	page.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyESC:
			tl.prevPageAction()
		}
		return event
	})

	return page
}

func (tl *TaskList) table() *tview.Table {
	header := []string{"TASK ARN", "VERSION", "CPU", "MEMORY", "HEALTH STATUS"}

	rows := make([][]string, len(tl.tasks))
	for i, task := range tl.tasks {
		rows[i] = make([]string, 0, len(header))
		rows[i] = append(rows[i],
			aws.ToString(task.TaskArn),
			strconv.FormatInt(task.Version, 10),
			aws.ToString(task.Cpu),
			aws.ToString(task.Memory),
			string(task.HealthStatus),
		)
	}

	return ui.CreateTable(header, rows, func(row, column int) {
		tl.taskSelectAction(tl.tasks[row-1])
	})
}
