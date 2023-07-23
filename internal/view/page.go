package view

import "github.com/rivo/tview"

type Page interface {
	Render() tview.Primitive
}
