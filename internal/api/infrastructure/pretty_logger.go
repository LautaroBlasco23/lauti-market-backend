package infrastructure

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"
)

// ANSI color codes
const (
	colorReset   = "\033[0m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorRed     = "\033[31m"
	colorBlue    = "\033[34m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
	colorMagenta = "\033[35m"
	colorBold    = "\033[1m"
)

// Icons for different log levels
const (
	iconSuccess = "●" // Green circle
	iconWarning = "●" // Yellow circle
	iconError   = "●" // Red circle
	iconInfo    = "→" // Arrow for incoming
)

type PrettyHandler struct {
	slog.Handler
	writer io.Writer
	mu     sync.Mutex
}

func NewPrettyHandler(w io.Writer, opts *slog.HandlerOptions) *PrettyHandler {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}
	return &PrettyHandler{
		Handler: slog.NewTextHandler(w, &slog.HandlerOptions{
			Level: opts.Level,
		}),
		writer: w,
	}
}

func (h *PrettyHandler) Handle(ctx context.Context, r slog.Record) error { //nolint:gocritic // Required by slog.Handler interface
	h.mu.Lock()
	defer h.mu.Unlock()

	// Format timestamp
	timestamp := r.Time.Format("15:04:05")

	// Determine style based on message content
	var prefix, color string
	msg := r.Message

	switch {
	case strings.Contains(msg, "server error"):
		color = colorRed
		prefix = iconError
	case strings.Contains(msg, "client error"):
		color = colorYellow
		prefix = iconWarning
	case strings.Contains(msg, "incoming request"):
		color = colorCyan
		prefix = iconInfo
	case strings.Contains(msg, "request completed"):
		// Check status code from attributes
		statusCode := getStatusCodeFromRecord(r)
		switch {
		case statusCode >= 500:
			color = colorRed
			prefix = iconError
		case statusCode >= 400:
			color = colorYellow
			prefix = iconWarning
		default:
			color = colorGreen
			prefix = iconSuccess
		}
	default:
		color = colorBlue
		prefix = "•"
	}

	// Build the formatted output
	var sb strings.Builder

	// Timestamp in gray
	sb.WriteString(colorGray)
	sb.WriteString(timestamp)
	sb.WriteString(colorReset)
	sb.WriteString(" ")

	// Icon with color
	sb.WriteString(color)
	sb.WriteString(prefix)
	sb.WriteString(colorReset)
	sb.WriteString(" ")

	// Message
	sb.WriteString(colorBold)
	sb.WriteString(color)
	sb.WriteString(msg)
	sb.WriteString(colorReset)

	// Attributes
	r.Attrs(func(a slog.Attr) bool {
		sb.WriteString(" ")
		sb.WriteString(formatAttr(a, color))
		return true
	})

	sb.WriteString("\n")

	_, err := h.writer.Write([]byte(sb.String()))
	return err
}

func getStatusCodeFromRecord(r slog.Record) int { //nolint:gocritic // slog.Record passed by value
	var statusCode int
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "status_code" {
			if v, ok := a.Value.Any().(int64); ok {
				statusCode = int(v)
			}
		}
		return true
	})
	return statusCode
}

func formatAttr(a slog.Attr, highlightColor string) string {
	key := a.Key
	value := a.Value.Any()

	switch v := value.(type) {
	case string:
		if key == "request_id" {
			// Shorten request ID for display
			if len(v) > 8 {
				v = v[:8]
			}
			return fmt.Sprintf("%s%s%s=%s%s%s",
				colorGray, key, colorReset,
				colorMagenta, v, colorReset)
		}
		return fmt.Sprintf("%s%s%s=%s%s%s",
			colorGray, key, colorReset,
			colorCyan, v, colorReset)
	case int:
		if key == "status_code" {
			var statusColor string
			switch {
			case v >= 500:
				statusColor = colorRed
			case v >= 400:
				statusColor = colorYellow
			default:
				statusColor = colorGreen
			}
			return fmt.Sprintf("%s%s%s=%s%d%s",
				colorGray, key, colorReset,
				statusColor, v, colorReset)
		}
		return fmt.Sprintf("%s%s%s=%s%d%s",
			colorGray, key, colorReset,
			colorYellow, v, colorReset)
	case time.Duration:
		return fmt.Sprintf("%s%s%s=%s%s%s",
			colorGray, key, colorReset,
			colorYellow, v, colorReset)
	default:
		return fmt.Sprintf("%s%s%s=%s%v%s",
			colorGray, key, colorReset,
			colorCyan, v, colorReset)
	}
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler.WithAttrs(attrs),
		writer:  h.writer,
	}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	return &PrettyHandler{
		Handler: h.Handler.WithGroup(name),
		writer:  h.writer,
	}
}

// InitLogger initializes the pretty logger
func InitLogger() {
	handler := NewPrettyHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
