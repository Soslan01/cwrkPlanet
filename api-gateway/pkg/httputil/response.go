package httputil

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type envelope map[string]any

func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write json response failed", slog.Any("err", err))
	}
}

// OK — «успешный» ответ с обёрткой.
func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, envelope{"data": data})
}

// Error — унифицированная ошибка (message + code).
func Error(ctx context.Context, w http.ResponseWriter, status int, msg string, meta map[string]any) {
	payload := envelope{
		"error": envelope{
			"message": msg,
		},
	}
	if len(meta) > 0 {
		payload["error"].(envelope)["meta"] = meta
	}
	JSON(w, status, payload)
}
