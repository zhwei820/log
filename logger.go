package log

import (
	"context"

	"go.uber.org/zap"
)

// SetLevel set log level, change the log level dynamically via HTTP or GRPC
func SetLevel(levelStr string) {
	level := toZapLevel(levelStr)
	if level == atomicZapLeveler.Level() {
		return
	}
	atomicZapLeveler.SetLevel(level)
}

func ErrorZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, genTraceIDZap(ctx))
	logger.Error(msg, fields...)
}

func WarnZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, genTraceIDZap(ctx))
	logger.Warn(msg, fields...)
}

func InfoZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, genTraceIDZap(ctx))
	logger.Info(msg, fields...)
}

func DebugZ(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, genTraceIDZap(ctx))
	logger.Debug(msg, fields...)
}

// =================================================================
func Error(msg string, fields ...zap.Field) {
	fields = append(fields)
	logger.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	fields = append(fields)
	logger.Warn(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	fields = append(fields)
	logger.Info(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	fields = append(fields)
	logger.Debug(msg, fields...)
}
