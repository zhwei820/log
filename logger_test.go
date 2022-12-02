package log

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func throwErr1() error {
	err := errors.WithStack(errors.New("fake error1"))
	return err
}

func throwErr2() error {
	err := errors.Wrap(throwErr1(), "fake error2")
	return err
}

func TestFormatEnv(t *testing.T) {
	_, err := FormatEnv("info")
	if err == ErrInvalidEnv {
		t.Log("ok, get ErrInvalidEnv")
	}
}

func TestInitLogger(t *testing.T) {
	InitLogger("risk", false, EnvTest, OnlyOutputStdout)
	if logger == nil {
		t.Error("init logger failed")
	}
}

func TestSetLvl(t *testing.T) {
	oldLvl := atomicZapLeveler.Level()
	SetLevel(StrLvlWarn)
	assert.Equal(t, zap.WarnLevel, atomicZapLeveler.Level())
	assert.False(t, Logger().Core().Enabled(zap.DebugLevel))
	assert.True(t, Logger().Core().Enabled(zap.WarnLevel))
	atomicZapLeveler.SetLevel(oldLvl)
}

func TestErrorZ(t *testing.T) {
	ErrorZ(genWithTraceCtx(), "error test with zap", zap.String("error", "err"))
}

func TestWarnZ(t *testing.T) {
	WarnZ(genWithTraceCtx(), "warn test", zap.String("test", "warn"))
}

func TestInfoZ(t *testing.T) {
	InitLogger("bench", true, "TEST", OnlyOutputStdout)
	InfoZ(genWithTraceCtx(), "info test", zap.String("test", "info"))
}
func TestInfoZProd(t *testing.T) {
	InitLogger("bench", true, "RELEASE", OnlyOutputStdout)
	InfoZ(genWithTraceCtx(), "info test", zap.String("test", "info"))
}

func TestDebugZ(t *testing.T) {
	InitLogger("bench", true, "RELEASE", OnlyOutputStdout)
	DebugZ(genWithTraceCtx(), "debug test", zap.String("test", "debug"))
}

func genWithTraceCtx() context.Context {
	return context.WithValue(context.TODO(), TraceID, uuid.NewString())
}

func BenchmarkLogInfoZ(b *testing.B) {
	InitLogger("bench", true, "DEBUG", OnlyOutputLog)
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		InfoZ(ctx, "info test", zap.String("test", "info"), zap.Int("int", i), zap.Any("out", map[string]interface{}{"df": 34, "aa": "b"}))
	}
}

func BenchmarkLogRaw(b *testing.B) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	for i := 0; i < b.N; i++ {
		logger.Info("failed to fetch URL",
			// Structured context as strongly typed Field values.
			zap.String("url", "url"),
			zap.Int("attempt", i),
			zap.Duration("backoff", time.Second),
		)
	}
}
