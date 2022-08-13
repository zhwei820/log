package log

import (
	"context"
	"os"
	"runtime"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var lumlog *lumberjack.Logger

type lumberjackSink struct {
	*lumberjack.Logger
}

var onceLum = &sync.Once{}

// Sync implements zap.Sink. The remaining methods are implemented
// by the embedded *lumberjack.Logger.
func (lumberjackSink) Sync() error { return nil }

func initFileLogger(runMode string, componentName string, fileName ...string) {
	hostname, _ := os.Hostname()
	fileNameStr := "logs/" + hostname + ".log"
	if len(fileName) > 0 {
		fileNameStr = fileName[0]
	}

	if (runMode == EnvDev || runMode == EnvTest) && runtime.GOOS != "windows" {
		fileNameStr = "/dev/null" // dev / test 环境日志输入到/dev/null 不支持windows
	}
	onceLum.Do(func() {
		lumlog = &lumberjack.Logger{
			Filename:  fileNameStr,
			MaxSize:   100, // megabytes
			LocalTime: true,
			// MaxBackups: 30, // number of log files // 由运维来处理日志清理
			// MaxAge:     15, // days
		}
		go func() {
			var hourSec int64 = 3600
			leftTs := hourSec - time.Now().Unix()%hourSec -1 // rotate at 59 sec
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
