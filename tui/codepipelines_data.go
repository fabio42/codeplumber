package tui

import (
	awsqueries "codeplumber/aws"
	"codeplumber/models/table"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
)

func (p *PipelinesTable) refresh() {
	go pipelinesTableRefresh(p.ui)
}

func pipelinesTableRefresh(c *uiData) {
	c.startSpinner()
	var err error

	if config.Mode.Replay {
		c.dataCache.pipelines, err = config.Recorder.ListCodePipelines("DataCachePipelines.json")
	} else {
		c.dataCache.pipelines, err = awsqueries.CodePipelinesListFiltered(config.AwsConfig, config.awsAccountID, config.NameFilter, config.TagFilter)
		config.Recorder.Record("DataCachePipelines.json", c.dataCache.pipelines)
	}
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to list AWS CodePipeline: %v", err)
	}

	pipelines := c.dataCache.pipelines
	rows := make([]table.Row, len(pipelines))

	var userID, status, lastExecTime string
	idx := 0
	for _, pipeline := range pipelines {
		log.Debug().Str("model", "tui").Str("func", "pipelinesTableRefresh").Msgf("Pipeline: %v", pipeline.LastExecutionID)
		if pipeline.LastExecutionID == "" {
			userID = ""
			status = "Unknown"
			lastExecTime = ""
		} else {
			user := strings.Split(string(*pipeline.ExecData.PipelineExecutionSummaries[0].Trigger.TriggerDetail), "/")
			userID = user[len(user)-1]
			if strings.HasPrefix(userID, "AWSCodeBuild") {
				userID = "CodeBuild"
			}
			status = string(pipeline.ExecData.PipelineExecutionSummaries[0].Status)
			lastExecTime = PrintTime(pipeline.ExecData.PipelineExecutionSummaries[0].LastUpdateTime)
		}
		rows[idx] = table.Row{
			pipeline.PipelineName,
			userID,
			status,
			lastExecTime,
		}
		idx++
	}

	if config.NameFilterExtra != "" {
		rows = extraNameFilter(rows, config.NameFilterExtra)
	}

	// Sort CodePipelines by name
	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	c.initialized = true
	c.stopSpinner()
	c.updateView(pipelinesView, rows)
}

func (p *PipelinesTable) filterOperations(f string) {
	rows := extraNameFilter(p.allRows, f)
	p.ui.updateView(pipelinesView, slices.Clip(rows))
}

func (p *PipelinesTable) browse() {
	browser.OpenURL(fmt.Sprintf("https://%s.console.aws.amazon.com/codesuite/codepipeline/pipelines", config.AwsConfig.Region))
}

func (p *PipelinesTable) start(pipelineName string) {
	awsqueries.StartPipelineExecution(config.AwsConfig, pipelineName)
}

func sliceContainString(xs []string, s string) bool {
	for _, v := range xs {
		if strings.Contains(v, s) {
			return true
		}
	}
	return false
}

func extraNameFilter(rowsSrc []table.Row, filter string) []table.Row {
	rows := make([]table.Row, len(rowsSrc))
	idx := 0
	for _, row := range rowsSrc {
		if sliceContainString(row, filter) {
			rows[idx] = row
			idx++
		}
	}
	return slices.Clip(rows[:idx])
}
