package grpc_logf

import (
	ctx_logf "github.com/richard-xtek/go-grpc-micro-kit/grpc-logf/ctx-logf"
	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
)

// AddFields adds zap fields to the logger.
// Deprecated: should use the ctxzap.AddFields instead
func AddFields(ctx context.Context, fields ...zapcore.Field) {
	ctx_logf.AddFields(ctx, fields...)
}

// Extract takes the call-scoped Logger from grpc_zap middleware.
// Deprecated: should use the ctxzap.Extract instead
func Extract(ctx context.Context) log.Factory {
	return ctx_logf.Extract(ctx)
}
