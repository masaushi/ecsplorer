package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/app"
	"github.com/masaushi/ecsplorer/internal/view"
)

func TaskDetailHandler(ctx context.Context, ecsAPI *api.ECS) app.Page {
	cluster := valueFromContext[types.Cluster](ctx)
	task := valueFromContext[types.Task](ctx)

	return view.NewTaskDetail(task, nil).
		SetContainerSelectAction(func(container types.Container) error {
			// TODO: refactor
			execSess, err := ecsAPI.CreateExecuteSession(ctx, &api.ECSCreateExecuteSessionParams{
				Cluster:   cluster,
				Task:      task,
				Container: container,
				Command:   "/bin/sh",
			})
			if err != nil {
				return err
			}

			sess, err := json.Marshal(execSess)
			if err != nil {
				return err
			}

			target := fmt.Sprintf("ecs:%s_%s_%s",
				aws.ToString(cluster.ClusterArn),
				aws.ToString(task.TaskArn),
				aws.ToString(container.RuntimeId),
			)
			ssmTarget, err := json.Marshal(map[string]string{"Target": target})
			if err != nil {
				return err
			}

			app.Suspend(func() {
				cmd := exec.Command(
					"session-manager-plugin",
					string(sess),
					app.Region(),
					"StartSession",
					"",
					string(ssmTarget),
					"https://ssm.ap-northeast-1.amazonaws.com",
				)
				cmd.Stdout = os.Stdout
				cmd.Stdin = os.Stdin
				cmd.Stderr = os.Stderr

				sigchan := make(chan os.Signal, 1)
				signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
				cmd.Run()
			})

			return nil
		}).
		SetPrevPageAction(func() {
			// TODO: refactor
			app.Goto(ctx, ClusterDetailHandler)
		})
}
