package awsqueries

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

// GetCloudWatchLogs returns the logs from a given log group and stream
func GetCloudWatchLogs(cfg aws.Config, logGroupName, logStreamName string, token *string) (*cloudwatchlogs.GetLogEventsOutput, *string, error) {
	client := cloudwatchlogs.NewFromConfig(cfg)

	input := &cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		NextToken:     token,
	}

	logEvents, err := client.GetLogEvents(context.Background(), input)
	return logEvents, logEvents.NextForwardToken, err
}
