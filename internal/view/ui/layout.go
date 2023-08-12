package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/rivo/tview"
)

func CreateLayout(body *tview.Flex) (layout *tview.Flex) {
	command := tview.NewTextView().
		SetText("▼ ▲ (j k): navigate, q: quit, esc: cancel, r: reload, ?: help").
		SetTextColor(tcell.ColorSkyblue)
	version := tview.NewTextView().
		SetText(app.Version).
		SetTextColor(tcell.ColorYellow)

	footer := tview.NewFlex().
		AddItem(command, 0, 1, false).
		AddItem(version, 7, 1, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(body, 0, 1, true).
		AddItem(footer, 1, 1, false)
}
