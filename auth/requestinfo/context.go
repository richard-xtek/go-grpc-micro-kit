package requestinfo

import "context"

var (
	ctxMarkerKey = "auth_request_info"
)

type keyRequestInfo struct{}

// ExtractRequestInfo ...
func ExtractRequestInfo(ctx context.Context) (info *RequestInfo, ok bool) {
	info, ok = ctx.Value(keyRequestInfo{}).(*RequestInfo)
	return
}

// NewContext ...
func NewContext(ctx context.Context, info *RequestInfo) context.Context {
	return context.WithValue(ctx, keyRequestInfo{}, info)
}
