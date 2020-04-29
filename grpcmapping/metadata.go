package grpcmapping

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"
)

// AppendRequestMetadata append cookies and headers to incoming context
func AppendRequestMetadata(ctx context.Context, req *http.Request) metadata.MD {
	md := metadata.MD{}

	// Append cookies
	cookies := req.Cookies()
	for _, cookie := range cookies {
		md.Append(cookie.Name, cookie.Value)
	}

	// Append ip
	clientIP := GetClientIP(req)
	md.Append("x-client-ip", clientIP)

	return md
}
