package log

import (
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func addFields(enc zapcore.ObjectEncoder, fields []zapcore.Field) {
	for i := range fields {
		fields[i].AddTo(enc)
	}
}

func PlainError(err error) zap.Field {
	return zap.Error(fmt.Errorf("%s", err.Error()))
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}
