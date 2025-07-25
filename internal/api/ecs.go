package api

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/samber/lo"
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

	var allServices []types.Service
	// DescribeServices can only return 10 services per call
	for _, arns := range lo.Chunk(serviceARNs, 10) {
		res, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{
			Cluster:  cluster.ClusterArn,
			Services: arns,
		})
		if err != nil {
			return nil, err
		}
		allServices = append(allServices, res.Services...)
	}

	return allServices, nil
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

func UpdateServiceDesiredCount(ctx context.Context, cluster *types.Cluster, service *types.Service, desiredCount int32) (*types.Service, error) {
	res, err := client.UpdateService(ctx, &ecs.UpdateServiceInput{
		Cluster:      cluster.ClusterArn,
		Service:      service.ServiceArn,
		DesiredCount: aws.Int32(desiredCount),
	})
	if err != nil {
		return nil, err
	}

	return res.Service, nil
}

type ClusterInsights struct {
	ContainerInsights string
	Configuration     *types.ClusterConfiguration
	Tags              []types.Tag
	CapacityProviders []types.CapacityProvider
	Statistics        []types.KeyValuePair
}

func GetClusterInsights(ctx context.Context, cluster *types.Cluster) (*ClusterInsights, error) {
	clusterName := aws.ToString(cluster.ClusterName)

	// Get cluster details with tags
	describeRes, err := client.DescribeClusters(ctx, &ecs.DescribeClustersInput{
		Clusters: []string{clusterName},
		Include: []types.ClusterField{
			types.ClusterFieldAttachments,
			types.ClusterFieldConfigurations,
			types.ClusterFieldSettings,
			types.ClusterFieldStatistics,
			types.ClusterFieldTags,
		},
	})
	if err != nil {
		return nil, err
	}

	if len(describeRes.Clusters) == 0 {
		return &ClusterInsights{}, nil
	}

	clusterDetails := describeRes.Clusters[0]

	insights := &ClusterInsights{
		ContainerInsights: "DISABLED",
		Configuration:     clusterDetails.Configuration,
		Tags:              clusterDetails.Tags,
		Statistics:        clusterDetails.Statistics,
	}

	// Check Container Insights status
	for _, setting := range clusterDetails.Settings {
		if string(setting.Name) == "containerInsights" {
			insights.ContainerInsights = aws.ToString(setting.Value)
			break
		}
	}

	// Get capacity providers
	if len(clusterDetails.CapacityProviders) > 0 {
		cpRes, err := client.DescribeCapacityProviders(ctx, &ecs.DescribeCapacityProvidersInput{
			CapacityProviders: clusterDetails.CapacityProviders,
		})
		if err == nil {
			insights.CapacityProviders = cpRes.CapacityProviders
		}
	}

	return insights, nil
}
