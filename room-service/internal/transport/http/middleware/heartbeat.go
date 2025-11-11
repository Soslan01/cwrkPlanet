package httpmw

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// HeartbeatMiddleware обновляет last_seen для {roomID,userID} если roomID есть в пути.
type HeartbeatToucher interface {
	TouchHeartbeat(ctx context.Context, roomID string, userID int64) error
}

func HeartbeatMiddleware(memberSvc HeartbeatToucher) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := UserIDFromCtx(r.Context())
			if userID != 0 {
				if roomID := chi.URLParam(r, "id"); roomID != "" {
					// best-effort: ошибки не прерывают запрос
					_ = memberSvc.TouchHeartbeat(r.Context(), roomID, userID)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
