package middleware

import (
	"log"
	"net/http"
	"tech-ip-sem2/shared/models"
	"time"
)

type ResponseWriterWrapper struct {
	http.ResponseWriter
	StatusCode int
}

func (rw *ResponseWriterWrapper) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := &ResponseWriterWrapper{ResponseWriter: w, StatusCode: http.StatusOK}

		next.ServeHTTP(wrappedWriter, r)

		requestID := r.Context().Value(models.RequestIDKey{})
		if requestID == nil {
			requestID = "unknown"
		}

		log.Printf("[%s] %s %s %d %v",
			requestID,
			r.Method,
			r.URL.Path,
			wrappedWriter.StatusCode,
			time.Since(start))
	})
}
