package utils

import (
	"github.com/afex/hystrix-go/hystrix"
	"github.com/richard-xtek/go-grpc-micro-kit/monitor/hystrixconfig"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// HystrixEnableFlag is a flag for enable/disable hystrix
var HystrixEnableFlag = true

// UnaryClientInterceptor ...
func UnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if HystrixEnableFlag {
			hystrix.ConfigureCommand(method, hystrixconfig.HystrixConfig())

			err := hystrix.Do(method, func() (err error) {
				err = invoker(ctx, method, req, reply, cc, opts...)
				return err
			}, nil)

			return err
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor ...
func StreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return streamer(ctx, desc, cc, method, opts...)
	}
}
