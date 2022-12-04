package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryClientLogTraceInterceptor(t *testing.T) {
	InitLogger("testgrpc", true, "DEBUG", OnlyOutputStdout)
	ctx := context.WithValue(context.Background(), TraceID, "test")
	err := UnaryClientLogTraceInterceptor(ctx, "POST", nil, nil, nil, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		assert.Equal(t, "test", ctx.Value(TraceID))
		return nil
	})
	assert.Nil(t, err)
	ctx = context.Background()
	err = UnaryClientLogTraceInterceptor(ctx, "POST", nil, nil, nil, func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		assert.NotEqual(t, "test", ctx.Value(TraceID))
		return nil
	})
	assert.Nil(t, err)
}

func TestUnaryServerLogTraceInterceptor(t *testing.T) {
	InitLogger("testgrpc", true, "DEBUG", OnlyOutputStdout)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{string(TraceID): []string{"test"}, "from_service_name": []string{"demo"}})

	_, err := UnaryServerLogTraceInterceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		assert.Equal(t, "test", ctx.Value(TraceID))
		return nil, nil
	})
	assert.Nil(t, err)

	ctx = context.Background()
	_, err = UnaryServerLogTraceInterceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		assert.NotEqual(t, "test", ctx.Value(TraceID))
		return nil, nil
	})
	assert.Nil(t, err)
}
