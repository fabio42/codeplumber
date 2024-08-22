package cmd

import (
	"fmt"
	"maps"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"github.com/rs/zerolog/log"
)

var k = koanf.New(".")

// NEW
func loadProfile(profile string) error {

	if k.Exists("profiles." + profile) {
		rootFlags.awsProfile = k.String("profiles." + profile + ".aws.profile")
		rootFlags.awsRegion = k.String("profiles." + profile + ".aws.region")

		if rootFlags.tagsFilter == nil {
			rootFlags.tagsFilter = make(map[string]string)
		}
		maps.Copy(rootFlags.tagsFilter, k.StringMap("profiles."+profile+".filters.tags"))
		rootFlags.nameFilter = k.String("profiles." + profile + ".filters.name")
	} else {
		return fmt.Errorf("Profile %s does not exist in config file", profile)
	}
	return nil
}

func loadConfigFile() error {
	path, err := expandPath(rootFlags.configFile)
	if err != nil {
		return err
	}

	log.Debug().Str("model", "cmd").Str("func", "loadConfigFile").Msgf("Loading config file %s", path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Warn().Msgf("Config file %s does not exist", path)
		return nil // no configuration file present
	}

	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return err
	}

	// If debug mode is enabled on the  command line, don't override it with the config file
	if !rootFlags.debug {
		rootFlags.debug = k.Bool("debug")
		setLogger(rootFlags.logLevel)
	}

	return nil
}

// ExpandPath takes a directory path as input, resolves any environment variables, and returns the full path.
// It replace ${VAR} and $VAR with the value of the environment variable
func expandPath(path string) (string, error) {
	var result string
	var err error
	result = path
	for {
		// Look for environment variable patterns
		start := strings.IndexAny(result, "$")
		if start == -1 {
			break
		}

		var end int
		// If the pattern is ${VAR}
		if result[start] == '$' && (start+1 < len(result) && result[start+1] == '{') {
			end = strings.IndexByte(result[start+2:], '}')
			if end == -1 {
				return "", fmt.Errorf("unmatched '{' in path: %s", result)
			}
			end += start + 2
			varName := result[start+2 : end]
			fmt.Println("VarName:", varName)
			varValue, found := os.LookupEnv(varName)
			if !found {
				return "", fmt.Errorf("environment variable %s not found", varName)
			}
			result = result[:start] + varValue + result[end+1:]
		} else if result[start] == '$' {
			end = strings.IndexAny(result[start+1:], "/\\:")
			if end == -1 {
				end = len(result)
			} else {
				end += start + 1
			}
			varName := result[start+1 : end]
			varValue, found := os.LookupEnv(varName)
			if !found {
				return "", fmt.Errorf("environment variable %s not found", varName)
			}
			result = result[:start] + varValue + result[end:]
		}
	}

	return result, err
}
