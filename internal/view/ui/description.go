package ui

import (
	"fmt"

	"github.com/rivo/tview"
)

func CreateDescription(title, value string) *tview.Flex {
	titleText := tview.NewTextView().
		SetDynamicColors(true).
		SetText(fmt.Sprintf("[yellow::b][ %s ]", title))

	valueText := tview.NewTextView().
		SetText(value)

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(titleText, 1, 0, false).
		AddItem(valueText, 2, 0, false)
}
