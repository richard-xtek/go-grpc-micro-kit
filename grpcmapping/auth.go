package grpcmapping

import (
	"context"

	"google.golang.org/grpc"
)

type authKey string

const (
	// AuthSession ...
	AuthSession authKey = "session"

	// AuthHeader ...
	AuthHeader string = "authorization"

	// OauthCookie ...
	OauthCookie string = "zalo_oauth"

	// AuthCookie ...
	AuthCookie authKey = "h5token"

	// DeviceIDCookie ...
	DeviceIDCookie authKey = "h5did"

	// TrackingSessionCookie ...
	TrackingSessionCookie authKey = "trackingsession"

	// TSessionCookie ...
	TSessionCookie authKey = "tsession"
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
