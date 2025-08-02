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
	"github.com/samber/lo"
)

type ServiceInsights struct {
	service        *types.Service
	cluster        *types.Cluster
	insights       *api.ServiceInsights
	reloadAction   func()
	prevPageAction func()
}

func NewServiceInsights(cluster *types.Cluster, service *types.Service, insights *api.ServiceInsights) *ServiceInsights {
	return &ServiceInsights{
		service:        service,
		cluster:        cluster,
		insights:       insights,
		reloadAction:   func() {},
		prevPageAction: func() {},
	}
}

func (si *ServiceInsights) SetReloadAction(action func()) *ServiceInsights {
	si.reloadAction = action
	return si
}

func (si *ServiceInsights) SetPrevPageAction(action func()) *ServiceInsights {
	si.prevPageAction = action
	return si
}

func (si *ServiceInsights) Render() tview.Primitive {
	// Create main layout
	body := tview.NewFlex().SetDirection(tview.FlexRow)

	// Add header
	body.AddItem(si.header(), 3, 1, false)

	// Add content sections in a grid layout
	contentFlex := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Left column
	leftColumn := tview.NewFlex().SetDirection(tview.FlexRow)
	leftColumn.AddItem(si.taskDefinitionSection(), 0, 1, false)
	leftColumn.AddItem(si.networkSection(), 0, 1, false)

	// Right column
	rightColumn := tview.NewFlex().SetDirection(tview.FlexRow)
	rightColumn.AddItem(si.loadBalancerSection(), 0, 1, false)
	rightColumn.AddItem(si.placementSection(), 0, 1, false)
	rightColumn.AddItem(si.tagsSection(), 0, 1, false)

	contentFlex.AddItem(leftColumn, 0, 1, false)
	contentFlex.AddItem(rightColumn, 0, 1, false)

	body.AddItem(contentFlex, 0, 1, true)

	body.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch {
		case event.Key() == tcell.KeyESC:
			si.prevPageAction()
		case event.Rune() == 'r':
			si.reloadAction()
		default:
		}
		return event
	})

	return ui.CreateLayout(body)
}

func (si *ServiceInsights) header() *tview.Flex {
	title := "Service Insights: " + aws.ToString(si.service.ServiceName)
	subtitle := "Detailed service configuration and dependencies"
	return ui.CreateHeader(title, subtitle)
}

func (si *ServiceInsights) taskDefinitionSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Task Definition ")

	var content strings.Builder

	if si.insights.TaskDefinition == nil {
		content.WriteString("[gray]Task definition details not available[white]")
	} else {
		td := si.insights.TaskDefinition
		fmt.Fprintf(&content, "[white]Family: [blue]%s[white]\n", aws.ToString(td.Family))
		fmt.Fprintf(&content, "[white]Revision: [yellow]%d[white]\n", td.Revision)
		fmt.Fprintf(&content, "[white]Status: [yellow]%s[white]\n", string(td.Status))
		fmt.Fprintf(&content, "[white]Network Mode: [yellow]%s[white]\n", string(td.NetworkMode))
		fmt.Fprintf(&content, "[white]Requires Compatibilities: [yellow]%s[white]\n",
			strings.Join(lo.Map(td.RequiresCompatibilities, func(rc types.Compatibility, _ int) string {
				return string(rc)
			}), ", "))

		if td.Cpu != nil {
			fmt.Fprintf(&content, "[white]CPU: [yellow]%s[white]\n", aws.ToString(td.Cpu))
		}
		if td.Memory != nil {
			fmt.Fprintf(&content, "[white]Memory: [yellow]%s[white]\n", aws.ToString(td.Memory))
		}

		fmt.Fprintf(&content, "[white]Containers: [yellow]%d[white]\n", len(td.ContainerDefinitions))

		if td.ExecutionRoleArn != nil {
			fmt.Fprintf(&content, "[white]Execution Role: [blue]%s[white]\n", aws.ToString(td.ExecutionRoleArn))
		}
		if td.TaskRoleArn != nil {
			fmt.Fprintf(&content, "[white]Task Role: [blue]%s[white]\n", aws.ToString(td.TaskRoleArn))
		}
	}

	textView.SetText(content.String())
	return textView
}

func (si *ServiceInsights) networkSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Network Configuration ")

	var content strings.Builder

	if si.insights.NetworkConfig == nil {
		content.WriteString("[gray]No network configuration[white]")
	} else {
		nc := si.insights.NetworkConfig.AwsvpcConfiguration
		if nc != nil {
			fmt.Fprintf(&content, "[white]Assign Public IP: [yellow]%s[white]\n", string(nc.AssignPublicIp))

			if len(nc.Subnets) > 0 {
				fmt.Fprintf(&content, "[white]Subnets:\n")
				for _, subnet := range nc.Subnets {
					fmt.Fprintf(&content, "• [blue]%s[white]\n", subnet)
				}
			}

			if len(nc.SecurityGroups) > 0 {
				fmt.Fprintf(&content, "[white]Security Groups:\n")
				for _, sg := range nc.SecurityGroups {
					fmt.Fprintf(&content, "• [blue]%s[white]\n", sg)
				}
			}
		}
	}

	textView.SetText(content.String())
	return textView
}

func (si *ServiceInsights) loadBalancerSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Load Balancers ")

	var content strings.Builder

	if len(si.insights.LoadBalancers) == 0 {
		content.WriteString("[gray]No load balancers configured[white]")
	} else {
		for i, lb := range si.insights.LoadBalancers {
			if i > 0 {
				content.WriteString("\n")
			}
			fmt.Fprintf(&content, "[white]Target Group: [blue]%s[white]\n", aws.ToString(lb.TargetGroupArn))
			fmt.Fprintf(&content, "[white]Container: [yellow]%s[white]\n", aws.ToString(lb.ContainerName))
			fmt.Fprintf(&content, "[white]Port: [yellow]%d[white]\n", lb.ContainerPort)
			if lb.LoadBalancerName != nil {
				fmt.Fprintf(&content, "[white]Load Balancer: [blue]%s[white]\n", aws.ToString(lb.LoadBalancerName))
			}
		}
	}

	// Service Registries
	if len(si.insights.ServiceRegistries) > 0 {
		content.WriteString("\n[white]Service Discovery:\n")
		for _, sr := range si.insights.ServiceRegistries {
			fmt.Fprintf(&content, "• [blue]%s[white]", aws.ToString(sr.RegistryArn))
			if sr.ContainerName != nil {
				fmt.Fprintf(&content, " (%s:%d)", aws.ToString(sr.ContainerName), sr.ContainerPort)
			}
			content.WriteString("\n")
		}
	}

	textView.SetText(content.String())
	return textView
}

func (si *ServiceInsights) placementSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Placement Configuration ")

	var content strings.Builder

	// Placement Strategy
	if len(si.insights.PlacementStrategy) > 0 {
		content.WriteString("[white]Placement Strategy:\n")
		for _, ps := range si.insights.PlacementStrategy {
			fmt.Fprintf(&content, "• [yellow]%s[white]", string(ps.Type))
			if ps.Field != nil {
				fmt.Fprintf(&content, " ([blue]%s[white])", aws.ToString(ps.Field))
			}
			content.WriteString("\n")
		}
	}

	// Placement Constraints
	if len(si.insights.PlacementConstraints) > 0 {
		if len(si.insights.PlacementStrategy) > 0 {
			content.WriteString("\n")
		}
		content.WriteString("[white]Placement Constraints:\n")
		for _, pc := range si.insights.PlacementConstraints {
			fmt.Fprintf(&content, "• [yellow]%s[white]", string(pc.Type))
			if pc.Expression != nil {
				fmt.Fprintf(&content, ": [blue]%s[white]", aws.ToString(pc.Expression))
			}
			content.WriteString("\n")
		}
	}

	if len(si.insights.PlacementStrategy) == 0 && len(si.insights.PlacementConstraints) == 0 {
		content.WriteString("[gray]No placement configuration[white]")
	}

	textView.SetText(content.String())
	return textView
}

func (si *ServiceInsights) tagsSection() *tview.TextView {
	textView := tview.NewTextView()
	textView.SetDynamicColors(true).
		SetBorder(true).
		SetTitle(" Tags ")

	var content strings.Builder

	if len(si.insights.Tags) == 0 {
		content.WriteString("[gray]No tags configured[white]")
	} else {
		for _, tag := range si.insights.Tags {
			fmt.Fprintf(&content, "[blue]%s:[white] %s\n", aws.ToString(tag.Key), aws.ToString(tag.Value))
		}
	}

	textView.SetText(content.String())
	return textView
}
