package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/stackloklabs/gollm/pkg/config"
)

var logger zerolog.Logger

func InitLogger() {
	cfg := config.InitializeViperConfig("config", "yaml", ".")
	// Set the time format for logs
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Retrieve log level from configuration (using viper as an example)
	log_level := cfg.Get("log_level")

	// Parse log level
	level, err := zerolog.ParseLevel(log_level)
	if err != nil {
		level = zerolog.InfoLevel // Default to InfoLevel
	}

	logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(level)
}

// Info logs an info level message
func Info(msg string) {
	logger.Info().Msg(msg)
}

// Infof logs an info level message with formatting
func Infof(format string, v ...interface{}) {
	logger.Info().Msgf(format, v...)
}

// Debug logs a debug level message
func Debug(msg string) {
	logger.Debug().Msg(msg)
}

// Fatal logs a fatal level message and then exits the program
func Fatal(msg string) {
	logger.Fatal().Msg(msg)
}

// Fatalf logs a fatal level message with formatting and then exits the program
func Fatalf(format string, v ...interface{}) {
	logger.Fatal().Msgf(format, v...)
}

func Errorf(format string, v ...interface{}) {
	logger.Error().Msgf(format, v...)
}

func Error(msg string) {
	logger.Error().Msg(msg)
}
