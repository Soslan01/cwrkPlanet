package httpmw

import (
	"context"
	"net/http"
	"strconv"
	"strings"
)

type ctxKey string

const (
	ctxKeyToken  ctxKey = "token"
	ctxKeyUserID ctxKey = "user_id"
)

// простая авторизация: требуем Bearer + X-User-ID (UUID), без валидации токена
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") || len(auth) <= 7 {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"missing bearer token"}`))
			return
		}

		uidHeader := r.Header.Get("X-User-ID")
		if uidHeader == "" {
			http.Error(w, `{"error":"missing X-User-ID"}`, http.StatusUnauthorized)
			return
		}

		uid, err := strconv.ParseInt(uidHeader, 10, 64)
		if err != nil {
			http.Error(w, `{"error":"invalid X-User-ID (must be int64)"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), ctxKeyToken, strings.TrimSpace(auth[7:]))
		ctx = context.WithValue(ctx, ctxKeyUserID, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserIDFromCtx(ctx context.Context) int64 {
	if v := ctx.Value(ctxKeyUserID); v != nil {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}
