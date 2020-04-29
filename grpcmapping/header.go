package grpcmapping

import "net/http"

// GetClientIP get client IP from HTTP request
func GetClientIP(req *http.Request) string {
	clientIP := req.Header.Get("HTTP_X_FORWARDED_FOR")
	if len(clientIP) == 0 {
		clientIP = req.Header.Get("X-Forwarded-For")
	}
	if len(clientIP) == 0 {
		clientIP = req.Header.Get("REMOTE_ADDR")
	}
	return clientIP
}
