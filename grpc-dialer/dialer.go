package dialer

import (
	"context"
	"net"

	consul "github.com/hashicorp/consul/api"
	lb "github.com/olivere/grpc/lb/consul"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	opentracing "github.com/opentracing/opentracing-go"
	grpc_logf "github.com/richard-xtek/go-grpc-micro-kit/grpc-logf"
	logf "github.com/richard-xtek/go-grpc-micro-kit/log"
	"google.golang.org/grpc"
)

// NewGrpcClientDialer ...
func NewGrpcClientDialer(consul *consul.Client, tracer opentracing.Tracer, logger logf.Factory) *GRPCClientDialer {
	return &GRPCClientDialer{
		consul: consul,
		tracer: tracer,
		logger: logger,
	}
}

// GRPCClientDialer ...
type GRPCClientDialer struct {
	consul *consul.Client
	tracer opentracing.Tracer
	logger logf.Factory
}

// ConnWithServiceName ...
func (d *GRPCClientDialer) ConnWithServiceName(serviceName string, clientOpts ...ClientOption) (*grpc.ClientConn, error) {
	return NewGrpcClientConsul(d.consul, serviceName, d.tracer, d.logger, clientOpts...)
}

// NewGrpcClientConsul return new client connection
// consulAddress - consul server
// serviceName - service name register in consul
func NewGrpcClientConsul(cc *consul.Client, serviceName string, tracer opentracing.Tracer, logger logf.Factory, clientOpts ...ClientOption) (*grpc.ClientConn, error) {
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

	// optsRetry := []grpc_retry.CallOption{
	// 	grpc_retry.WithBackoff(grpc_retry.BackoffExponential(50 * time.Millisecond)),
	// 	grpc_retry.WithCodes(codes.Unavailable),
	// 	grpc_retry.WithMax(3),
	// 	grpc_retry.WithPerRetryTimeout(3 * time.Second),
	// }

	// opts = append(opts,
	// 	grpc.WithDefaultCallOptions(grpc.FailFast(false)),
	// )

	sIntOpt := grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
		StreamClientInterceptor(),
		grpc_opentracing.StreamClientInterceptor(grpc_opentracing.WithTracer(tracer)),
		grpc_prometheus.StreamClientInterceptor,
		grpc_logf.StreamClientInterceptor(logger),
		grpc_logf.PayloadStreamClientInterceptor(logger, alwaysLoggingDeciderClient),
		// grpc_retry.StreamClientInterceptor(optsRetry...),
	))

	opts = append(opts, sIntOpt)

	grpc_prometheus.EnableClientHandlingTimeHistogram()

	uIntOpt := grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
		UnaryClientInterceptor(),
		grpc_opentracing.UnaryClientInterceptor(grpc_opentracing.WithTracer(tracer)),
		grpc_prometheus.UnaryClientInterceptor,
		grpc_logf.UnaryClientInterceptor(logger),
		grpc_logf.PayloadUnaryClientInterceptor(logger, alwaysLoggingDeciderClient),
		// grpc_retry.UnaryClientInterceptor(optsRetry...),
	))

	// consule
	r, err := lb.NewResolver(cc, serviceName, "")
	if err != nil {
		return nil, err
	}

	b := grpc.RoundRobin(r)

	opts = append(opts, uIntOpt, grpc.WithInsecure(), grpc.WithBalancer(b))

	conn, err := grpc.Dial(serviceName, opts...)

	return conn, err
}
