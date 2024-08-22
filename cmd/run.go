package cmd

import (
	"context"
	"fmt"

	"github.com/fabio42/codeplumber/tui"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [--tags key1=value1,key2=value2] [NAME_FILTER]",
	Short: "Run codeplumber with optional tags and name filter",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {

		if len(args) > 0 {
			rootFlags.nameFilter = args[0]
		}

		log.Debug().Str("model", "cmd").Str("func", "runCmdRun").Msgf("Name filter: %s", rootFlags.nameFilter)
		log.Debug().Str("model", "cmd").Str("func", "runCmdRun").Msgf("Extra name filter: %s", rootFlags.nameFilterExtra)
		log.Debug().Str("model", "cmd").Str("func", "runCmdRun").Msgf("AWS profile: %s", rootFlags.awsProfile)
		log.Debug().Str("model", "cmd").Str("func", "runCmdRun").Msgf("AWS region: %s", rootFlags.awsRegion)
		log.Debug().Str("model", "cmd").Str("func", "runCmdRun").Msgf("Tags filter: %v", rootFlags.tagsFilter)

		// Filtering only on tags is not supported natively by AWS APIs and quite ineffective as we need to describe
		// all resources one by one to extract the tags. For this reason filtering only tags is not allowed.
		if len(rootFlags.tagsFilter) > 0 && rootFlags.nameFilter == "" {
			log.Fatal().Msg("Filtering only on tags is not supported at the moment, please provide a name filter as well.")
		}

		run()
	},
}

func init() {
	runCmd.Flags().StringToStringVarP(&rootFlags.tagsFilter, "tags", "t", nil, "Filter resources by tags, e.g. --tags key1=value1,key2=value2")
	rootCmd.AddCommand(runCmd)
}

func run() error {
	log.Info().Msg("Starting codeplumber")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(rootFlags.awsRegion),
		config.WithSharedConfigProfile(rootFlags.awsProfile),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), 5)
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS configuration: %w", err)
	}

	var tuicfg tui.Config
	tuicfg.AwsConfig = cfg
	if rootFlags.replay && rootFlags.record {
		// TDOD this should be managed by cobra
		return fmt.Errorf("--record and --replay are mutually exclusive")
	}

	tuicfg.Mode.Record = rootFlags.record
	tuicfg.Mode.Replay = rootFlags.replay
	recordDir, err := expandPath(rootFlags.recordDir)
	if err != nil {
		return err
	}
	tuicfg.RecordDir = recordDir
	tuicfg.NameFilter = rootFlags.nameFilter
	tuicfg.NameFilterExtra = rootFlags.nameFilterExtra
	tuicfg.TagFilter = rootFlags.tagsFilter

	m := tui.NewModel(tuicfg)
	_, err = tea.NewProgram(m).Run()
	return err
}
