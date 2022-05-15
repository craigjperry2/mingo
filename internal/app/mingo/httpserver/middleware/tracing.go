package middleware

import (
	"net/http"
	"strconv"

	"github.com/craigjperry2/mingo/internal/app/mingo/config"
)

// Facilitate tracing requests by assigning unique id's where they don't exist already
type IdFountain func() string

func NewIdFountain() IdFountain {
	clock := config.GetInstance().GetClock()
	return func() string {
		return strconv.FormatInt(clock().UTC().UnixNano(), 36)
	}
}

func NewTracingMiddleware(nextRequestID IdFountain) middleware {
	return func(hdlr http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestID := req.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			w.Header().Set("X-Request-Id", requestID)
			hdlr.ServeHTTP(w, req)
		})
	}
}
