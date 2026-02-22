package api

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
)

var cwClient *cloudwatch.Client

// SetCloudWatchClient initializes the CloudWatch Metrics client.
func SetCloudWatchClient(config aws.Config) {
	cwClient = cloudwatch.NewFromConfig(config)
}

// MetricDataPoint represents a single metric data point.
type MetricDataPoint struct {
	Timestamp time.Time
	Average   float64
	Maximum   float64
	Minimum   float64
}

// MetricResult holds the result of a metric query.
type MetricResult struct {
	MetricName string
	DataPoints []MetricDataPoint
	Unit       string
}

// GetECSMetrics retrieves CPU and Memory utilization metrics for an ECS service.
func GetECSMetrics(ctx context.Context, clusterName, serviceName string, duration time.Duration, period int32) ([]MetricResult, error) {
	endTime := time.Now()
	startTime := endTime.Add(-duration)

	metricNames := []string{"CPUUtilization", "MemoryUtilization"}
	results := make([]MetricResult, 0, len(metricNames))

	for _, metricName := range metricNames {
		input := &cloudwatch.GetMetricStatisticsInput{
			Namespace:  aws.String("AWS/ECS"),
			MetricName: aws.String(metricName),
			Dimensions: []cwtypes.Dimension{
				{Name: aws.String("ClusterName"), Value: aws.String(clusterName)},
				{Name: aws.String("ServiceName"), Value: aws.String(serviceName)},
			},
			StartTime:  aws.Time(startTime),
			EndTime:    aws.Time(endTime),
			Period:     aws.Int32(period),
			Statistics: []cwtypes.Statistic{cwtypes.StatisticAverage, cwtypes.StatisticMaximum, cwtypes.StatisticMinimum},
		}

		res, err := cwClient.GetMetricStatistics(ctx, input)
		if err != nil {
			return nil, err
		}

		dataPoints := make([]MetricDataPoint, len(res.Datapoints))
		var unit string
		for i, dp := range res.Datapoints {
			dataPoints[i] = MetricDataPoint{
				Timestamp: aws.ToTime(dp.Timestamp),
				Average:   aws.ToFloat64(dp.Average),
				Maximum:   aws.ToFloat64(dp.Maximum),
				Minimum:   aws.ToFloat64(dp.Minimum),
			}
			if unit == "" {
				unit = string(dp.Unit)
			}
		}

		results = append(results, MetricResult{
			MetricName: metricName,
			DataPoints: dataPoints,
			Unit:       unit,
		})
	}

	return results, nil
}
