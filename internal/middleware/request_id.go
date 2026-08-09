package middleware

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

var requestCounter uint64

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")

		if requestID == "" {
			id := atomic.AddUint64(&requestCounter, 1)
			requestID = fmt.Sprintf("req-%d", id)
		}

		w.Header().Set("X-Request-ID", requestID)

		ctx := r.Context()
		ctx = contextWithRequestID(ctx, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
