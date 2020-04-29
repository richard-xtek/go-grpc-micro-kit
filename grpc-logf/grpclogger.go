// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package grpc_logf

import (
	"fmt"

	"github.com/richard-xtek/go-grpc-micro-kit/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/grpclog"
)

// ReplaceGrpcLogger sets the given zap.Logger as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
func ReplaceGrpcLogger(logger log.Factory) {
	zgl := &zapGrpcLogger{logger.With(SystemField, zap.Bool("grpc_log", true))}
	grpclog.SetLogger(zgl)
}

type zapGrpcLogger struct {
	logger log.Factory
}

func (l *zapGrpcLogger) Fatal(args ...interface{}) {
	l.logger.Bg().Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Bg().Fatal(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLogger) Fatalln(args ...interface{}) {
	l.logger.Bg().Fatal(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Print(args ...interface{}) {
	l.logger.Bg().Info(fmt.Sprint(args...))
}

func (l *zapGrpcLogger) Printf(format string, args ...interface{}) {
	l.logger.Bg().Info(fmt.Sprintf(format, args...))
}

func (l *zapGrpcLogger) Println(args ...interface{}) {
	l.logger.Bg().Info(fmt.Sprint(args...))
}
