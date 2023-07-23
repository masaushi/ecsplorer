package cmd

import (
	"context"
	"log"

	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/handler"
)

func Execute(args []string) {
	ctx := context.Background()
	if err := app.Start(ctx, handler.ClusterListHandler); err != nil {
		log.Fatal(err)
	}
}
