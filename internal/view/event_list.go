package view

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type EventList struct {
	service *types.Service
}

func NewEventList(service *types.Service) *EventList {
	return &EventList{
		service: service,
	}
}

func (el *EventList) Render() tview.Primitive {
	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(el.table(), 0, 1, true)

	return page
}

func (el *EventList) table() *tview.Table {
	header := []string{"CREATED AT", "MESSAGE"}

	rows := make([][]string, len(el.service.Events))
	for i, event := range el.service.Events {
		rows[i] = make([]string, 0, len(header))
		rows[i] = append(rows[i],
			event.CreatedAt.Format(time.RFC3339),
			aws.ToString(event.Message),
		)
	}

	return ui.CreateTable(header, rows, func(row, column int) {})
}
