package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// TODO: add completion function
// https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md

var loadCmd = &cobra.Command{
	Use:   "load PROFILE_NAME [EXTRA_NAME_FILTER]",
	Short: "Load a profile",
	Args:  cobra.RangeArgs(1, 2),

	ValidArgsFunction: func(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return getProfiles(), cobra.ShellCompDirectiveNoFileComp
	},

	Run: func(_ *cobra.Command, args []string) {
		if err := loadProfile(args[0]); err != nil {
			log.Fatal().Msgf("Error loading profile: %s", err)
		}
		if len(args) > 1 {
			rootFlags.nameFilterExtra = args[1]
		}

		log.Debug().Str("model", "cmd").Str("func", "loadCmdRun").Msgf("Loading profile %s", args[0])
		log.Debug().Str("model", "cmd").Str("func", "loadCmdRun").Msgf("Name filter: %s", rootFlags.nameFilter)
		log.Debug().Str("model", "cmd").Str("func", "loadCmdRun").Msgf("Extra name filter: %s", rootFlags.nameFilterExtra)
		log.Debug().Str("model", "cmd").Str("func", "loadCmdRun").Msgf("AWS profile: %s", rootFlags.awsProfile)
		log.Debug().Str("model", "cmd").Str("func", "loadCmdRun").Msgf("AWS region: %s", rootFlags.awsRegion)
		log.Debug().Str("model", "cmd").Str("func", "loadCmdRun").Msgf("Tags filter: %v", rootFlags.tagsFilter)

		run()
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}
