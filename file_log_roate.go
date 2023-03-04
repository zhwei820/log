package log

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/natefinch/lumberjack/v3"
)

var lumlog *lumberjack.Roller

type lumberjackSink struct {
	*lumberjack.Roller
}

var onceLum = &sync.Once{}

// Sync implements zap.Sink. The remaining methods are implemented
// by the embedded *lumberjack.Logger.
func (lumberjackSink) Sync() error { return nil }

func initFileLogger(runMode RunModeType, componentName string, fileName ...string) {
	hostname, _ := os.Hostname()
	fileNameStr := "logs/" + hostname + ".log"
	if len(fileName) > 0 {
		fileNameStr = fileName[0]
	}

	onceLum.Do(func() {
		lumlog, _ = lumberjack.NewRoller(
			fileNameStr,
			100000,
			&lumberjack.Options{},
		)
		go func() {
			var hourSec int64 = 3600
			leftTs := hourSec - time.Now().Unix()%hourSec - 1 // rotate at 59 sec
			timer := time.NewTimer(time.Duration(leftTs) * time.Second)
			for range timer.C {
				if err := lumlog.Rotate(); err != nil {
					WarnZ(context.Background(), "log rotate error")
				}
				timer.Reset(time.Duration(hourSec) * time.Second)
			}
		}()
	})
}
