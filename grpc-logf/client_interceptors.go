// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logf

import (
	"path"
	"time"

	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	// ClientField is used in every client-side log statement made through grpc_zap. Can be overwritten before initialization.
	ClientField = zap.String("span.kind", "client")
)

// UnaryClientInterceptor returns a new unary client interceptor that optionally logs the execution of external gRPC calls.
func UnaryClientInterceptor(logger log.Factory, opts ...Option) grpc.UnaryClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		fields := newClientLoggerFields(ctx, method)
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		logFinalClientLine(ctx, o, logger.With(fields...), startTime, err, "finished client unary call")
		return err
	}
}

// StreamClientInterceptor returns a new streaming client interceptor that optionally logs the execution of external gRPC calls.
func StreamClientInterceptor(logger log.Factory, opts ...Option) grpc.StreamClientInterceptor {
	o := evaluateClientOpt(opts)
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		fields := newClientLoggerFields(ctx, method)
		startTime := time.Now()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		logFinalClientLine(ctx, o, logger.With(fields...), startTime, err, "finished client streaming call")
		return clientStream, err
	}
}

func logFinalClientLine(ctx context.Context, o *options, logger log.Factory, startTime time.Time, err error, msg string) {
	if isWriteLog(ctx) {
		code := o.codeFunc(err)
		level := o.levelFunc(code)
		logger.For(ctx).CheckWrite(level, msg,
			zap.Error(err),
			zap.String("grpc.code", code.String()),
			o.durationFunc(time.Now().Sub(startTime)),
		)
	}
}

func newClientLoggerFields(ctx context.Context, fullMethodString string) []zapcore.Field {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return []zapcore.Field{
		SystemField,
		ClientField,
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}
}
