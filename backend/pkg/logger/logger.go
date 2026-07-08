// Package logger provides a production-ready structured logger built on Zap.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps *zap.SugaredLogger so we can satisfy the interfaces.Logger contract.
type Logger struct {
	sugar *zap.SugaredLogger
}

// New builds a Logger with the given level ("debug","info","warn","error").
// In production it uses JSON encoding; in development, console encoding.
func New(level string, development bool) (*Logger, error) {
	var cfg zap.Config
	if development {
		cfg = zap.NewDevelopmentConfig()
	} else {
		cfg = zap.NewProductionConfig()
	}

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(zapLevel)

	z, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}
	return &Logger{sugar: z.Sugar()}, nil
}

// Sync flushes any buffered log entries. Call on shutdown.
func (l *Logger) Sync() { _ = l.sugar.Sync() }

func (l *Logger) Info(msg string, args ...interface{})  { l.sugar.Infow(msg, args...) }
func (l *Logger) Warn(msg string, args ...interface{})  { l.sugar.Warnw(msg, args...) }
func (l *Logger) Error(msg string, args ...interface{}) { l.sugar.Errorw(msg, args...) }
func (l *Logger) Debug(msg string, args ...interface{}) { l.sugar.Debugw(msg, args...) }
func (l *Logger) Fatal(msg string, args ...interface{}) { l.sugar.Fatalw(msg, args...) }

// With returns a child logger with an additional key-value field.
func (l *Logger) With(key string, value interface{}) *Logger {
	return &Logger{sugar: l.sugar.With(key, value)}
}
