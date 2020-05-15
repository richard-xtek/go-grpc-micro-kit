package requestinfo

import (
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	// AuthHeaderKey ...
	AuthHeaderKey = "authorization"
	// ClientIPKey ...
	ClientIPKey = "grpcgateway-client-ip"
	// UserAgentKey ...
	UserAgentKey = "grpcgateway-user-agent"
	// UserIDRequestKey ...
	UserIDRequestKey = "grpcgateway-user-id-request"
)

func getHeaderString(md metadata.MD, key string, defaultValue ...string) string {
	values := md[key]
	value := ""
	if len(values) > 0 {
		value = values[0]
	}
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}
	return value
}

func getClientUserAgent(md metadata.MD) string {
	return getHeaderString(md, UserAgentKey, "")
}

func getClientIP(md metadata.MD) string {
	return getHeaderString(md, ClientIPKey, "127.0.0.1")
}

func getUserIDRequest(md metadata.MD) string {
	return getHeaderString(md, UserIDRequestKey)
}

func getSessionID(md metadata.MD) string {
	// Check header Authorization
	auth := getHeaderString(md, AuthHeaderKey)
	if auth != "" {
		tokens := strings.Split(auth, " ")
		if len(tokens) > 1 {
			return tokens[1]
		}
	}

	return ""
}
