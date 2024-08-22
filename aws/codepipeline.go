package awsqueries

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	"github.com/rs/zerolog/log"
)

// Pipeline is a struct to hold the data of a AWS CodePipeline
type Pipeline struct {
	PipelineName        string
	LastExecutionID     string
	LastExecutionStatus string
	Data                *codepipeline.GetPipelineOutput
	ExecData            *codepipeline.ListPipelineExecutionsOutput
	StateData           *codepipeline.GetPipelineStateOutput
	Tags                map[string]string
}

// CodePipelinesListFiltered is a function that returns a list of AWS CodePipeLine filtered by name and tags
func CodePipelinesListFiltered(cfg aws.Config, accountID, pattern string, tags map[string]string) (map[string]Pipeline, error) {
	client := codepipeline.NewFromConfig(cfg)
	pipelines := map[string]Pipeline{}

	totalPipelines, err := getTotalNumberOfPipelines(cfg)
	if err != nil {
		return nil, err
	}

	params := &codepipeline.ListPipelinesInput{}
	paginator := codepipeline.NewListPipelinesPaginator(client, params)

	resultChan := make(chan Pipeline, totalPipelines)
	var wg sync.WaitGroup

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			return nil, err
		}

		for _, pipeline := range page.Pipelines {
			name := *pipeline.Name

			// Check if pipeline name matches the pattern
			if pattern == "" || strings.Contains(name, pattern) {
				wg.Add(1)
				go getPipelineTags(client, cfg.Region, accountID, name, &wg, resultChan)
			}
		}
		wg.Wait()
		close(resultChan)

		for result := range resultChan {
			if len(tags) != 0 && tagsMatch(result.Tags, tags) {
				pipelines[result.PipelineName] = result
			} else if len(tags) == 0 {
				pipelines[result.PipelineName] = result
			}
		}
	}
	return pipelines, nil
}

// getTotalNumberOfPipelines is a function that returns the total number of AWS CodePipeLine
func getTotalNumberOfPipelines(cfg aws.Config) (int, error) {
	client := codepipeline.NewFromConfig(cfg)

	params := &codepipeline.ListPipelinesInput{}
	resp, err := client.ListPipelines(context.Background(), params)
	return len(resp.Pipelines), err
}

// getPipelineTags is a function that returns the tags of a AWS CodePipeLine TODO Rename this as it returns more than just tags
func getPipelineTags(client *codepipeline.Client, region, accountID, pipelineName string, wg *sync.WaitGroup, resultChan chan<- Pipeline) {
	defer wg.Done()
	result := Pipeline{
		PipelineName: pipelineName,
		Tags:         map[string]string{},
		Data:         nil,
	}

	var err error

	params := &codepipeline.ListPipelineExecutionsInput{
		PipelineName: aws.String(pipelineName),
		MaxResults:   aws.Int32(1),
	}

	// AWS doesn't provide any easy way to filter CodePipelines by tags
	// so we have to get all the pipelines and then filter them locally
	// AWS also seems to be pretty aggressive with their rate limiting
	// This is a bit of a hack but it works
	// Always filter by CodePipeline name first
	// Then filter by tags

	exponentialBackoff := 1
	for i := 0; i < 10; i++ {
		result.ExecData, err = client.ListPipelineExecutions(context.Background(), params)
		if err != nil {
			log.Debug().Str("model", "aws").Str("func", "getPipelineTags").Msgf("list executions query error (%v/10): %v", i+1, err)
			time.Sleep(time.Duration(exponentialBackoff) * time.Second)
			exponentialBackoff += exponentialBackoff
		} else {
			break
		}
	}
	if len(result.ExecData.PipelineExecutionSummaries) > 0 {
		result.LastExecutionID = *result.ExecData.PipelineExecutionSummaries[0].PipelineExecutionId
		result.LastExecutionStatus = string(result.ExecData.PipelineExecutionSummaries[0].Status)
	} else {
		result.LastExecutionID = ""
		result.LastExecutionStatus = "Unknown"
	}

	if len(result.ExecData.PipelineExecutionSummaries) == 0 {
		log.Debug().Str("model", "aws").Str("func", "getPipelineTags").Msgf("no executions found for %v", pipelineName)
	}

	pipelineArn := getArnPrefix(region, accountID, "codepipeline") + pipelineName

	var tagsResp *codepipeline.ListTagsForResourceOutput
	exponentialBackoff = 1
	for i := 0; i < 10; i++ {
		tagsResp, err = client.ListTagsForResource(context.Background(), &codepipeline.ListTagsForResourceInput{
			ResourceArn: &pipelineArn,
		})
		if err != nil {
			log.Debug().Str("model", "aws").Str("func", "getPipelineTags").Msgf("list tags query error (%v/10): %v", i+1, err)
			time.Sleep(time.Duration(exponentialBackoff) * time.Second)
			exponentialBackoff += exponentialBackoff
		} else {
			break
		}
	}

	tags := map[string]string{}
	for _, tag := range tagsResp.Tags {
		tags[*tag.Key] = *tag.Value
	}

	result.Tags = tags
	resultChan <- result
}

func tagsMatch(actualTags, expectedTags map[string]string) bool {
	for key, value := range expectedTags {
		if actualValue, found := actualTags[key]; !found || actualValue != value {
			return false
		}
	}
	return true
}

// GetPipelineInfo is a function that returns the information of a AWS CodePipeLine
func GetPipelineInfo(cfg aws.Config, pipelineName string) (*codepipeline.GetPipelineOutput, error) {
	client := codepipeline.NewFromConfig(cfg)

	params := &codepipeline.GetPipelineInput{
		Name: aws.String(pipelineName),
	}
	resp, err := client.GetPipeline(context.Background(), params)
	return resp, err
}

// GetPipelineState is a function that returns the state of a AWS CodePipeLine
func GetPipelineState(cfg aws.Config, pipelineName string) (*codepipeline.GetPipelineStateOutput, error) {
	client := codepipeline.NewFromConfig(cfg)

	params := &codepipeline.GetPipelineStateInput{
		Name: aws.String(pipelineName),
	}
	resp, err := client.GetPipelineState(context.Background(), params)
	return resp, err
}

// RetryPipelineStage is a function that retries a stage of a AWS CodePipeLine
func RetryPipelineStage(cfg aws.Config, pipelineExecutionID, pipelineName, stageName string) error {
	client := codepipeline.NewFromConfig(cfg)
	params := &codepipeline.RetryStageExecutionInput{
		PipelineExecutionId: aws.String(pipelineExecutionID),
		PipelineName:        aws.String(pipelineName),
		StageName:           aws.String(stageName),
		// Enforce Retry mode as "failed" actions for now
		RetryMode: types.StageRetryModeFailedActions,
	}
	_, err := client.RetryStageExecution(context.Background(), params)
	return err
}

// StartPipelineExecution is a function that starts a AWS CodePipeLine
func StartPipelineExecution(cfg aws.Config, pipelineName string) error {
	client := codepipeline.NewFromConfig(cfg)
	params := &codepipeline.StartPipelineExecutionInput{
		Name: aws.String(pipelineName),
	}
	_, err := client.StartPipelineExecution(context.Background(), params)
	return err
}

// DisablePipelineStageTransition is a function that disables a stage transition of a AWS CodePipeLine
func DisablePipelineStageTransition(cfg aws.Config, pipelineName, stageName, reason string) error {
	client := codepipeline.NewFromConfig(cfg)
	params := &codepipeline.DisableStageTransitionInput{
		PipelineName:   aws.String(pipelineName),
		StageName:      aws.String(stageName),
		Reason:         aws.String(reason),
		TransitionType: types.StageTransitionTypeInbound,
	}
	_, err := client.DisableStageTransition(context.Background(), params)
	return err
}

// EnablePipelinStageTransition is a function that enables a stage transition of a AWS CodePipeLine
func EnablePipelinStageTransition(cfg aws.Config, pipelineName, stageName string) error {
	client := codepipeline.NewFromConfig(cfg)
	params := &codepipeline.EnableStageTransitionInput{
		PipelineName:   aws.String(pipelineName),
		StageName:      aws.String(stageName),
		TransitionType: types.StageTransitionTypeInbound,
	}
	_, err := client.EnableStageTransition(context.Background(), params)
	return err
}
