package view

import (
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

// AIAnalysis is a view for displaying AI analysis results.
type AIAnalysis struct {
	title          string
	subtitle       string
	content        string
	modelInfo      string
	reloadAction   func()
	prevPageAction func()
}

// NewAIAnalysis creates a new AI analysis view.
func NewAIAnalysis(title, subtitle, content, modelInfo string) *AIAnalysis {
	return &AIAnalysis{
		title:          title,
		subtitle:       subtitle,
		content:        content,
		modelInfo:      modelInfo,
		reloadAction:   func() {},
		prevPageAction: func() {},
	}
}

// SetReloadAction sets the action to perform when the user presses 'r'.
func (a *AIAnalysis) SetReloadAction(action func()) *AIAnalysis {
	a.reloadAction = action
	return a
}

// SetPrevPageAction sets the action to perform when the user presses Escape.
func (a *AIAnalysis) SetPrevPageAction(action func()) *AIAnalysis {
	a.prevPageAction = action
	return a
}

// Render returns the tview primitive for this view.
func (a *AIAnalysis) Render() tview.Primitive {
	body := tview.NewFlex().SetDirection(tview.FlexRow)

	// Header
	body.AddItem(ui.CreateHeader(a.title, a.subtitle), 3, 1, false)

	// Content
	contentView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetWordWrap(true).
		SetText(a.content)
	contentView.SetBorder(true).
		SetTitle(" Analysis Result ")

	body.AddItem(contentView, 0, 1, true)

	// Model info footer
	if a.modelInfo != "" {
		modelView := tview.NewTextView().
			SetDynamicColors(true).
			SetText("[gray]Model: " + a.modelInfo + "[white]").
			SetTextAlign(tview.AlignRight)
		body.AddItem(modelView, 1, 0, false)
	}

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyESC:
			a.prevPageAction()
		case event.Rune() == 'r':
			a.reloadAction()
		default:
		}
		return event
	})

	return ui.CreateLayout(body, ui.WithAdditionalCommands([]string{"a: AI analysis"}))
}
