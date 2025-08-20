// Package utils internal/utils/logger.go
package utils

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func InitLogger(environment string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if environment == "development" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	Logger = log.With().
		Str("service", "messenger-backend").
		Logger()
}

func LogError(err error, msg string, fields ...map[string]interface{}) {
	event := Logger.Error().Err(err).Str("message", msg)
	for _, f := range fields {
		for k, v := range f {
			event = event.Interface(k, v)
		}
	}
	event.Send()
}

func LogInfo(msg string, fields ...map[string]interface{}) {
	event := Logger.Info().Str("message", msg)
	for _, f := range fields {
		for k, v := range f {
			event = event.Interface(k, v)
		}
	}
	event.Send()
}

func LogDebug(msg string, fields ...map[string]interface{}) {
	event := Logger.Debug().Str("message", msg)
	for _, f := range fields {
		for k, v := range f {
			event = event.Interface(k, v)
		}
	}
	event.Send()
}
