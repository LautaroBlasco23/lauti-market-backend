package infrastructure

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// responseRecorder wraps http.ResponseWriter to capture the status code
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rr *responseRecorder) WriteHeader(code int) {
	if !rr.written {
		rr.statusCode = code
		rr.written = true
		rr.ResponseWriter.WriteHeader(code)
	}
}

func (rr *responseRecorder) Write(b []byte) (int, error) {
	if !rr.written {
		rr.written = true
	}
	return rr.ResponseWriter.Write(b)
}

// loggingContextKey is a type for context keys to avoid collisions
type loggingContextKey string

const RequestIDKey loggingContextKey = "request_id"

// GetRequestID retrieves the request ID from the context
func GetRequestID(r *http.Request) string {
	if id, ok := r.Context().Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// LoggingMiddleware logs all incoming HTTP requests with method, path, status code, duration, and request ID
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Generate a unique request ID
		requestID := uuid.New().String()

		// Create a new context with the request ID
		ctx := r.Context()
		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		r = r.WithContext(ctx)

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Wrap the response writer to capture status code
		recorder := newResponseRecorder(w)

		// Log the incoming request
		slog.Info("incoming request",
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("remote_addr", r.RemoteAddr),
			slog.String("user_agent", r.UserAgent()),
		)

		// Process the request
		next.ServeHTTP(recorder, r)

		// Calculate duration
		duration := time.Since(start)

		// Determine log level based on status code
		statusCode := recorder.statusCode
		logAttrs := []slog.Attr{
			slog.String("request_id", requestID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status_code", statusCode),
			slog.Duration("duration", duration),
		}

		switch {
		case statusCode >= 500:
			slog.LogAttrs(r.Context(), slog.LevelError, "request completed with server error", logAttrs...)
		case statusCode >= 400:
			slog.LogAttrs(r.Context(), slog.LevelWarn, "request completed with client error", logAttrs...)
		default:
			slog.LogAttrs(r.Context(), slog.LevelInfo, "request completed", logAttrs...)
		}
	})
}
