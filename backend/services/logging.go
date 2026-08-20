package services

import (
	"fmt"
	"log/slog"
	"strings"

	boxlog "github.com/sagernet/sing-box/log"
)

const DefaultLogLevel = "warn"

// PanelLogLevel is shared by slog and the Gin output filters. It is initialized
// before the database is opened so startup logs use the same default level.
var PanelLogLevel slog.LevelVar

func init() {
	PanelLogLevel.Set(slog.LevelWarn)
}

// NormalizeLogLevel validates the levels exposed by the settings UI. The same
// names are understood by sing-box, while slog maps them to its levels.
func NormalizeLogLevel(value string) (string, error) {
	level := strings.ToLower(strings.TrimSpace(value))
	switch level {
	case "debug", "info", "warn", "error":
		return level, nil
	default:
		return "", fmt.Errorf("unsupported log level %q (use debug, info, warn, or error)", value)
	}
}

func ApplyLogLevel(value string) error {
	level, err := NormalizeLogLevel(value)
	if err != nil {
		return err
	}
	switch level {
	case "debug":
		PanelLogLevel.Set(slog.LevelDebug)
	case "info":
		PanelLogLevel.Set(slog.LevelInfo)
	case "warn":
		PanelLogLevel.Set(slog.LevelWarn)
	case "error":
		PanelLogLevel.Set(slog.LevelError)
	}
	return nil
}

func CurrentLogLevel() string {
	switch PanelLogLevel.Level() {
	case slog.LevelDebug:
		return "debug"
	case slog.LevelInfo:
		return "info"
	case slog.LevelError:
		return "error"
	default:
		return "warn"
	}
}

func CurrentSingBoxLogLevel() boxlog.Level {
	switch CurrentLogLevel() {
	case "debug":
		return boxlog.LevelDebug
	case "info":
		return boxlog.LevelInfo
	case "error":
		return boxlog.LevelError
	default:
		return boxlog.LevelWarn
	}
}
