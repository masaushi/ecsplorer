package api

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

var client *ecs.Client

type ECSCreateExecuteSessionParams struct {
	Cluster   *types.Cluster
	Task      *types.Task
	Container *types.Container
	Command   string
}

func SetClient(config aws.Config) {
	client = ecs.NewFromConfig(config)
}

func GetClusters(ctx context.Context) ([]types.Cluster, error) {
	clusterARNs := make([]string, 0)
	var nextToken *string
	for {
		res, err := client.ListClusters(ctx, &ecs.ListClustersInput{
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

	if len(clusterARNs) == 0 {
		return []types.Cluster{}, nil
	}

	res, err := client.DescribeClusters(ctx, &ecs.DescribeClustersInput{
		Clusters: clusterARNs,
	})
	if err != nil {
		return nil, err
	}

	return res.Clusters, nil
}

func GetServices(ctx context.Context, cluster *types.Cluster) ([]types.Service, error) {
	serviceARNs := make([]string, 0)
	var nextToken *string
	for {
		res, err := client.ListServices(ctx, &ecs.ListServicesInput{
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

	if len(serviceARNs) == 0 {
		return []types.Service{}, nil
	}

	res, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{
		Cluster:  cluster.ClusterArn,
		Services: serviceARNs,
	})
	if err != nil {
		return nil, err
	}

	return res.Services, nil
}

func GetTasks(ctx context.Context, cluster *types.Cluster, service *types.Service) ([]types.Task, error) {
	var nextToken *string
	var serviceName *string
	if service != nil {
		serviceName = service.ServiceName
	}

	taskARNs := make([]string, 0)
	for {
		res, err := client.ListTasks(ctx, &ecs.ListTasksInput{
			Cluster:     cluster.ClusterArn,
			ServiceName: serviceName,
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

	if len(taskARNs) == 0 {
		return []types.Task{}, nil
	}

	describeRes, err := client.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: cluster.ClusterArn,
		Tasks:   taskARNs,
	})
	if err != nil {
		return nil, err
	}

	return describeRes.Tasks, nil
}

func CreateExecuteSession(ctx context.Context, params *ECSCreateExecuteSessionParams) (*types.Session, error) {
	res, err := client.ExecuteCommand(ctx, &ecs.ExecuteCommandInput{
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
