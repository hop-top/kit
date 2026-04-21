package api

import (
	"fmt"
	"net/http"
	"time"
)

// Logger returns a middleware that logs each request using the provided
// log function. It records method, path, status code, and duration.
func Logger(log func(msg string, args ...any)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)
			log("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration", fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
			)
		})
	}
}

// statusWriter wraps http.ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (w *statusWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.status = code
		w.wroteHeader = true
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.wroteHeader = true
	}
	return w.ResponseWriter.Write(b)
}
