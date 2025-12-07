// Package logger provides structured logging for Krustron
// Author: Anubhav Gain <anubhavg@infopercept.com>
package logger

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	log  *zap.Logger
	once sync.Once
)

// Config holds logger configuration
type Config struct {
	Level       string `mapstructure:"level" json:"level" yaml:"level"`
	Format      string `mapstructure:"format" json:"format" yaml:"format"` // json or console
	Output      string `mapstructure:"output" json:"output" yaml:"output"` // stdout, stderr, or file path
	Development bool   `mapstructure:"development" json:"development" yaml:"development"`
}

// DefaultConfig returns default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:       "info",
		Format:      "json",
		Output:      "stdout",
		Development: false,
	}
}

// Init initializes the global logger
func Init(cfg *Config) error {
	var err error
	once.Do(func() {
		log, err = NewLogger(cfg)
	})
	return err
}

// NewLogger creates a new zap logger instance
func NewLogger(cfg *Config) (*zap.Logger, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Configure encoder
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
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Configure encoder based on format
	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Configure output
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Output {
	case "stdout":
		writeSyncer = zapcore.AddSync(os.Stdout)
	case "stderr":
		writeSyncer = zapcore.AddSync(os.Stderr)
	default:
		file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writeSyncer = zapcore.AddSync(file)
	}

	// Build core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Build logger options
	opts := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	return zap.New(core, opts...), nil
}

// Get returns the global logger instance
func Get() *zap.Logger {
	if log == nil {
		log, _ = NewLogger(DefaultConfig())
	}
	return log
}

// Sugar returns the global sugared logger
func Sugar() *zap.SugaredLogger {
	return Get().Sugar()
}

// With creates a child logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}

// Named creates a named child logger
func Named(name string) *zap.Logger {
	return Get().Named(name)
}

// Debug logs a debug message
func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Info logs an info message
func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

// Warn logs a warning message
func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

// Error logs an error message
func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

// Panic logs a panic message and panics
func Panic(msg string, fields ...zap.Field) {
	Get().Panic(msg, fields...)
}

// Sync flushes any buffered log entries
func Sync() error {
	return Get().Sync()
}

// Common field constructors for convenience
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

func Time(key string, val time.Time) zap.Field {
	return zap.Time(key, val)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func Err(err error) zap.Field {
	return zap.Error(err)
}

// RequestLogger returns a logger for HTTP requests
func RequestLogger(requestID, method, path, clientIP string) *zap.Logger {
	return Get().With(
		zap.String("request_id", requestID),
		zap.String("method", method),
		zap.String("path", path),
		zap.String("client_ip", clientIP),
	)
}

// ClusterLogger returns a logger for cluster operations
func ClusterLogger(clusterName, operation string) *zap.Logger {
	return Get().With(
		zap.String("cluster", clusterName),
		zap.String("operation", operation),
	)
}

// PipelineLogger returns a logger for pipeline operations
func PipelineLogger(pipelineID, stage string) *zap.Logger {
	return Get().With(
		zap.String("pipeline_id", pipelineID),
		zap.String("stage", stage),
	)
}

// GitOpsLogger returns a logger for GitOps operations
func GitOpsLogger(repo, branch, commit string) *zap.Logger {
	return Get().With(
		zap.String("repository", repo),
		zap.String("branch", branch),
		zap.String("commit", commit),
	)
}
