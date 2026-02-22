package app

import (
	"context"
	"fmt"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/masaushi/ecsplorer/internal/ai"
	"github.com/masaushi/ecsplorer/internal/api"
)

type Handler func(ctx context.Context, option ...any) (Page, error)

type Page interface {
	Render() tview.Primitive
}

var (
	Version string
	Cmd     *string

	app        *tview.Application
	pages      *tview.Pages
	awsConfig  *aws.Config
	aiProvider ai.Provider
)

// AIProvider returns the configured AI provider, or nil if AI is disabled.
func AIProvider() ai.Provider {
	return aiProvider
}

func CreateApplication(ctx context.Context, version string, profile string, cmd *string, aiCfg ai.Config) (start func(Handler) error, err error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return nil, err
	}

	api.SetClient(cfg)
	api.SetLogsClient(cfg)
	api.SetCloudWatchClient(cfg)

	provider, err := ai.CreateProvider(aiCfg, cfg)
	if err != nil {
		return nil, err
	}
	aiProvider = provider

	Version = version
	Cmd = cmd
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

// GotoAsync executes a handler asynchronously with a loading indicator.
// Returns a cancel function that can be used to abort the async operation.
func GotoAsync(ctx context.Context, handler Handler, loadingMessage string, option ...any) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	// Show loading page
	loadingView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("\n\n\n[yellow]%s[white]\n\nPress [blue]Esc[white] to cancel", loadingMessage))

	loadingPage := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(loadingView, 0, 1, true)

	loadingPage.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyESC {
			cancel()
		}
		return event
	})

	pages.AddAndSwitchToPage("loading", loadingPage, true)

	go func() {
		page, err := handler(ctx, option...)

		app.QueueUpdateDraw(func() {
			pages.RemovePage("loading")
			if err != nil {
				if ctx.Err() != nil {
					// Context was cancelled, don't show error
					return
				}
				ErrorModal(err)
				return
			}
			pages.AddAndSwitchToPage(pageName(page), page.Render(), true)
		})
	}()

	return cancel
}

// ShowAIMenu displays a modal menu for selecting AI analysis features.
func ShowAIMenu(list *tview.List) {
	list.SetBorder(true).
		SetTitle(" AI Analysis ").
		SetBackgroundColor(tcell.ColorDefault)

	// Create a centered overlay
	overlay := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(list, 12, 1, true).
			AddItem(nil, 0, 1, false), 50, 1, true).
		AddItem(nil, 0, 1, false)

	list.SetDoneFunc(func() {
		pages.RemovePage("ai-menu")
	})

	pages.AddPage("ai-menu", overlay, true, true)
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

func InfoModal(title, message string) {
	modal(fmt.Sprintf("%s\n\n%s", title, message), []string{"OK"}, func(_ string) {})
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
