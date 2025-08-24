package nexora

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger

// LoggerConfig holds Nexora's logger configuration
type LoggerConfig struct {
	Production  bool          // Enable json console output
	Level       zerolog.Level // Minimum log level
	TimeFormat  string        // Time format for logs
	Output      io.Writer     // Where logs are written (default: os.Stdout)
	WithCaller  bool          // Include caller info (file:line)
	ServiceName string        // Optional service or app name
}

// InitDefaultLogger initializes the logger with default configuration.
// This should be called if no custom LoggerConfig is provided.
func InitDefaultLogger() {
	defaultCfg := &LoggerConfig{
		Production:  false,              // Pretty console output by default
		Level:       zerolog.DebugLevel, // Debug level for dev mode
		TimeFormat:  time.RFC3339,       // ISO timestamp
		Output:      os.Stdout,          // Logs to stdout
		WithCaller:  false,              // Show caller file:line
		ServiceName: "nexora",           // Default service name
	}
	initLogger(defaultCfg)
}

func initLogger(cfg *LoggerConfig) {
	// Set default output if not provided
	output := cfg.Output
	if output == nil {
		output = os.Stdout
	}

	// Set default time format
	if cfg.TimeFormat == "" {
		cfg.TimeFormat = time.RFC3339
	}

	zerolog.TimeFieldFormat = cfg.TimeFormat

	// Pretty output in debug mode
	var baseLogger zerolog.Logger
	if cfg.Production {
		baseLogger = zerolog.New(output).With().Timestamp().Logger()
	} else {
		baseLogger = zerolog.New(zerolog.ConsoleWriter{Out: output, TimeFormat: cfg.TimeFormat}).With().Timestamp().Logger()
	}

	// Add service name if provided
	if cfg.ServiceName != "" {
		baseLogger = baseLogger.With().Str("service", cfg.ServiceName).Logger()
	}

	// Add caller info if enabled
	if cfg.WithCaller {
		baseLogger = baseLogger.With().CallerWithSkipFrameCount(3).Logger()
	}

	// Set global log level
	zerolog.SetGlobalLevel(cfg.Level)

	logger = baseLogger
}

// Trace logs a trace-level message
func Trace(msg string, fields ...any) {
	logger.Trace().Fields(fieldsToMap(fields...)).Msg(msg)
}

// Debug logs a debug-level message
func Debug(msg string, fields ...any) {
	logger.Debug().Fields(fieldsToMap(fields...)).Msg(msg)
}

// Info logs an info-level message
func Info(msg string, fields ...any) {
	logger.Info().Fields(fieldsToMap(fields...)).Msg(msg)
}

// Warn logs a warning-level message
func Warn(msg string, fields ...any) {
	logger.Warn().Fields(fieldsToMap(fields...)).Msg(msg)
}

// Error logs an error-level message
func Error(msg string, fields ...any) {
	logger.Error().Fields(fieldsToMap(fields...)).Msg(msg)
}

// Fatal logs a fatal-level message and exits with status 1
func Fatal(msg string, fields ...any) {
	logger.Fatal().Fields(fieldsToMap(fields...)).Msg(msg)
}

// Panic logs a panic-level message and panics
func Panic(msg string, fields ...any) {
	logger.Panic().Fields(fieldsToMap(fields...)).Msg(msg)
}

func fieldsToMap(fields ...any) map[string]any {
	m := make(map[string]any)
	for i := 0; i < len(fields)-1; i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			continue
		}
		m[key] = fields[i+1]
	}
	return m
}
