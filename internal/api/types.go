package api

import (
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// ECSCreateExecuteSessionParams holds parameters for creating an execute session in ECS.
type ECSCreateExecuteSessionParams struct {
	Cluster   *types.Cluster
	Task      *types.Task
	Container *types.Container
	Command   string
}

// ClusterInsights holds insights about an ECS cluster.
type ClusterInsights struct {
	ContainerInsights string
	Configuration     *types.ClusterConfiguration
	Tags              []types.Tag
	CapacityProviders []types.CapacityProvider
	Statistics        []types.KeyValuePair
}

// ServiceInsights holds insights about an ECS service.
type ServiceInsights struct {
	TaskDefinition       *types.TaskDefinition
	LoadBalancers        []types.LoadBalancer
	ServiceRegistries    []types.ServiceRegistry
	NetworkConfig        *types.NetworkConfiguration
	Tags                 []types.Tag
	PlacementStrategy    []types.PlacementStrategy
	PlacementConstraints []types.PlacementConstraint
}

// TaskInsights holds insights about an ECS task.
type TaskInsights struct {
	TaskDefinition    *types.TaskDefinition
	ContainerDetails  []ContainerDetail
	NetworkInterfaces []types.NetworkInterface
	Attachments       []types.Attachment
}

// ContainerDetail holds details about a container within a task.
type ContainerDetail struct {
	Container       types.Container
	Definition      *types.ContainerDefinition
	NetworkBindings []types.NetworkBinding
}
