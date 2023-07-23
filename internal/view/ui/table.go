package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func CreateTable(
	header []string,
	rows [][]string,
	selectedFunc func(row, column int),
) *tview.Table {
	table := tview.NewTable()

	for i, col := range header {
		cell := tview.NewTableCell(col).
			SetExpansion(1).
			SetMaxWidth(1).
			SetSelectable(false).
			SetTextColor(tcell.ColorSkyblue)
		table.SetCell(0, i, cell)
	}

	for ri, row := range rows {
		for ci, col := range row {
			table.SetCell(ri+1, ci, tview.NewTableCell(col))
		}
	}

	table.SetBorder(true)
	table.SetBorderPadding(0, 0, 1, 1)
	table.Select(0, 0)
	table.SetBorders(false)
	table.SetFixed(1, 0)
	table.SetSelectable(true, false)
	table.SetSelectedFunc(selectedFunc)

	return table
}
