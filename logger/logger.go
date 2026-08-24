package logger

import (
	"io"
	"time"

	"github.com/rs/zerolog"
)

func New(writer io.Writer) zerolog.Logger {
	return zerolog.New(writer).
		With().
		Timestamp().
		Logger()
}

func Configure() {
	zerolog.TimeFieldFormat = time.RFC3339Nano
}
