// Copyright (c) 2017 Uber Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracing

import (
	"context"
	"fmt"
	"strings"
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/richard-xtek/go-grpc-micro-kit/log"
	jaeger "github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/rpcmetrics"
	"github.com/uber/jaeger-client-go/transport"
	"github.com/uber/jaeger-lib/metrics"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
)

// Init creates a new instance of Jaeger tracer.
func Init(serviceName string, logger log.Factory, backendHostPort string, options ...config.Option) opentracing.Tracer {
	if backendHostPort == "" {
		backendHostPort = "localhost:6831"
	}

	metricsFactory := prometheus.New().Namespace(metrics.NSOptions{Name: serviceName})

	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type: "remote",
		},
	}

	jaegerLogger := jaegerLoggerAdapter{logger.Bg()}
	var sender jaeger.Transport
	if strings.HasPrefix(backendHostPort, "http://") {
		sender = transport.NewHTTPTransport(
			backendHostPort,
			transport.HTTPBatchSize(1),
		)
	} else {
		if s, err := jaeger.NewUDPTransport(backendHostPort, 0); err != nil {
			logger.Bg().Fatal("cannot initialize UDP sender", zap.Error(err))
		} else {
			sender = s
		}
	}

	// at least 10 traces per second, up to 1% of total traces
	sampler, _ := jaeger.NewGuaranteedThroughputProbabilisticSampler(10, 0.01)

	// default options
	applyOptions := []config.Option{
		config.Reporter(jaeger.NewRemoteReporter(
			sender,
			jaeger.ReporterOptions.BufferFlushInterval(1*time.Second),
			jaeger.ReporterOptions.Logger(jaegerLogger),
		)),
		config.Sampler(sampler),
		config.Logger(jaegerLogger),
		config.Metrics(metricsFactory),
		config.Observer(rpcmetrics.NewObserver(metricsFactory, rpcmetrics.DefaultNameNormalizer)),
	}

	// override option setting
	if len(options) > 0 {
		applyOptions = append(applyOptions, options...)
	}

	tracer, _, err := cfg.New(
		serviceName,
		applyOptions...,
	)
	if err != nil {
		logger.Bg().Fatal("cannot initialize Jaeger Tracer", zap.Error(err))
	}
	return tracer
}

type jaegerLoggerAdapter struct {
	logger log.Logger
}

func (l jaegerLoggerAdapter) Error(msg string) {
	l.logger.Error(msg)
}

func (l jaegerLoggerAdapter) Infof(msg string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(msg, args...))
}

// SetBaggageItem sets a key:value pair on this Span and its SpanContext
// that also propagates to descendants of this Span.
func SetBaggageItem(ctx context.Context, key string, value string) {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		span.SetBaggageItem(key, value)
	}
}

// WriteLogGRPCMessage if isEnable is true: write to log with request/reponse message of grpc. Otherwise no write log
func WriteLogGRPCMessage(ctx context.Context, isEnable bool) {
	if isEnable {
		SetBaggageItem(ctx, "output_grpc_message", "true")
	} else {
		SetBaggageItem(ctx, "output_grpc_message", "")
	}
}
