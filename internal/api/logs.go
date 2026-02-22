package api

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwltypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

var logsClient *cloudwatchlogs.Client

// SetLogsClient initializes the CloudWatch Logs client.
func SetLogsClient(config aws.Config) {
	logsClient = cloudwatchlogs.NewFromConfig(config)
}

// LogEntry represents a single log event.
type LogEntry struct {
	Timestamp time.Time
	Message   string
}

// GetRecentLogs retrieves recent log events from CloudWatch Logs.
func GetRecentLogs(ctx context.Context, logGroup string, logStreamPrefix string, duration time.Duration, limit int32) ([]LogEntry, error) {
	startTime := time.Now().Add(-duration).UnixMilli()

	input := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: aws.String(logGroup),
		StartTime:    aws.Int64(startTime),
		Limit:        aws.Int32(limit),
		Interleaved:  aws.Bool(true),
	}

	if logStreamPrefix != "" {
		input.LogStreamNamePrefix = aws.String(logStreamPrefix)
	}

	// Use FilterLogEvents with ordering
	input.FilterPattern = nil

	maxEntries := int(limit)
	var entries []LogEntry
	var nextToken *string

	for {
		input.NextToken = nextToken
		res, err := logsClient.FilterLogEvents(ctx, input)
		if err != nil {
			return nil, err
		}

		for _, event := range res.Events {
			entries = append(entries, LogEntry{
				Timestamp: time.UnixMilli(aws.ToInt64(event.Timestamp)),
				Message:   aws.ToString(event.Message),
			})
		}

		nextToken = res.NextToken
		if nextToken == nil || len(entries) >= maxEntries {
			break
		}
	}

	if len(entries) > maxEntries {
		entries = entries[:maxEntries]
	}

	return entries, nil
}

// GetLogGroups lists log groups matching a prefix.
func GetLogGroups(ctx context.Context, prefix string) ([]cwltypes.LogGroup, error) {
	input := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(prefix),
	}

	var groups []cwltypes.LogGroup
	var nextToken *string

	for {
		input.NextToken = nextToken
		res, err := logsClient.DescribeLogGroups(ctx, input)
		if err != nil {
			return nil, err
		}

		groups = append(groups, res.LogGroups...)

		nextToken = res.NextToken
		if nextToken == nil {
			break
		}
	}

	return groups, nil
}
