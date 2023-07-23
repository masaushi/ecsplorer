package api

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

type ECS struct {
	client *ecs.Client
}

type ECSCreateExecuteSessionParams struct {
	Cluster   types.Cluster
	Task      types.Task
	Container types.Container
	Command   string
}

func NewECS(config aws.Config) *ECS {
	client := ecs.NewFromConfig(config)
	return &ECS{client: client}
}

func (e *ECS) GetClusters(ctx context.Context) ([]types.Cluster, error) {
	clusterARNs := make([]string, 0)
	var nextToken *string
	for {
		res, err := e.client.ListClusters(ctx, &ecs.ListClustersInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		clusterARNs = append(clusterARNs, res.ClusterArns...)
		nextToken = res.NextToken
		if nextToken == nil {
			break
		}
	}

	res, err := e.client.DescribeClusters(ctx, &ecs.DescribeClustersInput{
		Clusters: clusterARNs,
	})
	if err != nil {
		return nil, err
	}

	return res.Clusters, nil
}

func (e *ECS) GetServices(ctx context.Context, cluster types.Cluster) ([]types.Service, error) {
	serviceARNs := make([]string, 0)
	var nextToken *string
	for {
		res, err := e.client.ListServices(ctx, &ecs.ListServicesInput{
			Cluster:   cluster.ClusterArn,
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		serviceARNs = append(serviceARNs, res.ServiceArns...)
		nextToken = res.NextToken
		if nextToken == nil {
			break
		}
	}

	res, err := e.client.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  cluster.ClusterArn,
		Services: serviceARNs,
	})
	if err != nil {
		return nil, err
	}

	return res.Services, nil

}

func (e *ECS) GetTasks(ctx context.Context, cluster types.Cluster, service types.Service) ([]types.Task, error) {
	taskARNs := make([]string, 0)
	var nextToken *string
	for {
		res, err := e.client.ListTasks(ctx, &ecs.ListTasksInput{
			Cluster:     cluster.ClusterArn,
			ServiceName: service.ServiceName,
			NextToken:   nextToken,
		})
		if err != nil {
			return nil, err
		}

		taskARNs = append(taskARNs, res.TaskArns...)
		nextToken = res.NextToken
		if nextToken == nil {
			break
		}
	}

	describeRes, err := e.client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: cluster.ClusterArn,
		Tasks:   taskARNs,
	})
	if err != nil {
		return nil, err
	}

	return describeRes.Tasks, nil
}

func (e *ECS) CreateExecuteSession(ctx context.Context, params *ECSCreateExecuteSessionParams) (*types.Session, error) {
	res, err := e.client.ExecuteCommand(ctx, &ecs.ExecuteCommandInput{
		Command:     aws.String(params.Command),
		Interactive: true,
		Cluster:     params.Cluster.ClusterArn,
		Task:        params.Task.TaskArn,
		Container:   params.Container.Name,
	})
	if err != nil {
		return nil, err
	}

	return res.Session, nil
}
