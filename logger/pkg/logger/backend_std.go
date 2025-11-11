package logger

import (
	"log/slog"
	"os"
)

func newStdHandler(cfg Config) slog.Handler {
	var level slog.Level
	// if level == 0 {
	// 	level = defaultLevelForEnv(cfg.Env)
	// }

	if cfg.Debug && cfg.Level == 0 {
		level = slog.LevelDebug
	} else {
		level = cfg.Level
	}

	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource || false,
	})
}
