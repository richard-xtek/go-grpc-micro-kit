package grpcmapping

import "net/http"

// HandleCrossOrigin serve OPTIONS method for CORS policy
func HandleCrossOrigin(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
		} else {
			handler.ServeHTTP(w, r)
		}
	})
}
