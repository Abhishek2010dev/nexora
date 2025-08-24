package middleware

import (
	"fmt"
	"time"

	"github.com/Abhishek2010dev/nexora"
)

// LoggerConfig defines options for the logger middleware
type LoggerConfig struct {
	SkipFunc      func(c *nexora.Context) bool // Skip logging for certain requests
	LogLatency    bool                         // Include latency
	LogIP         bool                         // Include client IP
	LogUserAgent  bool                         // Include user agent
	MessageFormat string                       // Custom message format
}

// DefaultLoggerConfig returns sensible defaults
func DefaultLoggerConfig() *LoggerConfig {
	return &LoggerConfig{
		SkipFunc:      nil,
		LogLatency:    true,
		LogIP:         true,
		LogUserAgent:  false,
		MessageFormat: "",
	}
}

// Logger creates a configurable logger middleware
func Logger(config ...*LoggerConfig) nexora.Handler {
	var cfg *LoggerConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = DefaultLoggerConfig()
	}

	return func(c *nexora.Context) error {
		if cfg.SkipFunc != nil && cfg.SkipFunc(c) {
			return c.Next()
		}

		start := time.Now()
		err := c.Next()
		latency := time.Since(start)

		// Collect values
		method := c.Method()
		path := c.Path()
		status := c.ResponseWriter().Status()
		ip := c.IP()
		ua := c.GetHeader("User-Agent")

		// Choose log level
		logFn := nexora.Info
		switch {
		case status >= 500:
			logFn = nexora.Error
		case status >= 400:
			logFn = nexora.Warn
		default:
			logFn = nexora.Info
		}

		// Use custom message format if given
		if cfg.MessageFormat != "" {
			msg := replaceLogPlaceholders(cfg.MessageFormat, map[string]string{
				"method":  method,
				"path":    path,
				"status":  fmt.Sprintf("%d", status),
				"latency": formatLatency(latency),
				"ip":      ip,
				"ua":      ua,
			})
			logFn(msg)
			return err
		}

		// Always structured logs
		fields := []any{
			"method", method,
			"path", path,
			"status", status,
		}
		if cfg.LogLatency {
			fields = append(fields, "latency", formatLatency(latency))
		}
		if cfg.LogIP {
			fields = append(fields, "ip", ip)
		}
		if cfg.LogUserAgent {
			fields = append(fields, "ua", ua)
		}
		fields = append(fields, "service", "nexora")

		logFn("request", fields...)
		return err
	}
}

// formatLatency scales ns → µs → ms → s
func formatLatency(d time.Duration) string {
	if d > time.Second {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	if d > time.Millisecond {
		return fmt.Sprintf("%.2fms", float64(d.Microseconds())/1000.0)
	}
	if d > time.Microsecond {
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000.0)
	}
	return fmt.Sprintf("%dns", d.Nanoseconds())
}

// replaceLogPlaceholders replaces {placeholder} with actual values
func replaceLogPlaceholders(tmpl string, values map[string]string) string {
	out := tmpl
	for k, v := range values {
		placeholder := fmt.Sprintf("{%s}", k)
		out = replaceAll(out, placeholder, v)
	}
	return out
}

// Simple replaceAll (no regex for performance)
func replaceAll(s, old, new string) string {
	for {
		idx := findIndex(s, old)
		if idx == -1 {
			return s
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
}

// findIndex finds substring position
func findIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
