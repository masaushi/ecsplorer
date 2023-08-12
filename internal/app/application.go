package app

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/masaushi/ecsplorer/internal/api"
)

type Handler func(ctx context.Context, option ...any) (Page, error)

type Page interface {
	Render() tview.Primitive
}

var (
	Version string

	app       *tview.Application
	pages     *tview.Pages
	awsConfig *aws.Config
)

func CreateApplication(ctx context.Context, version string) (start func(Handler) error, err error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	api.SetClient(cfg)

	Version = version
	awsConfig = &cfg
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

func Goto(ctx context.Context, handler Handler, option ...any) {
	page, err := handler(ctx, option...)
	if err != nil {
		ErrorModal(err)
		return
	}

	pages.AddAndSwitchToPage(pageName(page), page.Render(), true)
}

func Region() string {
	return awsConfig.Region
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

func pageName(page Page) string {
	t := reflect.TypeOf(page)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}
