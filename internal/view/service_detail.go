package view

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type ServiceDetail struct {
	service        types.Service
	tabs           []*ui.Tab
	prevPageAction func()
}

func NewServiceDetail(service types.Service) *ServiceDetail {
	return &ServiceDetail{
		service:        service,
		tabs:           make([]*ui.Tab, 0),
		prevPageAction: func() {},
	}
}

func (sd *ServiceDetail) AddTab(title string, page Page) *ServiceDetail {
	sd.tabs = append(sd.tabs, &ui.Tab{
		Title:   title,
		Content: page.Render(),
	})
	return sd
}

func (sd *ServiceDetail) SetPrevPageAction(action func()) *ServiceDetail {
	sd.prevPageAction = action
	return sd
}

func (sd *ServiceDetail) Render() tview.Primitive {
	tab, nextTab, prevTab := ui.CreateTabPage(sd.tabs)

	body := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(sd.header(), 3, 1, false).
		AddItem(sd.description(), 3, 1, false).
		AddItem(tab, 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			nextTab()
		case tcell.KeyBacktab:
			prevTab()
		case tcell.KeyESC:
			sd.prevPageAction()
		}
		return event
	})

	return ui.CreateLayout(body)
}

func (sd *ServiceDetail) header() *tview.Flex {
	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText(*sd.service.ServiceName), 0, 1, false).
		AddItem(tview.NewTextView().SetTextAlign(tview.AlignCenter).SetText("("+*sd.service.ServiceArn+")"), 0, 1, false)
}

func (sd *ServiceDetail) description() *tview.Flex {
	return tview.NewFlex().
		AddItem(ui.CreateDescription("Status", *sd.service.Status), 0, 1, false).
		AddItem(ui.CreateDescription("Running Tasks", fmt.Sprintf("%d tasks", sd.service.RunningCount)), 0, 1, false).
		AddItem(ui.CreateDescription("Pending Tasks", fmt.Sprintf("%d tasks", sd.service.PendingCount)), 0, 1, false).
		AddItem(ui.CreateDescription("Healthcheck Grace Period", fmt.Sprintf("%d seconds", *sd.service.HealthCheckGracePeriodSeconds)), 0, 1, false)

}
