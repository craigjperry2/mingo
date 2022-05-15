package middleware

import (
	"net/http"

	"github.com/craigjperry2/mingo/internal/app/mingo/config"
	"github.com/craigjperry2/mingo/internal/app/mingo/logger"
)

// A middleware that can log accesses
func NewLoggingMiddleware() middleware {
	clock := config.GetInstance().GetClock()
	loggingDestination := config.GetInstance().GetLoggingDestination()
	hostname := config.GetInstance().GetHostname()

	accessLogger := logger.NewComponentLogger(clock, loggingDestination, hostname, "access")

	return func(hdlr http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := clock().UTC()
			lrw := NewLoggingResponseWriter(w)
			defer func() {
				requestID := w.Header().Get("X-Request-Id")
				if requestID == "" {
					requestID = "unknown"
				}
				var pathWithQuery string
				if req.URL.RawQuery != "" {
					pathWithQuery = req.URL.Path + "?" + req.URL.RawQuery
				} else {
					pathWithQuery = req.URL.Path
				}
				accessLogger.Println(requestID, req.Method, lrw.statusCode, pathWithQuery, req.RemoteAddr, req.UserAgent(), clock().UTC().Sub(start))

			}()
			hdlr.ServeHTTP(lrw, req)
		})
	}
}

// I want to log the HTTP Status code of each request
// with thanks to https://gist.github.com/Boerworz/b683e46ae0761056a636
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
