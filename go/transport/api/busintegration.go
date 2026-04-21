package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// WithEventPublisher enables event publishing on the router. It publishes
// "api.request.start" and "api.request.end" events for every request.
func WithEventPublisher(p EventPublisher) RouterOption {
	return func(r *Router) {
		r.middleware = append(r.middleware, eventMiddleware(p))
	}
}

func eventMiddleware(p EventPublisher) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			_ = p.Publish(r.Context(), "api.request.start", "api.router",
				map[string]string{
					"method": r.Method,
					"path":   r.URL.Path,
				},
			)

			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(sw, r)

			_ = p.Publish(context.Background(), "api.request.end", "api.router",
				map[string]any{
					"method":   r.Method,
					"path":     r.URL.Path,
					"status":   sw.status,
					"duration": fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
				},
			)
		})
	}
}
