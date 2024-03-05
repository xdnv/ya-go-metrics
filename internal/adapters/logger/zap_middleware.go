package logger

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ResponseWriter struct {
	http.ResponseWriter
	status int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{w, http.StatusOK}
}

func (rw *ResponseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

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
		zapLog.Info("Incoming request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.status),
			zap.Duration("duration", duration),
		)

	})
}
