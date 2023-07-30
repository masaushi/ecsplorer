package app

import (
	"context"
	"reflect"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/rivo/tview"
)

type operator struct {
	app    *tview.Application
	pages  *tview.Pages
	ecs    *api.ECS
	config *aws.Config // TODO refactor
}

func newOperator(
	app *tview.Application,
	pages *tview.Pages,
	ecs *api.ECS,
	config *aws.Config,
) *operator {
	return &operator{app, pages, ecs, config}
}

func (op *operator) Goto(ctx context.Context, handler Handler) {
	page, err := handler(ctx, op)
	if err != nil {
		op.ErrorModal(err)
		return
	}

	name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	op.pages.AddAndSwitchToPage(name, page.Render(), true)
}

func (op *operator) ECS() *api.ECS {
	return op.ecs
}

func (op *operator) Suspend(f func()) bool {
	return op.app.Suspend(f)
}

func (op *operator) ConfirmModal(text string, okFunc func()) {
	op.modal(text, []string{"Cancel", "Yes"}, func(buttonLabel string) {
		if buttonLabel == "Yes" {
			okFunc()
		}
	})
}

func (op *operator) ErrorModal(err error) {
	op.modal(err.Error(), []string{"Close"}, func(_ string) {})
}

func (op *operator) modal(text string, buttons []string, f func(buttonLabel string)) {
	modal := tview.NewModal().
		SetText(text).
		SetBackgroundColor(tcell.ColorDefault).
		AddButtons(buttons).
		SetDoneFunc(func(_ int, buttonLabel string) {
			f(buttonLabel)
			op.pages.RemovePage("modal")
		})

	op.pages.AddPage("modal", modal, true, true)
}

func (op *operator) Region() string {
	return op.config.Region
}
