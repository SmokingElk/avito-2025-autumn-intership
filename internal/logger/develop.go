package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func NewDevelop() zerolog.Logger {
	consoleWriter := zerolog.ConsoleWriter{
		Out: os.Stdout,
	}

	return zerolog.New(consoleWriter).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Logger()
}
