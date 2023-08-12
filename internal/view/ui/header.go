package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func CreateHeader(primary, secondary string) *tview.Flex {
	p := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetTextColor(tcell.ColorYellow).
		SetText("[ " + primary + " ]")

	s := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("(" + secondary + ")")

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(p, 0, 1, false).
		AddItem(s, 0, 1, false)
}
