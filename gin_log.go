package log

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestLog 输入及输出打印
func RequestLog() gin.HandlerFunc {
	return func(g *gin.Context) {
		traceID := g.Request.Header.Get("X-Request-Id")
		if len(traceID) == 0 {
			traceID = uuid.NewString()
		}
		ctx := context.WithValue(g.Request.Context(), TraceID, traceID)

		g.Request = g.Request.WithContext(ctx)

		defer ginRecover(ctx, g)

		// 打印处理日志
		startTime := time.Now()
		var err error
		var bodyBytes []byte // body中传入的参数
		path := g.Request.URL.Path
		queryParams := g.Request.URL.RawQuery
		// 获取请求BODY中的数据（POST和PUT请求中可能包含该参数）
		contentType := g.Request.Header.Get("Content-Type")
		if len(contentType) == 0 || contentType == "application/json" {
			if g.Request.Body != nil {
				bodyBytes, err = ioutil.ReadAll(g.Request.Body)
				if err != nil {
					WarnZ(ctx, "ReadAll from path failed", zap.Error(err), zap.String("path", path))
				}
				g.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// 处理用户请求
		g.Next()

		// 打印处理日志

		nowTimestamp := startTime.UnixNano() / 1e6
		InfoZ(ctx, "user request logged",
			zap.String("user_id", g.Request.Header.Get("USER-ID")),
			zap.String("language", g.Request.Header.Get("LANGUAGE-TYPE")),
			zap.Int64("request_receive_time", nowTimestamp),
			zap.String("method", g.Request.Method),
			zap.String("client_ip", g.Request.Header.Get("Client-Ip")),
			zap.String("request_path", path),
			zap.String("request_query_params", queryParams),
			zap.String("from_service_name", g.Request.Header.Get("X-Service-Name")),
			zap.ByteString("request_body", bodyBytes),
			zap.Int64("response_time", time.Now().UnixNano()/1e6),
			zap.Duration("elapsed", time.Since(startTime)),
			zap.Int("response_http_status", g.Writer.Status()),
			zap.Int("response_size", g.Writer.Size()),
		)
	}
}

// ginRecover
func ginRecover(ctx context.Context, g *gin.Context) {
	if err := recover(); err != nil {
		// Check for a broken connection, as it is not really a
		// condition that warrants a panic stack trace.
		requestPath := g.Request.URL.Path
		brokenPipe := false
		if ne, ok := err.(*net.OpError); ok {
			if se, ok := ne.Err.(*os.SyscallError); ok {
				if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
					brokenPipe = true
				}
			}
		}
		ErrorZ(ctx, "gin handler unknown error",
			zap.String("user_id", g.Request.Header.Get("USER-ID")),
			zap.String("language", g.Request.Header.Get("LANGUAGE-TYPE")),
			zap.String("method", g.Request.Method),
			zap.String("client_ip", g.Request.Header.Get("Client-Ip")),
			zap.String("request_path", requestPath),
			zap.String("from_service_name", g.Request.Header.Get("X-Service-Name")),
			zap.Int64("response_time", time.Now().UnixNano()/1e6),
			zap.Int("response_http_status", http.StatusInternalServerError), zap.Reflect("error", "error"), zap.Stack("stack"))
		if !brokenPipe {
			g.JSON(http.StatusInternalServerError, nil)
		}
		g.Abort()
	}
}
