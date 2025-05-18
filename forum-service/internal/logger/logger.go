package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

func init() {
	zerolog.TimeFieldFormat = time.RFC3339
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log = zerolog.New(output).With().Timestamp().Caller().Logger()
}

// GetLogger returns the global logger instance
func GetLogger() *zerolog.Logger {
	return &log
}

// Info logs an info message
func Info(msg string, fields ...map[string]interface{}) {
	event := log.Info()
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}
	event.Msg(msg)
}

// Error logs an error message
func Error(err error, msg string, fields ...map[string]interface{}) {
	event := log.Error().Err(err)
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}
	event.Msg(msg)
}

// Debug logs a debug message
func Debug(msg string, fields ...map[string]interface{}) {
	event := log.Debug()
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}
	event.Msg(msg)
}

// Fatal logs a fatal message and exits
func Fatal(err error, msg string, fields ...map[string]interface{}) {
	event := log.Fatal().Err(err)
	if len(fields) > 0 {
		for key, value := range fields[0] {
			event = event.Interface(key, value)
		}
	}
	event.Msg(msg)
}
