package logger

import (
	"log/slog"
	"os"
	"time"

	slogzap "github.com/samber/slog-zap/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newZapHandler(cfg Config) slog.Handler {
	var lvl slog.Level
	if cfg.Debug && cfg.Level == 0 {
		lvl = slog.LevelDebug
	} else {
		lvl = cfg.Level
	}

	encCfg := zap.NewProductionEncoderConfig()

	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	if cfg.AddSource {
		encCfg.EncodeCaller = zapcore.ShortCallerEncoder
	}
	enc := zapcore.NewJSONEncoder(encCfg)

	ws := zapcore.AddSync(os.Stdout)
	core := zapcore.NewCore(enc, ws, toZapLevel(lvl))

	// Sampling при всплесках логов
	initial := cfg.SampleInitial
	if initial <= 0 {
		initial = 100
	}
	thereafter := cfg.SampleThereafter
	if thereafter <= 0 {
		thereafter = 10
	}

	core = zapcore.NewSamplerWithOptions(core, time.Second, initial, thereafter)

	z := zap.New(
		core,
		zap.AddCaller(),
		zap.AddCallerSkip(1), // чтобы источник указывал на место вызова slog, а не обертку
	)

	return slogzap.Option{Logger: z}.NewZapHandler()
}

func toZapLevel(lvl slog.Level) zapcore.Level {
	switch {
	case lvl <= slog.LevelDebug:
		return zapcore.DebugLevel
	case lvl == slog.LevelInfo:
		return zapcore.InfoLevel
	case lvl == slog.LevelWarn:
		return zapcore.WarnLevel
	default:
		return zapcore.ErrorLevel
	}
}
