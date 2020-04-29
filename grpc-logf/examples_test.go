package grpc_logf_test

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	ctx_zap "github.com/grpc-ecosystem/go-grpc-middleware/tags/zap"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
	grpc_logf "github.com/richard-xtek/go-grpc-micro-kit/grpc-logf"
	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

var (
	zapLogger  log.Factory
	customFunc grpc_logf.CodeToLevel
)

// Initialization shows a relatively complex initialization sequence.
func Example_initialization() {
	// Shared options for the logger, with a custom gRPC code to log level function.
	opts := []grpc_logf.Option{
		grpc_logf.WithLevels(customFunc),
	}
	// Make sure that log statements internal to gRPC library are logged using the zapLogger as well.
	grpc_logf.ReplaceGrpcLogger(zapLogger)
	// Create a server, make sure we put the grpc_ctxtags context before everything else.
	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logf.UnaryServerInterceptor(zapLogger, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_logf.StreamServerInterceptor(zapLogger, opts...),
		),
	)
}

// Initialization shows an initialization sequence with the duration field generation overridden.
func Example_initializationWithDurationFieldOverride() {
	opts := []grpc_logf.Option{
		grpc_logf.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grpc.time_ns", duration.Nanoseconds())
		}),
	}

	_ = grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logf.UnaryServerInterceptor(zapLogger, opts...),
		),
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_logf.StreamServerInterceptor(zapLogger, opts...),
		),
	)
}

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func ExampleExtract_unary() {
	_ = func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields the ctxtags of the request which will be added to all extracted loggers.
		grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)

		// Extract a single request-scoped zap.Logger and log messages. (containing the grpc.xxx tags)
		l := ctx_zap.Extract(ctx)
		l.Info("some ping")
		l.Info("another ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
}

func Example_initializationWithDecider() {
	opts := []grpc_logf.Option{
		grpc_logf.WithDecider(func(fullMethodName string, err error) bool {
			// will not log gRPC calls if it was a call to healthcheck and no error was raised
			if err == nil && fullMethodName == "foo.bar.healthcheck" {
				return false
			}

			// by default everything will be logged
			return true
		}),
	}

	_ = []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_logf.StreamServerInterceptor(log.NewFactory(zap.NewNop()), opts...)),
		grpc_middleware.WithUnaryServerChain(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_logf.UnaryServerInterceptor(log.NewFactory(zap.NewNop()), opts...)),
	}
}
