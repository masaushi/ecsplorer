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

type ServiceDetail struct {
	service        *types.Service
	currentTab     int
	tabs           []*ui.Tab
	reloadAction   func(currentTab int)
	prevPageAction func()
}

func NewServiceDetail(service *types.Service, currentTab int) *ServiceDetail {
	return &ServiceDetail{
		service:        service,
		currentTab:     currentTab,
		reloadAction:   func(currentTab int) {},
		prevPageAction: func() {},
	}
}

func (sd *ServiceDetail) AddTab(title string, page app.Page) *ServiceDetail {
	sd.tabs = append(sd.tabs, &ui.Tab{
		Title:   title,
		Content: page.Render(),
	})
	return sd
}

func (sd *ServiceDetail) SetReloadAction(action func(currentTab int)) *ServiceDetail {
	sd.reloadAction = action
	return sd
}

func (sd *ServiceDetail) SetPrevPageAction(action func()) *ServiceDetail {
	sd.prevPageAction = action
	return sd
}

func (sd *ServiceDetail) Render() tview.Primitive {
	tabPage := ui.CreateTabPage(sd.tabs, sd.currentTab)

	body := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(sd.header(), 3, 1, false).
		AddItem(sd.description(), 3, 1, false).
		AddItem(tabPage.Page, 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//nolint:exhaustive
		switch {
		case event.Key() == tcell.KeyTab:
			sd.currentTab = tabPage.Next()
		case event.Key() == tcell.KeyBacktab:
			sd.currentTab = tabPage.Prev()
		case event.Key() == tcell.KeyESC:
			sd.prevPageAction()
		case event.Rune() == 'r':
			sd.reloadAction(sd.currentTab)
		default:
		}

		return event
	})

	return ui.CreateLayout(body)
}

func (sd *ServiceDetail) header() *tview.Flex {
	return ui.CreateHeader(
		"SERVICE: "+aws.ToString(sd.service.ServiceName),
		aws.ToString(sd.service.ServiceArn),
	)
}

func (sd *ServiceDetail) description() *tview.Flex {
	return tview.NewFlex().
		AddItem(ui.CreateDescription("Status", *sd.service.Status), 0, 1, false).
		AddItem(ui.CreateDescription("Running Tasks", fmt.Sprintf("%d tasks", sd.service.RunningCount)), 0, 1, false).
		AddItem(ui.CreateDescription("Pending Tasks", fmt.Sprintf("%d tasks", sd.service.PendingCount)), 0, 1, false).
		AddItem(ui.CreateDescription("Healthcheck Grace Period", fmt.Sprintf("%d seconds", *sd.service.HealthCheckGracePeriodSeconds)), 0, 1, false)
}
