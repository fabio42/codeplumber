package tui

import (
	"fmt"
	"slices"
	"time"

	awsqueries "github.com/fabio42/codeplumber/aws"
	"github.com/fabio42/codeplumber/models/table"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline"
	"github.com/aws/aws-sdk-go-v2/service/codepipeline/types"
	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
)

const (
	separatorTransition = "--"
	separatorStage      = "└"
)

// PipelineResource describe a AWS CodePipeline Resource
type PipelineResource struct {
	PipelineName        string
	StageName           string
	ActionName          string
	ExternalExecutionID string
	Status              string
}

func (m *PipelineTable) refresh() {
	go refreshPipelineOps(m.name, m.ui)
}

func refreshPipelineOps(name string, c *uiData) {
	var rows []table.Row
	var err error
	var stateData *codepipeline.GetPipelineStateOutput
	var infoData *codepipeline.GetPipelineOutput

	c.startSpinner()

	if config.Mode.Replay {
		stateData = c.dataCache.pipelines[name].StateData
		infoData = c.dataCache.pipelines[name].Data
	} else if pipeline, ok := c.dataCache.pipelines[name]; ok {
		stateData, err = awsqueries.GetPipelineState(config.AwsConfig, name)
		if err != nil {
			log.Fatal().Msgf("error: %v", err)
		}

		infoData, err = awsqueries.GetPipelineInfo(config.AwsConfig, name)
		if err != nil {
			log.Fatal().Msgf("error: %v", err)
		}

		pipeline.StateData = stateData
		pipeline.Data = infoData

		c.dataCache.pipelines[name] = pipeline
		config.Recorder.Record("DataCachePipelines.json", c.dataCache.pipelines)
	}

	for stageKey, stage := range c.dataCache.pipelines[name].Data.Pipeline.Stages {
		stageState := stateData.StageStates[stageKey]
		if len(rows) > 0 {
			var transitionState string

			if stageState.InboundTransitionState.Enabled {
				transitionState = "Enabled"
			} else {
				transitionState = "Disabled"
			}

			rows = append(rows, table.Row{"", "", "", "", ""})
			rows = append(rows, table.Row{
				separatorTransition + " Transition",
				"Transition",
				*stage.Name,
				transitionState,
				"",
			})
			rows = append(rows, table.Row{"", "", "", "", ""})
		}

		var execStatus string
		if stageState.LatestExecution == nil {
			execStatus = "N/A"
		} else {
			execStatus = string(stageState.LatestExecution.Status)
		}
		rows = append(rows, table.Row{
			*stage.Name,
			"",
			"",
			execStatus,
			"",
		})

		for actionKey, action := range stage.Actions {
			var actionStatus string
			var actionLastUpdateTime string

			actions := stageState.ActionStates[actionKey]

			switch action.ActionTypeId.Category {
			case "Approval":
				actionStatus = string(actions.LatestExecution.Status)
				if actionStatus == "InProgress" && actions.LatestExecution.LastStatusChange == nil {
					actionStatus = "Pending"
					actionLastUpdateTime = "N/A"
				} else {
					actionLastUpdateTime = PrintTime(actions.LatestExecution.LastStatusChange)
				}

			default:
				if actions.LatestExecution == nil || actions.LatestExecution.LastStatusChange == nil {
					actionStatus = "Waiting"
					actionLastUpdateTime = "..."
				} else {
					log.Debug().Str("model", "tui").Str("func", "refreshPipelineOps").Msgf("action: %v", actions)
					actionStatus = string(actions.LatestExecution.Status)
					actionLastUpdateTime = PrintTime(actions.LatestExecution.LastStatusChange)
				}
			}

			actionName := *action.Name
			actionType := aws.ToString(action.ActionTypeId.Provider)
			actionCategory := string(action.ActionTypeId.Category)

			rows = append(rows, table.Row{
				fmt.Sprintf("%s %v", separatorStage, actionName),
				fmt.Sprintf("%v/%v", actionType, actionCategory),
				*stage.Name,
				actionStatus,
				actionLastUpdateTime,
			})
		}
	}
	c.stopSpinner()
	c.updateView(pipelineView, rows)
}

// Return PipelineData for the given pipeline name
func (m *PipelineTable) getCodebuildResource(pipelineName, stageName, stageSection, actionName string) (PipelineResource, error) {
	var pipelineData PipelineResource
	var stateData *codepipeline.GetPipelineStateOutput
	var err error

	pipeline := m.ui.dataCache.pipelines[pipelineName]
	stateData = pipeline.StateData
	stageIdx := findStageByName(stateData.StageStates, stageName)
	actionIdx := findActionByName(stateData.StageStates[stageIdx].ActionStates, actionName)

	pipelineData.PipelineName = pipelineName
	pipelineData.StageName = stageSection
	pipelineData.ActionName = actionName
	pipelineData.Status = string(stateData.StageStates[stageIdx].ActionStates[actionIdx].LatestExecution.Status)

	for stateData.StageStates[stageIdx].ActionStates[actionIdx].LatestExecution.ExternalExecutionId == nil {
		return pipelineData, fmt.Errorf("pipelineDataNotReady")
	}
	pipelineData.ExternalExecutionID = *stateData.StageStates[stageIdx].ActionStates[actionIdx].LatestExecution.ExternalExecutionId

	return pipelineData, err
}

func findStageByName(stages []types.StageState, name string) int {
	for idx, stage := range stages {
		if *stage.StageName == name {
			return idx
		}
	}
	return -1
}

func findActionByName(actions []types.ActionState, name string) int {
	for idx, action := range actions {
		if *action.ActionName == name {
			return idx
		}
	}
	return -1
}

func (m *PipelineTable) start() {
	awsqueries.StartPipelineExecution(config.AwsConfig, m.name)
}

func (m *PipelineTable) restartStage(pipeline PipelineResource) {
	pipelineExcutionID := m.ui.dataCache.pipelines[pipeline.PipelineName].LastExecutionID
	err := awsqueries.RetryPipelineStage(config.AwsConfig, pipelineExcutionID, pipeline.PipelineName, pipeline.StageName)
	if err != nil {
		log.Debug().Str("model", "tui").Str("func", "PipelineTable.restartStage").Msgf("Error starting pipeline: %v", err)
	}
}

func (m *PipelineTable) toggleTransition(action string, reason string) error {
	m.ui.startSpinner()
	row := m.SelectedRow()
	var err error
	if action == "disable" {
		err = awsqueries.DisablePipelineStageTransition(config.AwsConfig, m.name, row[2], reason)
	} else {
		go m.start()
		for true {
			inProgress := false
			time.Sleep(2 * time.Second)
			go refreshPipelineOps(m.name, m.ui)
			for _, r := range m.Rows() {
				if slices.Contains(r, "InProgress") {
					inProgress = true
					break
				}
			}
			if !inProgress {
				break
			}
		}

		err = awsqueries.EnablePipelinStageTransition(config.AwsConfig, m.name, row[2])
	}
	m.ui.stopSpinner()
	return err
}

func (m *PipelineTable) browse() {
	browser.OpenURL(fmt.Sprintf("https://%s.console.aws.amazon.com/codesuite/codepipeline/pipelines/%v/view", config.AwsConfig.Region, m.name))
}
