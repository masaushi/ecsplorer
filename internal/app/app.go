package app

import (
	"context"
	"reflect"
	"runtime"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/rivo/tview"
)

var (
	app    *tview.Application
	pages  *tview.Pages
	ecsAPI *api.ECS
	cfg    *aws.Config // TODO refactor
)

type Handler func(context.Context, *api.ECS) Page

type Page interface {
	Render() tview.Primitive
}

func Goto(ctx context.Context, handler Handler) {
	name := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
	pages.AddAndSwitchToPage(name, handler(ctx, ecsAPI).Render(), true)
}

func Suspend(f func()) bool {
	return app.Suspend(f)
}

func Refresh() {
	// app.Draw()
	app.QueueUpdateDraw(func() {})
}

func Region() string {
	return cfg.Region
}
