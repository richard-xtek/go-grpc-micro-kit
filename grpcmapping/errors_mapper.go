package grpcmapping

import (
	"context"
	"io"
	"net/http"

	"github.com/richard-xtek/go-grpc-micro-kit/grpc-gen/errors"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

type responseBody struct {
	Error *errors.Error `json:"error"`
}

// NewError return an error
func NewError(code int32, message string) error {
	return status.ErrorProto(&spb.Status{
		Code:    code,
		Message: message,
	})
}

// TransformErrors transform function errors to HTTP errors
func TransformErrors(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	const fallback = `{"error":{"code":-1,"message":"failed to marshal error message","domain":"zpi"}}`

	s, _ := status.FromError(err)

	w.Header().Del("Trailer")

	contentType := marshaler.ContentType()
	// Check marshaler on run time in order to keep backwards compatability
	// An interface param needs to be added to the ContentType() function on
	// the Marshal interface to be able to remove this check
	if httpBodyMarshaler, ok := marshaler.(*runtime.HTTPBodyMarshaler); ok {
		pb := s.Proto()
		contentType = httpBodyMarshaler.ContentTypeFromMessage(pb)
	}
	w.Header().Set("Content-Type", contentType)

	// Transform to API error structure
	body := responseFromGrpcCode(ctx, s.Code())
	// Log if get unknown error
	switch body.Error.GetCode() {
	case int32(errors.GeneralErrorCode_AMOUNT_INVALID):
		grpclog.Errorln(err)
	}

	buf, merr := marshaler.Marshal(body)
	if merr != nil {
		grpclog.Infof("Failed to marshal error message %q: %v", body, merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			grpclog.Infof("Failed to write response: %v", err)
		}
		return
	}

	// md, ok := ServerMetadataFromContext(ctx)
	// if !ok {
	// 	grpclog.Infof("Failed to extract ServerMetadata from context")
	// }

	// handleForwardResponseServerMetadata(w, mux, md)
	// handleForwardResponseTrailerHeader(w, md)
	if _, err := w.Write(buf); err != nil {
		grpclog.Infof("Failed to write response: %v", err)
	}

	// handleForwardResponseTrailer(w, md)
}

func responseFromGrpcCode(ctx context.Context, code codes.Code) responseBody {
	var err *errors.Error

	switch code {
	case codes.OK:
		err = errors.NewInternalError(ctx, errors.CodeSuccess)
	case codes.Unauthenticated:
		err = errors.NewInternalError(ctx, errors.CodeUnauthenticated)
	default:
		err = errors.NewInternalError(ctx, errors.CodeUnknown)
	}

	return responseBody{
		Error: err,
	}
}
