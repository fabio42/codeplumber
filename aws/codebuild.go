package awsqueries

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
)

// CodebuildData holds the details of a codebuild
type CodebuildData struct {
	Name    string
	Project *codebuild.BatchGetProjectsOutput
	Builds  *codebuild.BatchGetBuildsOutput
}

// GetCodeBuildData returns a list of codebuild details
func GetCodeBuildData(cfg aws.Config, accountID, name, buildID string) (CodebuildData, error) {
	var cb CodebuildData
	var err error
	cb.Name = name
	cb.Builds, err = GetCodeBuildBuilds(cfg, accountID, buildID)
	if err != nil {
		return cb, err
	}
	cb.Project, err = GetCodeBuildProjects(cfg, accountID, name)
	if err != nil {
		return cb, err
	}
	if len(cb.Builds.Builds) != 1 || len(cb.Project.Projects) != 1 {
		return cb, fmt.Errorf("multiple codebuild returned for %v", buildID)
	}
	return cb, nil
}

// GetCodeBuildBuilds returns a list of buils details
func GetCodeBuildBuilds(cfg aws.Config, accountID, buildID string) (*codebuild.BatchGetBuildsOutput, error) {
	var builds *codebuild.BatchGetBuildsOutput
	var err error

	client := codebuild.NewFromConfig(cfg)

	arnPrefix := getArnPrefix(cfg.Region, accountID, "codebuildBuild")
	arn := fmt.Sprintf("%s%s", arnPrefix, buildID)
	input := &codebuild.BatchGetBuildsInput{
		Ids: []string{arn},
	}

	builds, err = client.BatchGetBuilds(context.TODO(), input)
	return builds, err
}

// GetCodeBuildProjects returns a list of projects
func GetCodeBuildProjects(cfg aws.Config, accountID, buildID string) (*codebuild.BatchGetProjectsOutput, error) {
	client := codebuild.NewFromConfig(cfg)

	arnPrefix := getArnPrefix(cfg.Region, accountID, "codebuildProject")
	arn := fmt.Sprintf("%s%s", arnPrefix, buildID)

	input := &codebuild.BatchGetProjectsInput{
		Names: []string{arn},
	}

	projects, err := client.BatchGetProjects(context.Background(), input)
	return projects, err
}
