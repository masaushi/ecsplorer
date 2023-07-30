package ui

import "github.com/rivo/tview"

func CreateHeader(primary, secondary string) *tview.Flex {
	p := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetDynamicColors(true).
		SetText("[::b]" + primary)

	s := tview.NewTextView().
		SetTextAlign(tview.AlignCenter).
		SetText("(" + secondary + ")")

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(p, 0, 1, false).
		AddItem(s, 0, 1, false)
}
