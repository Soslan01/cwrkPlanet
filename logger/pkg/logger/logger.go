package logger

import "log/slog"

var def *slog.Logger

// Init настраивает slog в зависимости от среды
func Init(cfg Config) {
	if cfg.Env == "" {
		cfg.Env = DetectEnv()
	}
	if cfg.Service == "" {
		cfg.Service = "app"
	}
	cfg.InstanceID = ensureInstanceID(cfg.InstanceID)

	// Выбор бекенда по умолчанию
	if cfg.Backend == "" {
		if cfg.Env == EnvDev {
			cfg.Backend = BackendStd
		} else {
			cfg.Backend = BackendZap
		}
	}

	var h slog.Handler
	switch cfg.Backend {
	case BackendZap:
		h = newZapHandler(cfg)
	default:
		h = newStdHandler(cfg)
	}

	h = h.WithAttrs(commonAttr(cfg))

	base := slog.New(h)
	slog.SetDefault(base)
	def = base
}

func L() *slog.Logger {
	if def != nil {
		return def
	}

	Init(Config{})
	return def
}
