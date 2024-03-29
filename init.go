package log

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/TheZeroSlave/zapsentry"
	"github.com/gin-gonic/gin"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type TraceIDType string

const TraceID TraceIDType = "x-request-id"

type LogOutputType uint8

type RunModeType string

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
	RunModeDebug RunModeType = "DEBUG"
	RunModeTest  RunModeType = "TEST"
	RunModeProd  RunModeType = "PROD"

	timeFormat = "2006-01-02 15:04:05.000 MST"

	OnlyOutputLog    LogOutputType = 1
	OnlyOutputStdout LogOutputType = 2
)

/*
	case DebugMode:
	case ReleaseMode:
	case TestMode:
*/

func (t RunModeType) ToGinRunMode() string {
	switch t {
	case RunModeDebug:
		return gin.DebugMode
	case RunModeTest:
		return gin.DebugMode
	case RunModeProd:
		return gin.DebugMode
	}
	panic("")

}

var (
	logger           *zap.Logger
	atomicZapLeveler zap.AtomicLevel
)

// InitLogger
func InitLogger(componentName string, disableStacktrace bool, runMode RunModeType, outputType LogOutputType, fileName ...string) {
	InitLoggerWithSample(componentName, disableStacktrace, runMode, outputType, "", nil, fileName...)
}

// InitLogger
func InitSentryLogger(componentName string, disableStacktrace bool, runMode RunModeType, outputType LogOutputType, sentryDsn string, fileName ...string) {
	InitLoggerWithSample(componentName, disableStacktrace, runMode, outputType, sentryDsn, nil, fileName...)
}

var globalComponentName string

func InitLoggerWithSample(componentName string, disableStacktrace bool, runMode RunModeType, outputType LogOutputType, sentryDsn string, samplingConfig *zap.SamplingConfig, fileName ...string) {
	var err error
	// reset logger
	Exit()
	dev, zapLogLevel := runModeToLogLevel(runMode)
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
	if sentryDsn != "" {
		fmt.Println("===>sentryDsn", sentryDsn)
		sentryClient := SentryClient(sentryDsn)

		// Setup zapsentry
		core, err := zapsentry.NewCore(zapsentry.Configuration{
			Level: zapcore.ErrorLevel, // when to send message to sentry
			// EnableBreadcrumbs: true,               // enable sending breadcrumbs to Sentry
			BreadcrumbLevel: zapcore.ErrorLevel, // at what level should we sent breadcrumbs to sentry
			Tags: map[string]string{
				"component": "system",
			},
		}, zapsentry.NewSentryClientFromClient(sentryClient))
		if err != nil {
			log.Fatal(err)
		}
		logger = zapsentry.AttachCoreToLogger(core, logger)
	}
}

func initLogOutput(outputType LogOutputType, runMode RunModeType, componentName string, fileName ...string) (outputPaths []string, errorOutputPaths []string) {
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

func runModeToLogLevel(runMode RunModeType) (bool, zapcore.Level) {
	switch runMode {
	case RunModeDebug, RunModeTest:
		return true, zapcore.DebugLevel
	case RunModeProd:
		return false, zapcore.InfoLevel
	}
	panic("runMode must be DEBUG,TEST,PROD")
}

func genTraceIDZap(ctx context.Context) zap.Field {
	return zap.Reflect("x_request_id", ctx.Value(TraceID))
}
