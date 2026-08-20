package main

import (
	"context"
	"io"
	"log/slog"
	"strings"

	"sing_panel/services"
)

// Gin does not expose a runtime log level. Filter its normal access output
// using the same LevelVar as slog; recovery/error output remains visible at
// error level and above.
type levelFilteredWriter struct {
	writer io.Writer
	level  slog.Level
}

func (w *levelFilteredWriter) Write(p []byte) (int, error) {
	if services.PanelLogLevel.Level() > w.level {
		return len(p), nil
	}
	return w.writer.Write(p)
}

type memoryLogWriter struct {
	writer io.Writer
	level  slog.Level
	source string
}

func (w *memoryLogWriter) Write(p []byte) (int, error) {
	services.GetMemoryLog().Append(memoryWriterLevelName(w.level), w.source, strings.TrimSpace(string(p)))
	if services.PanelLogLevel.Level() > w.level {
		return len(p), nil
	}
	return w.writer.Write(p)
}
func memoryWriterLevelName(level slog.Level) string {
	switch {
	case level <= slog.LevelDebug:
		return "debug"
	case level <= slog.LevelInfo:
		return "info"
	case level <= slog.LevelWarn:
		return "warn"
	default:
		return "error"
	}
}

type fanoutHandler struct{ handlers []slog.Handler }

func (h *fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, record.Level) {
			if err := handler.Handle(ctx, record); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	result := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		result[i] = handler.WithAttrs(attrs)
	}
	return &fanoutHandler{handlers: result}
}

func (h *fanoutHandler) WithGroup(name string) slog.Handler {
	result := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		result[i] = handler.WithGroup(name)
	}
	return &fanoutHandler{handlers: result}
}
