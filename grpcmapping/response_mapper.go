package grpcmapping

import (
	"context"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func syncCookie(sd runtime.ServerMetadata, w http.ResponseWriter, key authKey) {
	md := sd.HeaderMD
	k := string(key)
	v := md[k]
	if len(v) > 0 && v[0] != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     k,
			HttpOnly: true,
			Value:    v[0],
			Domain:   ".zalopay.vn",
			Path:     "/",
		})
	}
}

// FormatHTTPResponse format http response from proto messages
func FormatHTTPResponse(ctx context.Context, w http.ResponseWriter, resp proto.Message) error {
	// fmt.Println("format", ctx, w, resp)

	md, _ := runtime.ServerMetadataFromContext(ctx)
	syncCookie(md, w, DeviceIDCookie)
	syncCookie(md, w, AuthCookie)
	syncCookie(md, w, TrackingSessionCookie)
	syncCookie(md, w, TSessionCookie)
	return nil
}
