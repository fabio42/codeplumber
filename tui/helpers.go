package tui

import (
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

func percent(number, percentage, max int) int {
	if max > 0 {
		res := (percentage * number) / 100
		if res > max {
			return max
		}
		return res
	}
	return (percentage * number) / 100
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// PrintTime return a string with the time in the format YYYY-MM-DD HH:MM:SS
func PrintTime(t *time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func getTermSize() (width, height int) {
	if term.IsTerminal(0) {
		log.Debug().Str("model", "tui").Str("func", "getTermSize").Msg("terminal detected")
	} else {
		log.Debug().Str("model", "tui").Str("func", "getTermSize").Msg("not in a terminal")
	}
	width, height, err := term.GetSize(0)
	if err != nil {
		log.Debug().Str("model", "tui").Str("func", "getTermSize").Msgf("error getting terminal size: %v", err)
		return
	}
	log.Debug().Str("model", "tui").Str("func", "getTermSize").Msgf("width: %v, height: %v", width, height)
	return width, height
}

func supportedView(s string) bool {
	for _, v := range supportedViews {
		if v == s {
			return true
		}
	}
	return false
}
