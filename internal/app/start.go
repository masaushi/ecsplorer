package app

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/rivo/tview"
)

func Start(ctx context.Context, handler Handler) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return err
	}

	app := tview.NewApplication()
	pages := tview.NewPages()

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

	op := newOperator(app, pages, api.NewECS(cfg), &cfg)
	op.Goto(ctx, handler)

	return app.Run()
}
