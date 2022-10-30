package logger

import (
	"os"

	"github.com/rs/zerolog"
)

// Log is the global logger for the application.
var Log zerolog.Logger

// InitLog initializes the global logger, setting the log level according to the debug parameter.
func InitLog() {
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	multi := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)

	Log = zerolog.New(multi).With().Timestamp().Logger()
}
