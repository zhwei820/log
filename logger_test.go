package log

import (
	"context"
	"fmt"
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
	InitLogger("risk", false, EnvTest, JsonEncoder, OnlyOutputStdout)
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
	InitLogger("bench", true, "TEST", "json", OnlyOutputStdout)
	InfoZ(genWithTraceCtx(), "info test", zap.String("test", "info"))
}
func TestInfoZProd(t *testing.T) {
	InitLogger("bench", true, "RELEASE", "json", OnlyOutputStdout)
	InfoZ(genWithTraceCtx(), "info test", zap.String("test", "info"))
}

func TestDebugZ(t *testing.T) {
	InitLogger("bench", true, "RELEASE", "json", OnlyOutputStdout)
	DebugZ(genWithTraceCtx(), "debug test", zap.String("test", "debug"))
}

func genWithTraceCtx() context.Context {
	return context.WithValue(context.TODO(), TraceID, uuid.NewString())
}

func BenchmarkLogInfoZ(b *testing.B) {
	InitLogger("bench", true, "DEBUG", "json", OnlyOutputLog)
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

/*
go test -benchmem -run=^$ -bench BenchmarkLog github.com/zhwei820/log -v -count=1 -failfast|grep Bench
BenchmarkLogInfoZ
BenchmarkLogInfoZ-12             2968312               438.8 ns/op           387 B/op          2 allocs/op
BenchmarkLogRaw
BenchmarkLogRaw-12               4444786               290.1 ns/op           195 B/op          1 allocs/op
*/

func BenchmarkStrConcat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = "logs/" + "componentName" + "-" + "res" + "-log.log"
	}
}

/*
BenchmarkStrConcat
BenchmarkStrConcat-4   	1000000000	         0.3129 ns/op	       0 B/op	       0 allocs/op
*/
func BenchmarkStrFmt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("logs/%s-%s-log.log", "componentName", "res")
	}
}

/*
BenchmarkStrFmt-4
 8570912	       140.2 ns/op	      32 B/op	       1 allocs/op
*/

func TestLogRotate(t *testing.T) {
	InitLogger("bench", true, "DEBUG", "json", OnlyOutputLog)
	ctx := context.Background()

	for ii := 0; ii < 10; ii++ {
		go func() {
			for i := 0; i < 1e8; i++ {
				InfoZ(ctx, "info test", zap.String("test", "info"), zap.Int("int", i), zap.Any("out", map[string]interface{}{"df": 34, "aa": "b"}))
			}
		}()
	}
	select {} // cancel by key board
}

func TestLogRotateManual(t *testing.T) {
	InitLogger("bench", true, "DEBUG", "json", OnlyOutputLog)
	ctx := context.Background()

	InfoZ(ctx, "info test", zap.String("test", "info"), zap.Int("int", 0), zap.Any("out", map[string]interface{}{"df": 34, "aa": "b"}))
	lumlog.Rotate()
	InfoZ(ctx, "info test", zap.String("test", "info"), zap.Int("int", 1), zap.Any("out", map[string]interface{}{"df": 34, "aa": "b"}))
	select {}
}

func Test_LeftTs(t *testing.T) {
	var daySec int64 = 86400
	leftTs := daySec - (time.Now().Unix()+8*3600)%daySec // fix +8 zone
	fmt.Println("leftTs", float64(leftTs)/3600)
}
