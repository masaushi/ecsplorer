package view

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/masaushi/ecsplorer/internal/api"
	"github.com/masaushi/ecsplorer/internal/view/ui"
	"github.com/rivo/tview"
)

type TaskInsights struct {
	task           *types.Task
	cluster        *types.Cluster
	insights       *api.TaskInsights
	reloadAction   func()
	prevPageAction func()
}

func NewTaskInsights(cluster *types.Cluster, task *types.Task, insights *api.TaskInsights) *TaskInsights {
	return &TaskInsights{
		task:           task,
		cluster:        cluster,
		insights:       insights,
		reloadAction:   func() {},
		prevPageAction: func() {},
	}
}

func (ti *TaskInsights) SetReloadAction(action func()) *TaskInsights {
	ti.reloadAction = action
	return ti
}

func (ti *TaskInsights) SetPrevPageAction(action func()) *TaskInsights {
	ti.prevPageAction = action
	return ti
}

func (ti *TaskInsights) Render() tview.Primitive {
	// Create main layout
	body := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add header
	body.AddItem(ti.header(), 3, 1, false)

	// Add content sections in a grid layout
	contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Left column
	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow)
	leftColumn.AddItem(ti.taskInfoSection(), 0, 1, false)
	leftColumn.AddItem(ti.networkSection(), 0, 1, false)

	// Right column
	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow)
	rightColumn.AddItem(ti.containersSection(), 0, 1, false)

	contentFlex.AddItem(leftColumn, 0, 1, false)
	contentFlex.AddItem(rightColumn, 0, 1, false)

	body.AddItem(contentFlex, 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyESC:
			ti.prevPageAction()
		case event.Rune() == 'r':
			ti.reloadAction()
		default:
		}
		return event
	})

	return ui.CreateLayout(body)
}

func (ti *TaskInsights) header() *tview.Flex {
	taskID := extractResourceName(aws.ToString(ti.task.TaskArn))
	title := "Task Insights: " + taskID
	subtitle := "Detailed task configuration and environment"
	return ui.CreateHeader(title, subtitle)
}

func extractResourceName(arn string) string {
	// Extract resource name from ARN
	// Example: arn:aws:ecs:region:account:cluster/my-cluster -> my-cluster
	parts := strings.Split(arn, "/")

	if len(parts) > 0 {
		return strings.TrimSpace(parts[len(parts)-1])
	}
	return arn
}

func (ti *TaskInsights) taskInfoSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Task Information ")

	var content strings.Builder

	// Basic task info
	fmt.Fprintf(&content, "[white]Task ARN:\n[blue]%s[white]\n\n", aws.ToString(ti.task.TaskArn))
	fmt.Fprintf(&content, "[white]Launch Type: [yellow]%s[white]\n", string(ti.task.LaunchType))
	fmt.Fprintf(&content, "[white]Platform Version: [yellow]%s[white]\n", aws.ToString(ti.task.PlatformVersion))
	// CPU Architecture field doesn't exist in task, get from task definition
	if ti.insights.TaskDefinition != nil && ti.insights.TaskDefinition.Cpu != nil {
		fmt.Fprintf(&content, "[white]Task CPU: [yellow]%s[white]\n", aws.ToString(ti.insights.TaskDefinition.Cpu))
	}

	if ti.task.AvailabilityZone != nil {
		fmt.Fprintf(&content, "[white]Availability Zone: [yellow]%s[white]\n", aws.ToString(ti.task.AvailabilityZone))
	}

	// Task Definition info
	if ti.insights.TaskDefinition != nil {
		td := ti.insights.TaskDefinition
		fmt.Fprintf(&content, "\n[white]Task Definition:\n")
		fmt.Fprintf(&content, "[white]Family: [blue]%s[white]\n", aws.ToString(td.Family))
		fmt.Fprintf(&content, "[white]Revision: [yellow]%d[white]\n", td.Revision)
		if td.Cpu != nil {
			fmt.Fprintf(&content, "[white]CPU: [yellow]%s[white]\n", aws.ToString(td.Cpu))
		}
		if td.Memory != nil {
			fmt.Fprintf(&content, "[white]Memory: [yellow]%s[white]\n", aws.ToString(td.Memory))
		}
	}

	textView.SetText(content.String())
	return textView
}

func (ti *TaskInsights) networkSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Network Configuration ")

	var content strings.Builder

	// Network interfaces and attachments
	if len(ti.insights.Attachments) > 0 {
		for i, attachment := range ti.insights.Attachments {
			if i > 0 {
				content.WriteString("\n")
			}
			fmt.Fprintf(&content, "[white]Attachment %d:\n", i+1)
			fmt.Fprintf(&content, "[white]Type: [yellow]%s[white]\n", aws.ToString(attachment.Type))
			fmt.Fprintf(&content, "[white]Status: [yellow]%s[white]\n", aws.ToString(attachment.Status))

			// Display attachment details
			for _, detail := range attachment.Details {
				if detail.Name != nil && detail.Value != nil {
					fmt.Fprintf(&content, "[blue]%s:[white] %s\n", aws.ToString(detail.Name), aws.ToString(detail.Value))
				}
			}
		}
	} else {
		content.WriteString("[gray]No network attachments[white]")
	}

	textView.SetText(content.String())
	return textView
}

func (ti *TaskInsights) containersSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Container Details ")

	var content strings.Builder

	if len(ti.insights.ContainerDetails) == 0 {
		content.WriteString("[gray]No container details available[white]")
	} else {
		for i, containerDetail := range ti.insights.ContainerDetails {
			if i > 0 {
				content.WriteString("\n")
			}

			container := containerDetail.Container
			definition := containerDetail.Definition

			fmt.Fprintf(&content, "[white]Container: [blue]%s[white]\n", aws.ToString(container.Name))
			fmt.Fprintf(&content, "[white]Status: [yellow]%s[white]\n", aws.ToString(container.LastStatus))
			fmt.Fprintf(&content, "[white]Health: [yellow]%s[white]\n", string(container.HealthStatus))

			// Resource usage
			if container.Cpu != nil {
				fmt.Fprintf(&content, "[white]CPU: [yellow]%s[white]\n", aws.ToString(container.Cpu))
			}
			if container.Memory != nil {
				fmt.Fprintf(&content, "[white]Memory: [yellow]%s[white]\n", aws.ToString(container.Memory))
			}

			// Image information
			if definition != nil {
				fmt.Fprintf(&content, "[white]Image: [blue]%s[white]\n", aws.ToString(definition.Image))

				// Port mappings
				if len(definition.PortMappings) > 0 {
					fmt.Fprintf(&content, "[white]Port Mappings:\n")
					for _, pm := range definition.PortMappings {
						protocol := string(pm.Protocol)
						if protocol == "" {
							protocol = "tcp"
						}
						fmt.Fprintf(&content, "• [yellow]%d:%d[white] (%s)\n",
							pm.HostPort, pm.ContainerPort, protocol)
					}
				}

				// Environment variables (first few)
				if len(definition.Environment) > 0 {
					fmt.Fprintf(&content, "[white]Environment Variables:\n")
					maxEnvVars := 3
					for j, env := range definition.Environment {
						if j >= maxEnvVars {
							fmt.Fprintf(&content, "• [gray]... and %d more[white]\n", len(definition.Environment)-maxEnvVars)
							break
						}
						fmt.Fprintf(&content, "• [blue]%s[white]=[yellow]%s[white]\n",
							aws.ToString(env.Name), aws.ToString(env.Value))
					}
				}

				// Mount points
				if len(definition.MountPoints) > 0 {
					fmt.Fprintf(&content, "[white]Mount Points:\n")
					for _, mp := range definition.MountPoints {
						readOnly := ""
						if mp.ReadOnly != nil && *mp.ReadOnly {
							readOnly = " (read-only)"
						}
						fmt.Fprintf(&content, "• [blue]%s[white] -> [yellow]%s[white]%s\n",
							aws.ToString(mp.SourceVolume), aws.ToString(mp.ContainerPath), readOnly)
					}
				}
			}

			// Network bindings
			if len(containerDetail.NetworkBindings) > 0 {
				fmt.Fprintf(&content, "[white]Network Bindings:\n")
				for _, nb := range containerDetail.NetworkBindings {
					protocol := string(nb.Protocol)
					if protocol == "" {
						protocol = "tcp"
					}
					fmt.Fprintf(&content, "• [yellow]%s:%d[white] -> [yellow]%d[white] (%s)\n",
						aws.ToString(nb.BindIP), nb.HostPort, nb.ContainerPort, protocol)
				}
			}
		}
	}

	textView.SetText(content.String())
	return textView
}
