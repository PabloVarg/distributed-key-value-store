package api

import (
	"log/slog"
	"net/http"
)

func hitLoggingMiddleware(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			l.Info("http hit", "path", r.URL.String(), "method", r.Method)
			next.ServeHTTP(w, r)
			l.Info("http response", "path", r.URL.String(), "method", r.Method)
		})
	}
}
