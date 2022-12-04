package log

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type TraceIDType string

const TraceID TraceIDType = "x-request-id"

type LogOutputType uint8

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
	EnvDebug   = "DEBUG"
	EnvDev     = "DEV"
	EnvTest    = "TEST"
	EnvPre     = "PRE"
	EnvProd    = "PROD"
	EnvRelease = "RELEASE"

	timeFormat = "2006-01-02 15:04:05.000 MST"

	OnlyOutputLog    LogOutputType = 1
	OnlyOutputStdout LogOutputType = 2
)

var (
	logger            *zap.Logger
	atomicZapLeveler  zap.AtomicLevel
	ErrInvalidEncoder = errors.New("console encoder can only be used in dev and debug environment")
	ErrInvalidEnv     = errors.New("invalid env")
)

// InitLogger
func InitLogger(componentName string, disableStacktrace bool, runMode string, outputType LogOutputType, fileName ...string) {
	InitLoggerWithSample(componentName, disableStacktrace, runMode, "json", outputType, nil, fileName...)
}

var globalComponentName string

func InitLoggerWithSample(componentName string, disableStacktrace bool, runMode string, encoderName string, outputType LogOutputType, samplingConfig *zap.SamplingConfig, fileName ...string) {
	var err error
	// reset logger
	Exit()
	if runMode, err = FormatEnv(runMode); err != nil {
		panic(err)
	}
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
		Encoding:          encoderName, // json or console
		EncoderConfig:     encodeCfg,
		InitialFields:     map[string]interface{}{"component": componentName},
		Sampling:          samplingConfig,
	}
	cfg.OutputPaths, cfg.ErrorOutputPaths = initLogOutput(outputType, runMode, componentName, fileName...)

	if err = checkEncoder(runMode, cfg.Encoding); err != nil {
		panic(err)
	}

	if logger, err = cfg.Build(zap.AddCallerSkip(1)); err != nil {
		panic(err)
	}

}

func initLogOutput(outputType LogOutputType, runMode string, componentName string, fileName ...string) (outputPaths []string, errorOutputPaths []string) {
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

// FormatEnv 用于检查并格式化运行环境配置值
func FormatEnv(env string) (nEnv string, err error) {
	nEnv = strings.ToUpper(env)
	switch nEnv {
	case EnvDebug, EnvDev, EnvTest, EnvPre, EnvProd, EnvRelease:
		return
	default:
		err = ErrInvalidEnv
	}

	return
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

func runModeToEnv(runMode string) (bool, zapcore.Level) {
	runMode = strings.ToUpper(runMode)
	switch runMode {
	case EnvDebug, EnvTest, EnvDev:
		return true, zapcore.DebugLevel
	}
	return false, zapcore.InfoLevel
}

// checkEncoder 用于检查编码器是否适配当前运行环境
func checkEncoder(env string, encoder string) (err error) {
	if (env != EnvDev && env != EnvDebug) && encoder != "json" {
		err = ErrInvalidEncoder
	}
	return
}

func genTraceIDZap(ctx context.Context) zap.Field {
	return zap.Reflect("x_request_id", ctx.Value(TraceID))
}
