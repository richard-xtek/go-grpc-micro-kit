package grpc_logf

import (
	"path"
	"time"

	ctx_logf "github.com/richard-xtek/go-grpc-micro-kit/grpc-logf/ctx-logf"
	"github.com/richard-xtek/go-grpc-micro-kit/log"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	// SystemField is used in every log statement made through grpc_zap. Can be overwritten before any initialization code.
	SystemField = zap.String("system", "grpc")

	// ServerField is used in every server-side log statement made through grpc_zap.Can be overwritten before initialization.
	ServerField = zap.String("span.kind", "server")
)

// UnaryServerInterceptor returns a new unary server interceptors that adds zap.Logger to the context.
func UnaryServerInterceptor(logger log.Factory, opts ...Option) grpc.UnaryServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()
		newCtx := newLoggerForCall(ctx, logger, info.FullMethod, startTime)

		resp, err := handler(newCtx, req)
		if !o.shouldLog(info.FullMethod, err) {
			return resp, err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)

		if isWriteLog(newCtx) {
			// re-extract logger from newCtx, as it may have extra fields that changed in the holder.
			ctx_logf.Extract(newCtx).For(newCtx).CheckWrite(level, "finished unary call with code "+code.String(),
				zap.Error(err),
				zap.String("grpc.code.2", code.String()),
				o.durationFunc(time.Since(startTime)),
			)
		}

		return resp, err
	}
}

// StreamServerInterceptor returns a new streaming server interceptor that adds zap.Logger to the context.
func StreamServerInterceptor(logger log.Factory, opts ...Option) grpc.StreamServerInterceptor {
	o := evaluateServerOpt(opts)
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()
		newCtx := newLoggerForCall(stream.Context(), logger, info.FullMethod, startTime)
		wrapped := grpc_middleware.WrapServerStream(stream)
		wrapped.WrappedContext = newCtx

		err := handler(srv, wrapped)
		if !o.shouldLog(info.FullMethod, err) {
			return err
		}
		code := o.codeFunc(err)
		level := o.levelFunc(code)

		if isWriteLog(newCtx) {
			// re-extract logger from newCtx, as it may have extra fields that changed in the holder.
			ctx_logf.Extract(newCtx).For(newCtx).CheckWrite(level, "finished streaming call with code "+code.String(),
				zap.Error(err),
				zap.String("grpc.code", code.String()),
				o.durationFunc(time.Since(startTime)),
			)
		}

		return err
	}
}

func serverCallFields(fullMethodString string) []zapcore.Field {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return []zapcore.Field{
		SystemField,
		ServerField,
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}
}

func newLoggerForCall(ctx context.Context, logger log.Factory, fullMethodString string, start time.Time) context.Context {
	f := ctx_logf.TagsToFields(ctx)
	f = append(f, zap.String("grpc.start_time", start.Format(time.RFC3339)))
	if d, ok := ctx.Deadline(); ok {
		f = append(f, zap.String("grpc.request.deadline", d.Format(time.RFC3339)))
	}
	callLog := logger.With(append(f, serverCallFields(fullMethodString)...)...)

	return ctx_logf.ToContext(ctx, callLog)
}
