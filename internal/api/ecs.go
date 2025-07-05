// Package api provides AWS ECS client operations for the ecsplorer application.
// It handles cluster, service, and task management operations with proper error handling
// and context propagation.
package api

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/samber/lo"
)

const (
	// MaxServicesPerDescribeCall is the maximum number of services that can be described in a single API call
	MaxServicesPerDescribeCall = 10
)

// ECSService provides methods for interacting with AWS ECS resources.
type ECSService struct {
	client *ecs.Client
}

// ECSCreateExecuteSessionParams contains parameters for creating an ECS execute session.
type ECSCreateExecuteSessionParams struct {
	Cluster   *types.Cluster
	Task      *types.Task
	Container *types.Container
	Command   string
}

// NewECSService creates a new ECS service instance with the provided AWS configuration.
func NewECSService(config aws.Config) *ECSService {
	return &ECSService{
		client: ecs.NewFromConfig(config),
	}
}

// Global instance for backward compatibility
var defaultService *ECSService

// SetClient initializes the global ECS service instance.
// Deprecated: Use NewECSService instead for better testability.
func SetClient(config aws.Config) {
	defaultService = NewECSService(config)
}

// GetClusters retrieves all ECS clusters from the configured AWS account.
// It handles pagination automatically and returns an empty slice if no clusters exist.
// Returns an error if the AWS API call fails or if there are authentication issues.
func GetClusters(ctx context.Context) ([]types.Cluster, error) {
	return defaultService.GetClusters(ctx)
}

// GetClusters retrieves all ECS clusters from the configured AWS account.
func (s *ECSService) GetClusters(ctx context.Context) ([]types.Cluster, error) {
	if s.client == nil {
		return nil, fmt.Errorf("ECS client not initialized")
	}
	clusterARNs := make([]string, 0)
	var nextToken *string
	for {
		res, err := s.client.ListClusters(ctx, &ecs.ListClustersInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list clusters: %w", err)
		}

		clusterARNs = append(clusterARNs, res.ClusterArns...)
		nextToken = res.NextToken
		if nextToken == nil {
			break
		}
	}

	if len(clusterARNs) == 0 {
		return []types.Cluster{}, nil
	}

	res, err := s.client.DescribeClusters(ctx, &ecs.DescribeClustersInput{
		Clusters: clusterARNs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe clusters: %w", err)
	}

	return res.Clusters, nil
}

// GetServices retrieves all services for the specified cluster.
// It handles pagination and batching for large numbers of services.
// Returns an empty slice if no services exist for the cluster.
func GetServices(ctx context.Context, cluster *types.Cluster) ([]types.Service, error) {
	return defaultService.GetServices(ctx, cluster)
}

// GetServices retrieves all services for the specified cluster.
func (s *ECSService) GetServices(ctx context.Context, cluster *types.Cluster) ([]types.Service, error) {
	if s.client == nil {
		return nil, fmt.Errorf("ECS client not initialized")
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster cannot be nil")
	}
	serviceARNs := make([]string, 0)
	var nextToken *string
	for {
		res, err := s.client.ListServices(ctx, &ecs.ListServicesInput{
			Cluster:   cluster.ClusterArn,
			NextToken: nextToken,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list services for cluster %s: %w", aws.ToString(cluster.ClusterName), err)
		}

		serviceARNs = append(serviceARNs, res.ServiceArns...)
		nextToken = res.NextToken
		if nextToken == nil {
			break
		}
	}

	if len(serviceARNs) == 0 {
		return []types.Service{}, nil
	}

	var allServices []types.Service
	// DescribeServices can only return a limited number of services per call
	for _, arns := range lo.Chunk(serviceARNs, MaxServicesPerDescribeCall) {
		res, err := s.client.DescribeServices(ctx, &ecs.DescribeServicesInput{
			Cluster:  cluster.ClusterArn,
			Services: arns,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to describe services for cluster %s: %w", aws.ToString(cluster.ClusterName), err)
		}
		allServices = append(allServices, res.Services...)
	}

	return allServices, nil
}

// GetTasks retrieves all tasks for the specified cluster and optionally filtered by service.
// If service is nil, returns all tasks in the cluster.
// Returns an empty slice if no tasks exist.
func GetTasks(ctx context.Context, cluster *types.Cluster, service *types.Service) ([]types.Task, error) {
	return defaultService.GetTasks(ctx, cluster, service)
}

// GetTasks retrieves all tasks for the specified cluster and optionally filtered by service.
func (s *ECSService) GetTasks(ctx context.Context, cluster *types.Cluster, service *types.Service) ([]types.Task, error) {
	if s.client == nil {
		return nil, fmt.Errorf("ECS client not initialized")
	}
	if cluster == nil {
		return nil, fmt.Errorf("cluster cannot be nil")
	}
	var nextToken *string
	var serviceName *string
	if service != nil {
		serviceName = service.ServiceName
	}

	taskARNs := make([]string, 0)
	for {
		res, err := s.client.ListTasks(ctx, &ecs.ListTasksInput{
			Cluster:     cluster.ClusterArn,
			ServiceName: serviceName,
			NextToken:   nextToken,
		})
		if err != nil {
			if serviceName != nil {
				return nil, fmt.Errorf("failed to list tasks for service %s in cluster %s: %w", *serviceName, aws.ToString(cluster.ClusterName), err)
			}
			return nil, fmt.Errorf("failed to list tasks for cluster %s: %w", aws.ToString(cluster.ClusterName), err)
		}

		taskARNs = append(taskARNs, res.TaskArns...)
		nextToken = res.NextToken
		if nextToken == nil {
			break
		}
	}

	if len(taskARNs) == 0 {
		return []types.Task{}, nil
	}

	describeRes, err := s.client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: cluster.ClusterArn,
		Tasks:   taskARNs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to describe tasks for cluster %s: %w", aws.ToString(cluster.ClusterName), err)
	}

	return describeRes.Tasks, nil
}

// CreateExecuteSession creates an ECS execute session for the specified container.
// This enables shell access to running containers in ECS tasks.
func CreateExecuteSession(ctx context.Context, params *ECSCreateExecuteSessionParams) (*types.Session, error) {
	return defaultService.CreateExecuteSession(ctx, params)
}

// CreateExecuteSession creates an ECS execute session for the specified container.
func (s *ECSService) CreateExecuteSession(ctx context.Context, params *ECSCreateExecuteSessionParams) (*types.Session, error) {
	if s.client == nil {
		return nil, fmt.Errorf("ECS client not initialized")
	}
	if params == nil {
		return nil, fmt.Errorf("params cannot be nil")
	}
	if params.Cluster == nil {
		return nil, fmt.Errorf("cluster cannot be nil")
	}
	if params.Task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}
	if params.Container == nil {
		return nil, fmt.Errorf("container cannot be nil")
	}
	if params.Command == "" {
		return nil, fmt.Errorf("command cannot be empty")
	}

	res, err := s.client.ExecuteCommand(ctx, &ecs.ExecuteCommandInput{
		Command:     aws.String(params.Command),
		Interactive: true,
		Cluster:     params.Cluster.ClusterArn,
		Task:        params.Task.TaskArn,
		Container:   params.Container.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create execute session for container %s in task %s: %w", aws.ToString(params.Container.Name), aws.ToString(params.Task.TaskArn), err)
	}

	return res.Session, nil
}
