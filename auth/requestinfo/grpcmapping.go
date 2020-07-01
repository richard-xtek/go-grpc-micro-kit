package requestinfo

import (
	"context"

	"github.com/richard-xtek/go-grpc-micro-kit/log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthFunc interceptor for authentication
type AuthFunc func(ctx context.Context, method string) (context.Context, error)

// UnaryServerAuth add auth interceptor function
func UnaryServerAuth(authFunc AuthFunc) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if authFunc == nil {
			return handler(ctx, req)
		}

		newCtx, err := authFunc(ctx, info.FullMethod)
		if err != nil {
			return nil, err
		}

		return handler(newCtx, req)
	}
}

// Authentication ...
func Authentication(log log.Factory) AuthFunc {
	return func(ctx context.Context, fullMethod string) (context.Context, error) {
		md, _ := metadata.FromIncomingContext(ctx)
		info := RequestInfo{
			ClientIP:              getClientIP(md),
			UserAgent:             getClientUserAgent(md),
			SessionID:             getSessionID(md),
			UserID:                getUserIDRequest(md),
			SessionIDForAuthorize: getSessionIDAuthorize(md),
		}
		newCtx := NewContext(ctx, &info)
		return newCtx, nil
	}
}
