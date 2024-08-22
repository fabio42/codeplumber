package cmd

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func listProfiles() {
	profiles := getProfiles()
	if len(profiles) == 0 { // no profiles
		log.Error().Msg("No profiles found in the config file")
	} else {
		for _, profile := range profiles {
			fmt.Println(profile)
		}
	}
}

func getProfiles() []string {
	return k.MapKeys("profiles")
}

var profilesCmd = &cobra.Command{
	Use:   "profiles",
	Short: "List available profiles",
	Run: func(_ *cobra.Command, _ []string) {
		listProfiles()
	},
}

func init() {
	rootCmd.AddCommand(profilesCmd)
}
