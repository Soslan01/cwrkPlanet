package logger

import "log/slog"

type Backend string

const (
	BackendStd Backend = "std" // Text в dev; JSON в stage/prod
	BackendZap Backend = "zap" // Slog-zap
)

type Config struct {
	// Метаданные для логгера
	Service    string
	Version    string
	InstanceID string

	// Управление выводом
	Level   slog.Level
	Env     Env
	Backend Backend // default: zap для stage/prod, std для dev
	Debug   bool

	// Zap sampling
	SampleInitial    int
	SampleThereafter int
	SampleTick       int

	// AddSource в dev
	AddSource bool
}
