package tui

import (
	"fmt"
	"strconv"
	"strings"

	awsqueries "github.com/fabio42/codeplumber/aws"
	"github.com/fabio42/codeplumber/models/table"

	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
)

func (m *CodeBuildTable) refresh(buildID string) {
	m.name = strings.Split(buildID, ":")[0]
	go refreshCodebuildOps(m.ui, m.name, buildID)
}

func refreshCodebuildOps(c *uiData, name, buildID string) {
	var cb awsqueries.CodebuildData
	var err error
	rows := make([]table.Row, 19)

	c.startSpinner()

	if config.Mode.Replay {
		cbCache, err := config.Recorder.GetCodeBuilds("DataCacheCodebuilds.json")
		if err != nil {
			log.Fatal().Msgf("error: %v", err)
		}
		cb = cbCache[name]
	} else {
		cb, err = awsqueries.GetCodeBuildData(config.AwsConfig, config.awsAccountID, name, buildID)
		if err != nil {
			log.Fatal().Msgf("error: %v", err)
		}
	}
	c.dataCache.codebuilds[cb.Name] = cb
	config.Recorder.Record("DataCacheCodebuilds.json", c.dataCache.codebuilds)

	build := cb.Builds.Builds[0]
	project := cb.Project.Projects[0]

	logFriendlyURL := fmt.Sprintf(
		"https://%v.console.aws.amazon.com/codesuite/codebuild/%v/projects/%v/build/%v/?region=%v",
		config.AwsConfig.Region,
		config.awsAccountID,
		*project.Name,
		*build.Id,
		config.AwsConfig.Region)

	rows[0] = table.Row{"Project Name", *project.Name}
	rows[1] = table.Row{"Description", *project.Description}
	rows[2] = table.Row{"Build Status", string(build.BuildStatus)}
	rows[3] = table.Row{"Build ID", *build.Id}
	buildSpecData := ""
	if strings.HasPrefix(*project.Source.Buildspec, "version:") {
		buildSpecData = "Inline BuildSpec, press enter to see it"
	} else {
		buildSpecData = *project.Source.Buildspec
	}
	rows[4] = table.Row{"BuildSpec", buildSpecData}
	rows[5] = table.Row{"", ""}
	rows[6] = table.Row{"Source Type", string(build.Source.Type)}
	rows[7] = table.Row{"", ""}
	rows[8] = table.Row{"Environment:", ""}
	rows[9] = table.Row{"  Compute Type", string(build.Environment.ComputeType)}
	rows[10] = table.Row{"  Image", *build.Environment.Image}
	rows[11] = table.Row{"  Type", string(build.Environment.Type)}
	rows[12] = table.Row{"  Privileged Mode", strconv.FormatBool(*build.Environment.PrivilegedMode)}
	rows[13] = table.Row{"", ""}
	rows[14] = table.Row{"Logs:", ""}
	rows[15] = table.Row{"  URL", logFriendlyURL}
	rows[16] = table.Row{"  Group Name", *build.Logs.CloudWatchLogs.GroupName}
	var streamName string
	if build.Logs.StreamName == nil {
		streamName = "N/A"
	} else {
		streamName = *build.Logs.StreamName
	}
	rows[17] = table.Row{"  Stream Name", streamName}
	rows[18] = table.Row{"", ""}
	rows = append(rows, table.Row{"Environment Variables:", ""})
	for _, envVar := range build.Environment.EnvironmentVariables {
		rows = append(rows, table.Row{"  " + *envVar.Name, *envVar.Value})
	}
	rows = append(rows, table.Row{"", ""})

	if build.VpcConfig == nil {
		rows = append(rows, table.Row{"VPC Configuration:", "Not configured"})
	} else {
		rows = append(rows, table.Row{"VPC Configuration:", ""})
		rows = append(rows, table.Row{"  VPC ID", *build.VpcConfig.VpcId})
		// TODO: get Subnet Name and SG Name
		rows = append(rows, table.Row{"  Subnets IDs", strings.Join(build.VpcConfig.Subnets, ", ")})
		rows = append(rows, table.Row{"  Security Groups IDs", strings.Join(build.VpcConfig.SecurityGroupIds, ", ")})
	}

	rows = append(rows, table.Row{"", ""})
	rows = append(rows, table.Row{"Phases:", ""})
	for _, phase := range build.Phases {
		rows = append(rows, table.Row{"  " + string(phase.PhaseType), string(phase.PhaseStatus)})
	}
	rows = append(rows, table.Row{"", ""})
	rows = append(rows, table.Row{"Tags:", ""})
	for _, tag := range project.Tags {
		rows = append(rows, table.Row{"  " + *tag.Key, *tag.Value})
	}

	c.stopSpinner()
	c.updateView(codebuildView, rows)
}

func (m *CodeBuildTable) browse() {
	for _, r := range m.Rows() {
		if strings.Contains(r[0], "URL") {
			browser.OpenURL(r[1])
		}
	}
}

func (m *CodeBuildTable) statusFailed() bool {
	if m.ui.dataCache.codebuilds[m.name].Builds.Builds[0].BuildStatus == "FAILED" {
		return true
	}
	return false
}

func (m *CodeBuildTable) start() {
	m.ui.startSpinner()
	pipelineExcutionID := m.ui.dataCache.pipelines[m.pipelineName].LastExecutionID
	awsqueries.RetryPipelineStage(config.AwsConfig, pipelineExcutionID, m.pipelineName, m.stageName)
	m.ui.stopSpinner()
}
