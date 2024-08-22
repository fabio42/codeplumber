package cmd

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	logFile = "./codeplumber.log"
)

// LevelWriter interface
type LevelWriter struct {
	io.Writer
	Level zerolog.Level
}

// WriteLevel defines the log level
func (lw *LevelWriter) WriteLevel(l zerolog.Level, p []byte) (n int, err error) {
	if l >= lw.Level {
		return lw.Write(p)
	}
	return len(p), nil
}

// setLogger sets the logger
// It provide a way to set multiple logs outputs depending of the log level
func setLogger(level string) error {
	var logWriter *os.File
	var debug bool
	var err error
	var logLevel zerolog.Level

	switch level {
	case "debug":
		logLevel = zerolog.DebugLevel
		debug = true
	case "fatal":
		logLevel = zerolog.FatalLevel
	default:
		logLevel = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(logLevel)
	if debug {
		logWriter, err = os.OpenFile(
			logFile,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0664,
		)
		if err != nil {
			panic(err)
		}
	}

	fileWriter := zerolog.New(zerolog.ConsoleWriter{
		Out:          logWriter,
		NoColor:      true,
		PartsExclude: []string{"time", "level"},
	})

	consoleWriter := zerolog.NewConsoleWriter(
		func(w *zerolog.ConsoleWriter) {
			w.Out = os.Stderr
			w.PartsExclude = []string{"time"}
		},
	)

	// We limit log level to info on the console
	// Debug messages will be found on the log file only if debug is enabled
	consoleWriterLeveled := &LevelWriter{Writer: consoleWriter, Level: zerolog.InfoLevel}
	if debug {
		log.Logger = zerolog.New(zerolog.MultiLevelWriter(fileWriter, consoleWriterLeveled)).With().Timestamp().Logger()
	} else {
		log.Logger = zerolog.New(consoleWriterLeveled).With().Timestamp().Logger()
	}
	return nil
}
