// Package logger provides structured logging with context support.
package logger

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/fiveseconds/server/pkg/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a structured logger wrapper with context support
type Logger struct {
	base        *zap.Logger
	atomicLevel zap.AtomicLevel
	level       atomic.Value // stores current level string
}

// Config holds logger configuration
type Config struct {
	Level       string `yaml:"level"`        // debug/info/warn/error
	Format      string `yaml:"format"`       // json/console
	ServiceName string `yaml:"service_name"`
	Environment string `yaml:"environment"`
}

// New creates a new Logger instance
func New(cfg *Config) (*Logger, error) {
	level := parseLevel(cfg.Level)
	atomicLevel := zap.NewAtomicLevelAt(level)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	format := cfg.Format
	if format == "" {
		format = "json"
	}

	config := zap.Config{
		Level:            atomicLevel,
		Development:      cfg.Environment != "production",
		Encoding:         format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"service": cfg.ServiceName,
			"env":     cfg.Environment,
		},
	}

	base, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	l := &Logger{
		base:        base,
		atomicLevel: atomicLevel,
	}
	l.level.Store(cfg.Level)

	return l, nil
}

// WithContext extracts trace info from context and returns a logger with those fields
func (l *Logger) WithContext(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return l.base
	}

	fields := []zap.Field{}

	if traceID := trace.GetTraceID(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	if sessionID := trace.GetSessionID(ctx); sessionID != "" {
		fields = append(fields, zap.String("session_id", sessionID))
	}
	if userID := trace.GetUserID(ctx); userID > 0 {
		fields = append(fields, zap.Int64("user_id", userID))
	}

	return l.base.With(fields...)
}


// With returns a logger with additional fields
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		base:        l.base.With(fields...),
		atomicLevel: l.atomicLevel,
	}
}

// SetLevel dynamically changes the log level
func (l *Logger) SetLevel(levelStr string) error {
	levelStr = strings.ToLower(levelStr)
	level := parseLevel(levelStr)
	l.atomicLevel.SetLevel(level)
	l.level.Store(levelStr)
	return nil
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() string {
	if v := l.level.Load(); v != nil {
		return v.(string)
	}
	return "info"
}

// Base returns the underlying zap.Logger
func (l *Logger) Base() *zap.Logger {
	return l.base
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.base.Sync()
}

// Info logs a message at info level
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.base.Info(msg, fields...)
}

// Debug logs a message at debug level
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.base.Debug(msg, fields...)
}

// Warn logs a message at warn level
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.base.Warn(msg, fields...)
}

// Error logs a message at error level
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.base.Error(msg, fields...)
}

// Fatal logs a message at fatal level and exits
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.base.Fatal(msg, fields...)
}

// parseLevel converts a string level to zapcore.Level
func parseLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

// ValidLevels returns the list of valid log levels
func ValidLevels() []string {
	return []string{"debug", "info", "warn", "error"}
}

// IsValidLevel checks if a level string is valid
func IsValidLevel(level string) bool {
	level = strings.ToLower(level)
	for _, valid := range ValidLevels() {
		if level == valid {
			return true
		}
	}
	return false
}
