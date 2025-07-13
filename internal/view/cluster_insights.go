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

type ClusterInsights struct {
	cluster        *types.Cluster
	insights       *api.ClusterInsights
	reloadAction   func()
	prevPageAction func()
}

func NewClusterInsights(cluster *types.Cluster, insights *api.ClusterInsights) *ClusterInsights {
	return &ClusterInsights{
		cluster:        cluster,
		insights:       insights,
		reloadAction:   func() {},
		prevPageAction: func() {},
	}
}

func (ci *ClusterInsights) SetReloadAction(action func()) *ClusterInsights {
	ci.reloadAction = action
	return ci
}

func (ci *ClusterInsights) SetPrevPageAction(action func()) *ClusterInsights {
	ci.prevPageAction = action
	return ci
}

func (ci *ClusterInsights) Render() tview.Primitive {
	// Create main layout
	body := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add header
	body.AddItem(ci.header(), 3, 1, false)

	// Add content sections
	contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Left column - Configuration
	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow)
	leftColumn.AddItem(ci.configurationSection(), 0, 1, false)

	// Right column - Tags and Statistics
	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow)
	rightColumn.AddItem(ci.tagsSection(), 0, 1, false)
	rightColumn.AddItem(ci.statisticsSection(), 0, 1, false)

	contentFlex.AddItem(leftColumn, 0, 1, false)
	contentFlex.AddItem(rightColumn, 0, 1, false)

	body.AddItem(contentFlex, 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyESC:
			ci.prevPageAction()
		case event.Rune() == 'r':
			ci.reloadAction()
		default:
		}
		return event
	})

	return ui.CreateLayout(body)
}

func (ci *ClusterInsights) header() *tview.Flex {
	title := "Cluster Insights: " + aws.ToString(ci.cluster.ClusterName)
	subtitle := "Detailed cluster configuration and status"
	return ui.CreateHeader(title, subtitle)
}

func (ci *ClusterInsights) configurationSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Configuration ")

	var content strings.Builder

	// Container Insights
	insightsStatus := ci.insights.ContainerInsights
	insightsColor := "[red]"
	if status := strings.ToLower(insightsStatus); status == "enabled" || status == "enhanced" {
		insightsColor = "[green]"
	}
	fmt.Fprintf(&content, "[white]Container Insights: %s%s[white]\n\n", insightsColor, insightsStatus)

	// Basic cluster info
	fmt.Fprintf(&content, "[white]Cluster ARN:\n[blue]%s[white]\n\n", aws.ToString(ci.cluster.ClusterArn))
	fmt.Fprintf(&content, "[white]Status: [yellow]%s[white]\n", aws.ToString(ci.cluster.Status))
	fmt.Fprintf(&content, "[white]Registered Container Instances: [yellow]%d[white]\n", ci.cluster.RegisteredContainerInstancesCount)
	fmt.Fprintf(&content, "[white]Active Services: [yellow]%d[white]\n", ci.cluster.ActiveServicesCount)
	fmt.Fprintf(&content, "[white]Running Tasks: [yellow]%d[white]\n", ci.cluster.RunningTasksCount)
	fmt.Fprintf(&content, "[white]Pending Tasks: [yellow]%d[white]\n", ci.cluster.PendingTasksCount)

	// Capacity Providers
	if len(ci.insights.CapacityProviders) > 0 {
		fmt.Fprintf(&content, "\n[white]Capacity Providers:\n")
		for _, cp := range ci.insights.CapacityProviders {
			status := string(cp.Status)
			statusColor := "[yellow]"
			if status == "ACTIVE" {
				statusColor = "[green]"
			}
			fmt.Fprintf(&content, "• %s%s[white] (%s)\n", statusColor, aws.ToString(cp.Name), status)
		}
	}

	// Cluster Configuration
	if ci.insights.Configuration != nil {
		fmt.Fprintf(&content, "\n[white]Cluster Configuration:\n")

		// Execute Command Configuration
		if ci.insights.Configuration.ExecuteCommandConfiguration != nil {
			execConfig := ci.insights.Configuration.ExecuteCommandConfiguration
			fmt.Fprintf(&content, "• Execute Command:\n")

			if execConfig.Logging != "" {
				fmt.Fprintf(&content, "  - Logging: [yellow]%s[white]\n", string(execConfig.Logging))
			}

			if execConfig.KmsKeyId != nil {
				fmt.Fprintf(&content, "  - KMS Key: [blue]%s[white]\n", aws.ToString(execConfig.KmsKeyId))
			}

			if execConfig.LogConfiguration != nil {
				logConfig := execConfig.LogConfiguration
				fmt.Fprintf(&content, "  - Log Configuration:\n")

				if logConfig.CloudWatchLogGroupName != nil {
					fmt.Fprintf(&content, "    • CloudWatch Log Group: [blue]%s[white]\n", aws.ToString(logConfig.CloudWatchLogGroupName))
				}

				encryptionStatus := "disabled"
				if logConfig.CloudWatchEncryptionEnabled {
					encryptionStatus = "enabled"
				}
				fmt.Fprintf(&content, "    • CloudWatch Encryption: [yellow]%s[white]\n", encryptionStatus)

				if logConfig.S3BucketName != nil {
					fmt.Fprintf(&content, "    • S3 Bucket: [blue]%s[white]\n", aws.ToString(logConfig.S3BucketName))
				}

				if logConfig.S3KeyPrefix != nil {
					fmt.Fprintf(&content, "    • S3 Key Prefix: [blue]%s[white]\n", aws.ToString(logConfig.S3KeyPrefix))
				}

				s3EncryptionStatus := "disabled"
				if logConfig.S3EncryptionEnabled {
					s3EncryptionStatus = "enabled"
				}
				fmt.Fprintf(&content, "    • S3 Encryption: [yellow]%s[white]\n", s3EncryptionStatus)
			}
		}

		// Managed Storage Configuration
		if ci.insights.Configuration.ManagedStorageConfiguration != nil {
			managedStorage := ci.insights.Configuration.ManagedStorageConfiguration
			fmt.Fprintf(&content, "• Managed Storage:\n")

			if managedStorage.FargateEphemeralStorageKmsKeyId != nil {
				fmt.Fprintf(&content, "  - Fargate Ephemeral Storage KMS Key: [blue]%s[white]\n", aws.ToString(managedStorage.FargateEphemeralStorageKmsKeyId))
			}

			if managedStorage.KmsKeyId != nil {
				fmt.Fprintf(&content, "  - KMS Key: [blue]%s[white]\n", aws.ToString(managedStorage.KmsKeyId))
			}
		}
	}

	textView.SetText(content.String())
	return textView
}

func (ci *ClusterInsights) tagsSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Tags ")

	if len(ci.insights.Tags) == 0 {
		textView.SetText("[gray]No tags configured[white]")
	} else {
		var content strings.Builder
		for _, tag := range ci.insights.Tags {
			fmt.Fprintf(&content, "[blue]%s:[white] %s\n", aws.ToString(tag.Key), aws.ToString(tag.Value))
		}
		textView.SetText(content.String())
	}

	return textView
}

func (ci *ClusterInsights) statisticsSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Resource Statistics ")

	if len(ci.insights.Statistics) == 0 {
		textView.SetText("[gray]No additional statistics available[white]")
	} else {
		var content strings.Builder
		for _, stat := range ci.insights.Statistics {
			fmt.Fprintf(&content, "[blue]%s:[white] %s\n", aws.ToString(stat.Name), aws.ToString(stat.Value))
		}
		textView.SetText(content.String())
	}

	return textView
}
