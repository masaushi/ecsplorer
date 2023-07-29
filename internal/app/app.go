package app

import (
	"context"
	"reflect"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/view"
	"github.com/rivo/tview"
)

var (
	app    *tview.Application
	pages  *tview.Pages
	ecsAPI *api.ECS
	cfg    *aws.Config // TODO refactor
)

func Start(ctx context.Context, handler Handler) error {
	defaultConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	cfg = &defaultConfig
	ecsAPI = api.NewECS(defaultConfig)
	app = tview.NewApplication()
	pages = tview.NewPages()

	app.
		SetRoot(pages, true).
		EnableMouse(true).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Rune() {
			case 'q':
				app.Stop()
			case '?':
			}
			return event
		})

	Goto(ctx, handler)
	return app.Run()
}

type Handler func(context.Context, *api.ECS) view.Page

func Goto(ctx context.Context, handler Handler) {
	name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	pages.AddAndSwitchToPage(name, handler(ctx, ecsAPI).Render(), true)
}

func Suspend(f func()) bool {
	return app.Suspend(f)
}

func Region() string {
	return cfg.Region
}
