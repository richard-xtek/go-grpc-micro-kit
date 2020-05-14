package dialer

import (
	"context"
	"net"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	opentracing "github.com/opentracing/opentracing-go"
	grpc_logf "github.com/richard-xtek/go-grpc-micro-kit/grpc-logf"
	logf "github.com/richard-xtek/go-grpc-micro-kit/log"
	wonaming "github.com/wothing/wonaming/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// NewGrpcClientConsul return new client connection
// consulAddress - consul server
// serviceName - service name register in consul
func NewGrpcClientConsul(consulAddress string, serviceName string, tracer opentracing.Tracer, logger logf.Factory, clientOpts ...ClientOption) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	var options clientOptions

	for _, option := range clientOpts {
		option.apply(&options)
	}

	if options.proxyDialer != nil {
		proxyDialer := func(ctx context.Context, addr string) (conn net.Conn, err error) {
			return options.proxyDialer.Dial("tcp", addr)
		}
		opts = append(opts, grpc.WithContextDialer(proxyDialer))
	}

	alwaysLoggingDeciderClient := func(ctx context.Context, fullMethodName string) bool { return true }

	optsRetry := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(50 * time.Millisecond)),
		grpc_retry.WithCodes(codes.Unavailable),
		grpc_retry.WithMax(3),
		grpc_retry.WithPerRetryTimeout(3 * time.Second),
	}

	opts = append(opts,
		grpc.WithDefaultCallOptions(grpc.FailFast(false)),
	)

	sIntOpt := grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
		StreamClientInterceptor(),
		grpc_opentracing.StreamClientInterceptor(grpc_opentracing.WithTracer(tracer)),
		grpc_prometheus.StreamClientInterceptor,
		grpc_logf.StreamClientInterceptor(logger),
		grpc_logf.PayloadStreamClientInterceptor(logger, alwaysLoggingDeciderClient),
		grpc_retry.StreamClientInterceptor(optsRetry...),
	))

	opts = append(opts, sIntOpt)

	grpc_prometheus.EnableClientHandlingTimeHistogram()

	uIntOpt := grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
		UnaryClientInterceptor(),
		grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(tracer)),
		grpc_prometheus.UnaryClientInterceptor,
		grpc_logf.UnaryClientInterceptor(logger),
		grpc_logf.PayloadUnaryClientInterceptor(logger, alwaysLoggingDeciderClient),
		grpc_retry.UnaryClientInterceptor(optsRetry...),
	))

	// consule
	r := wonaming.NewResolver(serviceName)
	b := grpc.RoundRobin(r)

	opts = append(opts, uIntOpt, grpc.WithInsecure(), grpc.WithBalancer(b))

	conn, err := grpc.Dial(consulAddress, opts...)

	return conn, err
}
