package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func NewProduction() zerolog.Logger {
	return zerolog.New(os.Stdout).
		Level(zerolog.InfoLevel).
		With().
		Timestamp().
		Logger()
}
