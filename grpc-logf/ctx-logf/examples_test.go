package ctx_logf_test

import (
	"context"

	"github.com/richard-xtek/go-grpc-micro-kit/log"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	pb_testproto "github.com/grpc-ecosystem/go-grpc-middleware/testing/testproto"
)

var zapLogger log.Factory

// Simple unary handler that adds custom fields to the requests's context. These will be used for all log statements.
func ExampleExtract_unary() {
	_ = func(ctx context.Context, ping *pb_testproto.PingRequest) (*pb_testproto.PingResponse, error) {
		// Add fields the ctxtags of the request which will be added to all extracted loggers.
		grpc_ctxtags.Extract(ctx).Set("custom_tags.string", "something").Set("custom_tags.int", 1337)

		// Extract a single request-scoped zap.Logger and log messages.
		l := ctx_logf.Extract(ctx)
		l.For(ctx).Info("some ping")
		l.For(ctx).Info("another ping")
		return &pb_testproto.PingResponse{Value: ping.Value}, nil
	}
}
