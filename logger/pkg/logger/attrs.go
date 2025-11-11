package logger

import (
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
)

func ensureInstanceID(v string) string {
	if v != "" {
		return v
	}

	hn, _ := os.Hostname()
	uid := uuid.New().String()[:8]
	return hn + "-" + uid
}

func commonAttr(cfg Config) []slog.Attr {
	return []slog.Attr{
		slog.String("service", cfg.Service),
		slog.String("env", string(cfg.Env)),
		slog.String("version", cfg.Version),
		slog.String("instance_id", cfg.InstanceID),
		slog.Time("started_at", time.Now()),
	}
}
