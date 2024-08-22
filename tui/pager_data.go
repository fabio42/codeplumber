package tui

import (
	awsqueries "codeplumber/aws"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/rs/zerolog/log"
)

type cwlData struct {
	content string
	token   *string
}

func (m *Pager) refreshLog() {
	go refreshLogOps(m)
}

func refreshLogOps(p *Pager) {
	var logStream strings.Builder
	var events *cloudwatchlogs.GetLogEventsOutput
	var err error

	p.ui.startSpinner()

	groupName := p.ui.dataCache.codebuilds[p.name].Builds.Builds[0].Logs.GroupName
	streamName := p.ui.dataCache.codebuilds[p.name].Builds.Builds[0].Logs.StreamName

	if config.Mode.Replay {
		events, err = config.Recorder.GetCloudWatchLogs(fmt.Sprintf("CloudWatchLogs-%s.txt", filepath.Base(*groupName)))
		p.lastLogToken = events.NextForwardToken
	} else {
		events, p.lastLogToken, err = awsqueries.GetCloudWatchLogs(config.AwsConfig, *groupName, *streamName, p.lastLogToken)
		config.Recorder.Record(fmt.Sprintf("CloudWatchLogs-%s.txt", filepath.Base(*groupName)), events)
	}
	if err != nil {
		log.Fatal().Msgf("error: %v", err)
	}

	for _, event := range events.Events {
		logStream.WriteString(*event.Message)
	}

	p.ui.stopSpinner()
	p.ui.updateView(logView, PagerSelector{
		name:    p.name,
		token:   p.lastLogToken,
		content: logStream.String(),
	})
}
