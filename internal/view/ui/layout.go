package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func CreateLayout(body *tview.Flex, err error) (layout *tview.Flex) {
	errorMessage := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorRed)
	if err != nil {
		errorMessage.SetText(err.Error())
	}

	command := tview.NewTextView().
		SetText("▼ ▲ (j k): navigate, q: quit, esc: cancel, ?: help").
		SetTextColor(tcell.ColorSkyblue)
	version := tview.NewTextView().
		SetText("v0.0.1").
		SetTextColor(tcell.ColorYellow)

	footer := tview.NewFlex().
		AddItem(command, 0, 1, false).
		AddItem(version, 6, 1, false)

	return tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(body, 0, 1, true).
		AddItem(errorMessage, 2, 1, false).
		AddItem(footer, 1, 1, false)
}
