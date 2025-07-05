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

// ecsSessionData holds the serialized session and target data for ECS session manager.
type ecsSessionData struct {
	sessionJSON []byte
	targetJSON  []byte
}

// createECSSessionData creates and marshals session data for ECS execute session.
func createECSSessionData(ctx context.Context, cluster *types.Cluster, task *types.Task, container *types.Container) (*ecsSessionData, error) {
	execSess, err := api.CreateExecuteSession(ctx, &api.ECSCreateExecuteSessionParams{
		Cluster:   cluster,
		Task:      task,
		Container: container,
		Command:   *app.Cmd,
	})
	if err != nil {
		return nil, err
	}

	sess, err := json.Marshal(execSess)
	if err != nil {
		return nil, err
	}

	target := buildECSTarget(cluster, task, container)
	ssmTarget, err := json.Marshal(map[string]string{"Target": target})
	if err != nil {
		return nil, err
	}

	return &ecsSessionData{
		sessionJSON: sess,
		targetJSON:  ssmTarget,
	}, nil
}

// executeSessionManagerPlugin executes the AWS session manager plugin with the provided session data.
func executeSessionManagerPlugin(data *ecsSessionData) error {
	//nolint:gosec
	cmd := exec.Command(
		"session-manager-plugin",
		string(data.sessionJSON),
		app.Region(),
		"StartSession",
		"",
		string(data.targetJSON),
		buildSSMEndpoint(),
	)

	setupCommandIO(cmd)

	return executeCommandWithSignalHandling(cmd)
}

// buildECSTarget constructs the ECS target string for session manager.
func buildECSTarget(cluster *types.Cluster, task *types.Task, container *types.Container) string {
	return fmt.Sprintf("ecs:%s_%s_%s",
		aws.ToString(cluster.ClusterArn),
		aws.ToString(task.TaskArn),
		aws.ToString(container.RuntimeId),
	)
}

// buildSSMEndpoint constructs the SSM endpoint URL for the current region.
func buildSSMEndpoint() string {
	return fmt.Sprintf("https://ssm.%s.amazonaws.com", app.Region())
}

// executeCommandWithSignalHandling starts a command with proper signal forwarding to handle Ctrl+C gracefully.
func executeCommandWithSignalHandling(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case sig := <-sigChan:
		return forwardSignalAndWait(cmd, sig, done)
	case err := <-done:
		return err
	}
}

// setupCommandIO configures the command to use standard input/output/error streams.
func setupCommandIO(cmd *exec.Cmd) {
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
}

// forwardSignalAndWait forwards a signal to the command's process group and waits for completion.
func forwardSignalAndWait(cmd *exec.Cmd, sig os.Signal, done <-chan error) error {
	if cmd.Process != nil {
		if err := syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal)); err != nil {
			return fmt.Errorf("failed to forward signal: %w", err)
		}
	}
	return <-done
}

// TaskDetailHandler displays detailed information about an ECS task and provides shell access to containers.
func TaskDetailHandler(ctx context.Context, _ ...any) (app.Page, error) {
	cluster := valueFromContext[*types.Cluster](ctx)
	task := valueFromContext[*types.Task](ctx)

	return view.NewTaskDetail(task).
		SetReloadAction(func() {
			app.Goto(ctx, TaskDetailHandler)
		}).
		SetSelectAction(func(container *types.Container) {
			app.ConfirmModal("Exec shell against the container?", func() {
				data, err := createECSSessionData(ctx, cluster, task, container)
				if err != nil {
					app.ErrorModal(err)
					return
				}

				app.Suspend(func() {
					err = executeSessionManagerPlugin(data)
				})

				if err != nil {
					app.ErrorModal(err)
				}
			})
		}).
		SetPrevPageAction(func() {
			app.Goto(ctx, ClusterDetailHandler)
		}), nil
}
