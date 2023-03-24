package log

import (
	"context"
	"testing"
	"time"

	"github.com/zhwei820/errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

func TestInitSentryLogger(t *testing.T) {

	InitSentryLogger("log-01", false, RunModeDebug, 3, "127.0.0.1:8031", "logs/log-01.log")
	ctx := context.WithValue(context.Background(), TraceID, uuid.NewString())

	for ii := 0; ii < 3; ii++ {
		ErrorZ(ctx, "info test", zap.String("test", "info"), zap.Int("int", 100), zap.Any("out", map[string]interface{}{"df": 34, "aa": "b"}), zap.Error(errors.Errorf("test")))
	}
	time.Sleep(3 * time.Second)
}
