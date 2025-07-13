package view

import (
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type ServiceScale struct {
	service        *types.Service
	scaleAction    func(desiredCount int32)
	cancelAction   func()
	currentDesired int32
}

func NewServiceScale(service *types.Service) *ServiceScale {
	return &ServiceScale{
		service:        service,
		scaleAction:    func(int32) {},
		cancelAction:   func() {},
		currentDesired: service.DesiredCount,
	}
}

func (ss *ServiceScale) SetScaleAction(action func(desiredCount int32)) *ServiceScale {
	ss.scaleAction = action
	return ss
}

func (ss *ServiceScale) SetCancelAction(action func()) *ServiceScale {
	ss.cancelAction = action
	return ss
}

func (ss *ServiceScale) Render() tview.Primitive {
	form := tview.NewForm()

	// Current service info
	form.AddTextView("Service", aws.ToString(ss.service.ServiceName), 40, 1, true, false)
	form.AddTextView("Current Desired Task Count", strconv.Itoa(int(ss.service.DesiredCount)), 40, 1, true, false)
	form.AddTextView("Current Running Task Count", strconv.Itoa(int(ss.service.RunningCount)), 40, 1, true, false)
	form.AddTextView("Current Pending Task Count", strconv.Itoa(int(ss.service.PendingCount)), 40, 1, true, false)

	// Input field for new desired count
	desiredCountStr := strconv.Itoa(int(ss.service.DesiredCount))
	form.AddInputField("New Desired Task Count", desiredCountStr, 20, nil, func(text string) {
		if count, err := strconv.ParseInt(text, 10, 32); err == nil && count >= 0 {
			ss.currentDesired = int32(count)
		}
	})

	// Buttons
	form.AddButton("Scale Service", func() {
		ss.scaleAction(ss.currentDesired)
	})

	form.AddButton("Cancel", func() {
		ss.cancelAction()
	})

	form.SetBorder(true).SetTitle(" Scale Service ").SetTitleAlign(tview.AlignLeft)

	body := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(ss.header(), 3, 1, false).
		AddItem(nil, 1, 1, false). // Spacer
		AddItem(form, 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			ss.cancelAction()
		}
		return event
	})

	return ui.CreateLayout(body)
}

func (ss *ServiceScale) header() *tview.Flex {
	return ui.CreateHeader(
		"Scale Service",
		"Adjust the desired task count for "+aws.ToString(ss.service.ServiceName),
	)
}
