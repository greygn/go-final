package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func GetLogger() zerolog.Logger {
	return zerolog.New(os.Stdout).With().Timestamp().Logger()
}
