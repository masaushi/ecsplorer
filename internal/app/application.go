package app

import (
	"context"
	"reflect"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/rivo/tview"
)

type Handler func(context.Context) (Page, error)

type Page interface {
	Render() tview.Primitive
}

var (
	Version string
)

var (
	app       *tview.Application
	pages     *tview.Pages
	ecs       *api.ECS
	awsConfig *aws.Config
)

func CreateApplication(ctx context.Context, version string) (start func(Handler) error, err error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	Version = version
	awsConfig = &cfg
	ecs = api.NewECS(cfg)
	app = tview.NewApplication()
	pages = tview.NewPages()

	app.
		SetRoot(pages, true).
		EnableMouse(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			//nolint:exhaustive
			switch event.Rune() {
			case 'q':
				app.Stop()
			case '?':
			}
			return event
		})

	return func(h Handler) error { Goto(ctx, h); return app.Run() }, nil
}

func Goto(ctx context.Context, handler Handler) {
	page, err := handler(ctx)
	if err != nil {
		ErrorModal(err)
		return
	}

	name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	pages.AddAndSwitchToPage(name, page.Render(), true)
}

func ECS() *api.ECS {
	return ecs
}

func Suspend(f func()) bool {
	return app.Suspend(f)
}

func ConfirmModal(text string, okFunc func()) {
	modal(text, []string{"Cancel", "Yes"}, func(buttonLabel string) {
		if buttonLabel == "Yes" {
			okFunc()
		}
	})
}

func ErrorModal(err error) {
	modal(err.Error(), []string{"Close"}, func(_ string) {})
}

func modal(text string, buttons []string, f func(buttonLabel string)) {
	modal := tview.NewModal().
		SetText(text).
		SetBackgroundColor(tcell.ColorDefault).
		AddButtons(buttons).
		SetDoneFunc(func(_ int, buttonLabel string) {
			f(buttonLabel)
			pages.RemovePage("modal")
		})

	pages.AddPage("modal", modal, true, true)
}

func Region() string {
	return awsConfig.Region
}
