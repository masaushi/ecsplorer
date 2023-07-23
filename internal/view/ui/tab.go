package ui

import (
	"fmt"
	"strconv"

	"github.com/rivo/tview"
)

type Tab struct {
	Title   string
	Content tview.Primitive
}

func CreateTabPage(tabs []*Tab) (page *tview.Flex, nextAction, prevAction func()) {
	pages := tview.NewPages()

	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetTextAlign(tview.AlignCenter).
		SetWrap(false).
		SetHighlightedFunc(func(added, removed, remaining []string) {
			pages.SwitchToPage(added[0])
		})

	previousTab := func() {
		tab, _ := strconv.Atoi(info.GetHighlights()[0])
		tab = (tab - 1 + len(tabs)) % len(tabs)
		info.Highlight(strconv.Itoa(tab)).
			ScrollToHighlight()
	}
	nextTab := func() {
		tab, _ := strconv.Atoi(info.GetHighlights()[0])
		tab = (tab + 1) % len(tabs)
		info.Highlight(strconv.Itoa(tab)).
			ScrollToHighlight()
	}
	for index, tab := range tabs {
		pages.AddPage(strconv.Itoa(index), tab.Content, true, index == 0)
		fmt.Fprintf(info, `["%d"][skyblue::b] %s [white][""]  `, index, tab.Title)
	}
	info.Highlight("0")

	page = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(info, 1, 1, false).
		AddItem(pages, 0, 1, true)

	return page, nextTab, previousTab
}
