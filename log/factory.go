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

package log

import (
	"context"
	"fmt"
	"os"

	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	tracingUtils "github.com/richard-xtek/go-grpc-micro-kit/tracing/utils"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Factory is the default logging wrapper that can create
// logger instances either for a given Context or context-less.
type Factory struct {
	level  *zap.AtomicLevel
	logger *zap.Logger
}

// NewFactory creates a new Factory.
func NewFactory(logger *zap.Logger) Factory {
	return Factory{logger: logger}
}

// NewDevelopFactory ...
func NewDevelopFactory(serviceName string) Factory {
	encoderCfg := zap.NewDevelopmentEncoderConfig()
	encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	atomicLevel := zap.NewAtomicLevelAt(zap.InfoLevel)

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		os.Stdout,
		atomicLevel,
	)
	zapLogger := zap.New(core, zap.AddCallerSkip(1), zap.AddCaller())
	zapLogger.With(zap.String("service", serviceName))
	return Factory{logger: zapLogger, level: &atomicLevel}
}

// NewStandardFactory creates a new Factory.
func NewStandardFactory(logFolder, serviceName string) Factory {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	atomicLevel := zap.NewAtomicLevelAt(zap.DebugLevel)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s.log", logFolder, serviceName),
			MaxSize:    10, // megabytes
			MaxBackups: 200,
			MaxAge:     90, // days
		}),
		atomicLevel,
	)
	zapLogger := zap.New(core, zap.AddCallerSkip(1), zap.AddCaller())
	zapLogger.With(zap.String("service", serviceName))
	return Factory{logger: zapLogger, level: &atomicLevel}
}

//NewLogFactory ...
func NewLogFactory(logFolder, serviceName string, logLevel zapcore.Level) Factory {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "ts"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	atomicLevel := zap.NewAtomicLevelAt(logLevel)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(&lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s.log", logFolder, serviceName),
			MaxSize:    10, // megabytes
			MaxBackups: 200,
			MaxAge:     90, // days
		}),
		atomicLevel,
	)
	zapLogger := zap.New(core, zap.AddCallerSkip(1), zap.AddCaller())
	zapLogger.With(zap.String("service", serviceName))
	return Factory{logger: zapLogger, level: &atomicLevel}
}

// IsNil return true if logger is nil, otherwise return false.
func (b Factory) IsNil() bool {
	if b.logger == nil {
		return true
	}
	return false
}

// Bg creates a context-unaware logger.
func (b Factory) Bg() Logger {
	return logger{logger: b.logger}
}

// For returns a context-aware Logger. If the context
// contains an OpenTracing span, all logging calls are also
// echo-ed into the span.
func (b Factory) For(ctx context.Context) Logger {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		values := tracingUtils.ExtractOpentracingIds(span)

		traceID, _ := values.Get(grpc_opentracing.TagTraceId).(string)
		isSampled, _ := values.Get(grpc_opentracing.TagSampled).(string)

		newFactory := b.With(zap.String(grpc_opentracing.TagTraceId, traceID))
		newFactory = newFactory.With(zap.String(grpc_opentracing.TagSampled, isSampled))

		return spanLogger{span: span, logger: newFactory.logger}
	}
	return b.Bg()
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (b Factory) With(fields ...zapcore.Field) Factory {
	return Factory{logger: b.logger.With(fields...)}
}

// EnableLevel ...
func (b Factory) EnableLevel(level zapcore.Level) {
	b.level.SetLevel(level)
}

// LogLevel ...
func (b Factory) LogLevel() zapcore.Level {
	return b.level.Level()
}
