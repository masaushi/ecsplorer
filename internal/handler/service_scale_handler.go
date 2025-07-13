package handler

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func ServiceScaleHandler(ctx context.Context, _ ...any) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)
	service := valueFromContext[*types.Service](ctx)

	return view.NewServiceScale(service).
		SetScaleAction(func(desiredCount int32) {
			if desiredCount == service.DesiredCount {
				app.InfoModal("No change required", fmt.Sprintf("Tasks are already at desired count: %d", desiredCount))
				return
			}

			app.ConfirmModal(
				fmt.Sprintf("Scale service from %d to %d tasks?", service.DesiredCount, desiredCount),
				func() {
					updatedService, err := api.UpdateServiceDesiredCount(ctx, cluster, service, desiredCount)
					if err != nil {
						app.ErrorModal(err)
						return
					}

					// Update the service in context and go back to service detail
					newCtx := contextWithValue(ctx, updatedService)
					app.Goto(newCtx, ServiceDetailHandler)
				},
			)
		}).
		SetCancelAction(func() {
			app.Goto(ctx, ServiceDetailHandler)
		}), nil
}
