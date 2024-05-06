package view

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type TaskList struct {
	tasks        []types.Task
	selectAction func(*types.Task)
}

func NewTaskList(tasks []types.Task) *TaskList {
	return &TaskList{
		tasks:        tasks,
		selectAction: func(*types.Task) {},
	}
}

func (tl *TaskList) SetSelectAction(action func(*types.Task)) *TaskList {
	tl.selectAction = action
	return tl
}

func (tl *TaskList) Render() tview.Primitive {
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tl.table(), 0, 1, true)

	return page
}

func (tl *TaskList) table() *tview.Table {
	header := []string{"TASK ARN", "CPU", "MEMORY", "HEALTH STATUS", "CREATED AT"}

	rows := make([][]string, len(tl.tasks))
	for i, task := range tl.tasks {
		rows[i] = make([]string, 0, len(header))
		rows[i] = append(rows[i],
			aws.ToString(task.TaskArn),
			aws.ToString(task.Cpu),
			aws.ToString(task.Memory),
			string(task.HealthStatus),
			task.CreatedAt.Format(time.RFC3339),
		)
	}

	return ui.CreateTable(header, rows, func(row, _ int) {
		tl.selectAction(&tl.tasks[row-1])
	})
}
