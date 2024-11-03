// the logger middleware provides elaborate HTTP command logging with duration times
package logger

import (
	"net/http"
	"time"
)

// ResponseWriter object with additional fields (i.e. HTTP status code)
type ResponseWriter struct {
	http.ResponseWriter     // next ResponseWriter in middleware chain
	status              int // HTTP status code
}

// ResponseWriter fabric
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

// embed additional ResponseWriter fields
func (rw *ResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// main logger middleware function
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer zapLog.Sync()

		start := time.Now()

		// Create a response writer that captures the response status code
		rw := NewResponseWriter(w)

		// Pass the request to the next handler
		next.ServeHTTP(rw, r)

		//count execution time
		duration := time.Since(start)

		// Log the incoming request
		CommandTrace(r.Method, r.URL.Path, rw.status, duration)
	})
}
