package server

import (
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	grpc_logf "github.com/richard-xtek/go-grpc-micro-kit/grpc-logf"
	"go.uber.org/zap"

	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"github.com/richard-xtek/go-grpc-micro-kit/registry"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	opentracing "github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// GRPCRegister ...
type GRPCRegister func(server *grpc.Server)

// NewGRPCServer ...
func NewGRPCServer(logger log.Factory, name string) *GRPCServer {
	s := &GRPCServer{logger: logger}
	s.name = name

	s.quite = make(chan bool)

	return s
}

// GRPCServer ...
type GRPCServer struct {
	name string
	host string
	port string

	logger log.Factory
	server *grpc.Server

	didStarted bool
	quite      chan bool

	grpcRegister GRPCRegister

	grpcUnaryInterceptors  []grpc.UnaryServerInterceptor
	grpcStreamInterceptors []grpc.StreamServerInterceptor

	tracer opentracing.Tracer
	consul *registry.ConsulRegister
}

// WithUnaryServerInterceptor ...
func (s *GRPCServer) WithUnaryServerInterceptor(usi grpc.UnaryServerInterceptor) *GRPCServer {
	s.grpcUnaryInterceptors = append(s.grpcUnaryInterceptors, usi)
	return s
}

// WithPort ...
func (s *GRPCServer) WithPort(port string) *GRPCServer {
	s.port = port
	return s
}

// WithHost ...
func (s *GRPCServer) WithHost(host string) *GRPCServer {
	s.host = host
	return s
}

// WithTracer ...
func (s *GRPCServer) WithTracer(tracer opentracing.Tracer) *GRPCServer {
	s.tracer = tracer
	return s
}

// WithConsul ...
func (s *GRPCServer) WithConsul(consul *registry.ConsulRegister) *GRPCServer {
	s.consul = consul
	return s
}

// WithHandler ...
func (s *GRPCServer) WithHandler(handler GRPCRegister) *GRPCServer {
	s.grpcRegister = handler
	return s
}

func (s *GRPCServer) makeServer() {
	alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }

	sIntOpt := grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_opentracing.StreamServerInterceptor(grpc_opentracing.WithTracer(s.tracer)),
		grpc_prometheus.StreamServerInterceptor,
		grpc_logf.StreamServerInterceptor(s.logger),
		grpc_logf.PayloadStreamServerInterceptor(s.logger, alwaysLoggingDeciderServer),
		grpc_recovery.StreamServerInterceptor(),
		grpc_middleware.ChainStreamServer(s.grpcStreamInterceptors...),
	))

	uIntOpt := grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
			grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(s.tracer), grpc_opentracing.WithFilterFunc(func(ctx context.Context, fullMethodName string) bool {
				if fullMethodName == "/grpc.health.v1.Health/Check" {
					return false
				}
				return true
			})),
			grpc_prometheus.UnaryServerInterceptor,
			grpc_logf.UnaryServerInterceptor(s.logger),
			grpc_logf.PayloadUnaryServerInterceptor(s.logger, alwaysLoggingDeciderServer),
			grpc_recovery.UnaryServerInterceptor(),
			grpc_middleware.ChainUnaryServer(s.grpcUnaryInterceptors...),
		),
	)
	s.server = grpc.NewServer(
		sIntOpt,
		uIntOpt,
	)
}

// EnablePrometheus ...
func (s *GRPCServer) EnablePrometheus(port string, isEnableHistogram bool) {
	if isEnableHistogram {
		grpc_prometheus.EnableClientHandlingTimeHistogram()
		grpc_prometheus.EnableHandlingTimeHistogram()
	}

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/debug/vars", expvar.Handler())
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(fmt.Sprintf(":%s", port), mux); err != nil {
			s.logger.Bg().Error("ListenAndServe Prometheus", zap.Error(err))
		}
	}()
	s.logger.Bg().Info("Prometheus enable on port " + port)
}

// GetAddressListen ...
func (s *GRPCServer) GetAddressListen() string {
	return fmt.Sprintf("%s:%s", s.host, s.port)
}

// GetServer ...
func (s *GRPCServer) GetServer() *grpc.Server {
	return s.server
}

// Start ...
func (s *GRPCServer) Start() error {
	return s.start()
}

func (s *GRPCServer) start() error {
	// create grpc server
	s.makeServer()

	ln, err := net.Listen("tcp", s.GetAddressListen())
	if err != nil {
		return err
	}

	s.logger.Bg().Info("Starting " + s.name + " GRPC Server on port " + s.port)
	if err := s.serveWithListener(ln); err != nil {
		return err
	}
	s.logger.Bg().Info("Started " + s.name + " GRPC Server on port " + s.port)

	return nil
}

func (s *GRPCServer) serveWithListener(l net.Listener) error {
	go func() {
		if s.consul != nil {
			id, err := s.consul.Register()
			if err != nil {
				// s.logger.WithError(err).Fatal("Consul register")
			}
			defer func() {
				if err := s.consul.Deregister(id); err != nil {
					// s.logger.WithError(err).Error("Deregister consul")
				}
			}()
		}

		s.grpcRegister(s.server)
		reflection.Register(s.server)
		grpc_prometheus.Register(s.server)

		err := s.server.Serve(l)
		if err != nil {
			// s.logger.WithError(err).Error()
		}
	}()
	return nil
}

// Stop ...
func (s *GRPCServer) Stop() error {
	return nil
}
