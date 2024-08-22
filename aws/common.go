package awsqueries

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/rs/zerolog/log"
)

// GetAwsAccountID returns the AWS Account ID
func GetAwsAccountID(cfg aws.Config) string {
	client := sts.NewFromConfig(cfg)
	accountID, err := client.GetCallerIdentity(context.Background(), &sts.GetCallerIdentityInput{})
	if err != nil {
		log.Fatal().Msgf("%v", err)
	}
	return *accountID.Account
}

// Little helper to get the ARN prefix for a given endpoint
// This save a lots of API calls to AWS, but this won't support special cases like gov-cloud at the moment
func getArnPrefix(region, accountID, endpoint string) string {
	switch endpoint {
	case "codebuildBuild":
		return fmt.Sprintf("arn:aws:codebuild:%v:%v:build/", region, accountID)
	case "codebuildProject":
		return fmt.Sprintf("arn:aws:codebuild:%v:%v:project/", region, accountID)
	case "codepipeline":
		return fmt.Sprintf("arn:aws:codepipeline:%v:%v:", region, accountID)
	default:
		return ""
	}
}
