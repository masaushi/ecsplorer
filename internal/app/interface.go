package app

import (
	"context"

	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/rivo/tview"
)

type Handler func(context.Context, Operator) (Page, error)

type Operator interface {
	Goto(context.Context, Handler)
	ECS() *api.ECS
	Suspend(func()) bool
	ConfirmModal(text string, okFunc func())
	ErrorModal(err error)
	Region() string
}

type Page interface {
	Render() tview.Primitive
}
