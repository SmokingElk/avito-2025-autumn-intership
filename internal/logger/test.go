package logger

import (
	"io"

	"github.com/rs/zerolog"
)

func NewTest() zerolog.Logger {
	return zerolog.New(io.Discard).Level(zerolog.Disabled)
}
