package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var rootFlags struct {
	awsProfile      string
	awsRegion       string
	configFile      string
	debug           bool
	logLevel        string
	listProfiles    bool
	nameFilter      string
	nameFilterExtra string
	noConfig        bool
	profile         string
	record, replay  bool
	recordDir       string
	tagsFilter      map[string]string
}

var rootCmd = &cobra.Command{
	Use:   "codeplumber [flags] command",
	Short: "codeplumber is the missing tool to manage your AWS CodePipeline resources.",

	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		if rootFlags.debug {
			rootFlags.logLevel = "debug"
		}

		if err := setLogger(rootFlags.logLevel); err != nil {
			log.Fatal().Msgf("Error failed to configure logger: %v", err)
		}

		if err := loadConfigFile(); err != nil {
			log.Fatal().Msgf("Error failed to load config file: %v", err)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootFlags.configFile, "config", "c", "$HOME/.config/codeplumber/config.yaml", "CodePlumber Configuration file location.")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.awsProfile, "aws-profile", "p", "", "Use a specific AWS profile from your AWS credential file, overwrite ENV variable AWS_PROFILE.")
	rootCmd.PersistentFlags().StringVarP(&rootFlags.awsRegion, "aws-region", "r", "us-east-1", "The AWS region to use, overwrite ENV variable AWS_REGION.")
	rootCmd.PersistentFlags().BoolVarP(&rootFlags.debug, "debug", "d", false, "Enable debug log, out will be saved in "+logFile)

	// TODO: Consider making this attribute mandatory when using record/replay options
	rootCmd.PersistentFlags().StringVar(&rootFlags.recordDir, "record-dir", "$HOME/.local/state/codeplumber", "Directory to store recordings")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.record, "record", false, "Record the output of the command to a directory (default:"+rootFlags.recordDir+")")
	rootCmd.PersistentFlags().BoolVar(&rootFlags.replay, "replay", false, "Replay the output of the command from a recording")

	// These flags are for development/debug purpose only, they are hidden from the user help
	rootCmd.MarkFlagsMutuallyExclusive("record", "replay")
	rootCmd.PersistentFlags().MarkHidden("record-dir")
	rootCmd.PersistentFlags().MarkHidden("record")
	rootCmd.PersistentFlags().MarkHidden("replay")

	err := setLogger(rootFlags.logLevel)
	if err != nil {
		fmt.Println("Error failed to configure logger:", err)
		os.Exit(1)
	}
}

// Execute is the main entry point for the CLI
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Msgf("Whoops. There was an error while executing your CLI '%s'", err)
	}
}
