package grpc_logf

import (
	ctx_logf "github.com/richard-xtek/go-grpc-micro-kit/grpc-logf/ctx-logf"
	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"github.com/richard-xtek/go-grpc-micro-kit/utils"
	"github.com/golang/protobuf/proto"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	opentracing "github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zap.UnaryServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadUnaryServerInterceptor(logger log.Factory, decider grpc_logging.ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !decider(ctx, info.FullMethod, info.Server) {
			return handler(ctx, req)
		}
		// Use the provided zap.Logger for logging but use the fields from context.
		logEntry := logger.With(append(serverCallFields(info.FullMethod), ctx_logf.TagsToFields(ctx)...)...)
		logProtoMessageAsJson(ctx, logEntry, req, "grpc.request.content", "server request payload logged as grpc.request.content field")
		resp, err := handler(ctx, req)
		if err == nil {
			logProtoMessageAsJson(ctx, logEntry, resp, "grpc.response.content", "server response payload logged as grpc.request.content field")
		}
		return resp, err
	}
}

// PayloadStreamServerInterceptor returns a new server server interceptors that logs the payloads of requests.
//
// This *only* works when placed *after* the `grpc_zap.StreamServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadStreamServerInterceptor(logger log.Factory, decider grpc_logging.ServerPayloadLoggingDecider) grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !decider(stream.Context(), info.FullMethod, srv) {
			return handler(srv, stream)
		}
		logEntry := logger.With(append(serverCallFields(info.FullMethod), ctx_logf.TagsToFields(stream.Context())...)...)
		newStream := &loggingServerStream{ServerStream: stream, logger: logEntry}
		return handler(srv, newStream)
	}
}

// PayloadUnaryClientInterceptor returns a new unary client interceptor that logs the paylods of requests and responses.
func PayloadUnaryClientInterceptor(logger log.Factory, decider grpc_logging.ClientPayloadLoggingDecider) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if !decider(ctx, method) {
			return invoker(ctx, method, req, reply, cc, opts...)
		}
		logEntry := logger.With(newClientLoggerFields(ctx, method)...)
		logProtoMessageAsJson(ctx, logEntry, req, "grpc.request.content", "client request payload logged as grpc.request.content")
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err == nil {
			logProtoMessageAsJson(ctx, logEntry, reply, "grpc.response.content", "client response payload logged as grpc.response.content")
		}
		return err
	}
}

// PayloadStreamClientInterceptor returns a new streaming client interceptor that logs the paylods of requests and responses.
func PayloadStreamClientInterceptor(logger log.Factory, decider grpc_logging.ClientPayloadLoggingDecider) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if !decider(ctx, method) {
			return streamer(ctx, desc, cc, method, opts...)
		}
		logEntry := logger.With(newClientLoggerFields(ctx, method)...)
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		newStream := &loggingClientStream{ClientStream: clientStream, logger: logEntry}
		return newStream, err
	}
}

type loggingClientStream struct {
	grpc.ClientStream
	logger log.Factory
}

func (l *loggingClientStream) SendMsg(m interface{}) error {
	err := l.ClientStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.ClientStream.Context(), l.logger, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func (l *loggingClientStream) RecvMsg(m interface{}) error {
	err := l.ClientStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.ClientStream.Context(), l.logger, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

type loggingServerStream struct {
	grpc.ServerStream
	logger log.Factory ``
}

func (l *loggingServerStream) SendMsg(m interface{}) error {
	err := l.ServerStream.SendMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.ServerStream.Context(), l.logger, m, "grpc.response.content", "server response payload logged as grpc.response.content field")
	}
	return err
}

func (l *loggingServerStream) RecvMsg(m interface{}) error {
	err := l.ServerStream.RecvMsg(m)
	if err == nil {
		logProtoMessageAsJson(l.ServerStream.Context(), l.logger, m, "grpc.request.content", "server request payload logged as grpc.request.content field")
	}
	return err
}

func logProtoMessageAsJson(ctx context.Context, logger log.Factory, pbMsg interface{}, key string, msg string) {
	if p, ok := pbMsg.(proto.Message); ok {
		if isWriteLog(ctx) {
			logger.For(ctx).CheckWrite(zapcore.InfoLevel, msg, zap.Object(key, utils.JsonpbObjectMarshaler{Pb: p}))
		}
	}
}

func isWriteLog(ctx context.Context) bool {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		if val := span.BaggageItem("output_grpc_message"); val != "" {
			return true
		}
	}
	return false
}
