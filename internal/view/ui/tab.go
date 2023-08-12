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

type TabPage struct {
	Page *tview.Flex
	Next func() (currentTab int)
	Prev func() (currentTab int)
}

func CreateTabPage(tabs []*Tab, selected int) *TabPage {
	pages := tview.NewPages()

	info := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetTextAlign(tview.AlignCenter).
		SetWrap(false).
		SetHighlightedFunc(func(added, removed, remaining []string) {
			pages.SwitchToPage(added[0])
		})

	previousTab := func() int {
		tab, _ := strconv.Atoi(info.GetHighlights()[0])
		tab = (tab - 1 + len(tabs)) % len(tabs)
		info.Highlight(strconv.Itoa(tab)).ScrollToHighlight()
		return tab
	}
	nextTab := func() int {
		tab, _ := strconv.Atoi(info.GetHighlights()[0])
		tab = (tab + 1) % len(tabs)
		info.Highlight(strconv.Itoa(tab)).ScrollToHighlight()
		return tab
	}

	for index, tab := range tabs {
		pages.AddPage(strconv.Itoa(index), tab.Content, true, index == selected)
		fmt.Fprintf(info, `["%d"][skyblue::b] %s [white][""]  `, index, tab.Title)
	}

	info.Highlight(strconv.Itoa(selected)).ScrollToHighlight()

	page := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(info, 1, 1, false).
		AddItem(pages, 0, 1, true)

	return &TabPage{page, nextTab, previousTab}
}
