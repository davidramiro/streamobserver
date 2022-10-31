package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Log is the global logger for the application.
var Log zerolog.Logger

// InitLog initializes the global logger, setting the log level according to the debug parameter.
func InitLog(jsonLogging bool) {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}

	multi := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)

	if jsonLogging {
		Log = zerolog.New(multi).With().Timestamp().Logger()
	} else {
		Log = zerolog.New(consoleWriter).With().Timestamp().Logger()
	}

}
