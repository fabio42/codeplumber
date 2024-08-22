package vcr

import (
	"fmt"
	"time"

	awsqueries "github.com/fabio42/codeplumber/aws"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
)

func (r *Recorder) ListCodePipelines(record string) (map[string]awsqueries.Pipeline, error) {
	pipelines := map[string]awsqueries.Pipeline{}
	time.Sleep(2 * time.Second)

	Load(fmt.Sprintf("%s/%s", r.RecordDir, record), &pipelines)
	return pipelines, nil
}

func (r *Recorder) GetCodeBuilds(record string) (map[string]awsqueries.CodebuildData, error) {
	codebuilds := map[string]awsqueries.CodebuildData{}
	err := Load(fmt.Sprintf("%s/%s", r.RecordDir, record), &codebuilds)
	return codebuilds, err
}

func (r *Recorder) GetCloudWatchLogs(record string) (*cloudwatchlogs.GetLogEventsOutput, error) {
	cwlData := &cloudwatchlogs.GetLogEventsOutput{}
	err := Load(fmt.Sprintf("%s/%s", r.RecordDir, record), &cwlData)
	return cwlData, err
}
