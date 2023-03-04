package log

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestInitSentryLogger(t *testing.T) {

	InitSentryLogger("log-01", false, RunModeDebug, 3, "127.0.0.1:8031", "logs/log-01.log")
	ctx := context.Background()

	for ii := 0; ii < 100; ii++ {
		ErrorZ(ctx, "info test", zap.String("test", "info"), zap.Int("int", 100), zap.Any("out", map[string]interface{}{"df": 34, "aa": "b"}))
	}
	time.Sleep(3 * time.Second)
}
