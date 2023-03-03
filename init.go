package log

import (
	"context"
	"net/url"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type TraceIDType string

const TraceID TraceIDType = "x-request-id"

type LogOutputType uint8

type EnvType string

const (
	// StrLvlDebug debug level
	StrLvlDebug = "DEBUG"
	// StrLvlInfo info level
	StrLvlInfo = "INFO"
	// StrLvlWarn warning level
	StrLvlWarn = "WARN"
	// StrLvlError error level
	StrLvlError = "ERROR"
	// StrlvlCrit critical level
	StrlvlCrit = "CRIT"

	//运行环境
	EnvDebug   EnvType = "DEBUG"
	EnvDev     EnvType = "DEV"
	EnvTest    EnvType = "TEST"
	EnvPre     EnvType = "PRE"
	EnvProd    EnvType = "PROD"
	EnvRelease EnvType = "RELEASE"

	timeFormat = "2006-01-02 15:04:05.000 MST"

	OnlyOutputLog    LogOutputType = 1
	OnlyOutputStdout LogOutputType = 2
)

var (
	logger           *zap.Logger
	atomicZapLeveler zap.AtomicLevel
)

// InitLogger
func InitLogger(componentName string, disableStacktrace bool, runMode EnvType, outputType LogOutputType, fileName ...string) {
	InitLoggerWithSample(componentName, disableStacktrace, runMode, outputType, nil, fileName...)
}

var globalComponentName string

func InitLoggerWithSample(componentName string, disableStacktrace bool, runMode EnvType, outputType LogOutputType, samplingConfig *zap.SamplingConfig, fileName ...string) {
	var err error
	// reset logger
	Exit()
	dev, zapLogLevel := runModeToEnv(runMode)
	encodeCfg := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "log_level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		EncodeTime: func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
			encoder.AppendString(t.Format(timeFormat))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	atomicZapLeveler = zap.NewAtomicLevelAt(zapLogLevel)
	if samplingConfig == nil {
		samplingConfig = &zap.SamplingConfig{
			Initial:    1000,
			Thereafter: 100,
		}
	}
	globalComponentName = componentName
	cfg := zap.Config{
		Level:             atomicZapLeveler,
		Development:       dev,
		DisableCaller:     false,
		DisableStacktrace: disableStacktrace,
		Encoding:          "json", // json or console
		EncoderConfig:     encodeCfg,
		InitialFields:     map[string]interface{}{"component": componentName},
		Sampling:          samplingConfig,
	}
	cfg.OutputPaths, cfg.ErrorOutputPaths = initLogOutput(outputType, runMode, componentName, fileName...)

	if logger, err = cfg.Build(zap.AddCallerSkip(1)); err != nil {
		panic(err)
	}

}

func initLogOutput(outputType LogOutputType, runMode EnvType, componentName string, fileName ...string) (outputPaths []string, errorOutputPaths []string) {
	if outputType&OnlyOutputLog == OnlyOutputLog {
		initFileLogger(runMode, componentName, fileName...) //initFileLogger
		outputPaths = append(outputPaths, "lumberjack:test.log")
		errorOutputPaths = append(errorOutputPaths, "lumberjack:test.log")
		zap.RegisterSink("lumberjack", func(u *url.URL) (zap.Sink, error) {
			return lumberjackSink{
				Roller: lumlog,
			}, nil
		})
	}
	if outputType&OnlyOutputStdout == OnlyOutputStdout {
		outputPaths = append(outputPaths, "stdout")
		errorOutputPaths = append(errorOutputPaths, "stdout")
	}
	return
}

// Logger allow other usage
func Logger() *zap.Logger {
	return logger
}

func Exit() {
	if logger != nil {
		_ = logger.Sync()
	}

}

func toZapLevel(levelStr string) zapcore.Level {
	switch levelStr {
	case StrLvlDebug:
		return zapcore.DebugLevel
	case StrLvlInfo:
		return zapcore.InfoLevel
	case StrLvlWarn:
		return zapcore.WarnLevel
	case StrLvlError:
		return zapcore.ErrorLevel
	case StrlvlCrit:
		return zapcore.PanicLevel
	default:
		logger.Warn("level str to zap unknown level", zap.String("level_string", levelStr))
		return zapcore.InfoLevel
	}
}

func runModeToEnv(runMode EnvType) (bool, zapcore.Level) {
	switch runMode {
	case EnvDebug, EnvTest, EnvDev:
		return true, zapcore.DebugLevel
	}
	return false, zapcore.InfoLevel
}

func genTraceIDZap(ctx context.Context) zap.Field {
	return zap.Reflect("x_request_id", ctx.Value(TraceID))
}
