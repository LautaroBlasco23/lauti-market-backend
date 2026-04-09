package infrastructure

import (
	"log/slog"
	"net/http"
	"strings"
)

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Debug logging
		slog.Debug("CORS request", "origin", origin, "method", r.Method, "url", r.URL.Path)

		// For development: allow any origin that contains common dev host patterns
		isAllowed := false

		// Always allow these origins
		allowedPatterns := []string{
			"localhost",
			"127.0.0.1",
			".loca.lt",    // localtunnel
			".ngrok-free", // ngrok
			".ngrok.io",   // ngrok classic
		}

		for _, pattern := range allowedPatterns {
			if strings.Contains(origin, pattern) {
				isAllowed = true
				break
			}
		}

		if isAllowed || origin == "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
