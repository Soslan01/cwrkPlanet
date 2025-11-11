package httputil

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey string

const (
	HeaderRequestID        = "X-Request-ID"
	ctxKeyReqID     ctxKey = "req_id"
)

// MiddlewareRequestID — пробрасывает/генерирует X-Request-ID.
func MiddlewareRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(HeaderRequestID)
		if reqID == "" {
			reqID = uuid.NewString()
		}
		w.Header().Set(HeaderRequestID, reqID)

		ctx := context.WithValue(r.Context(), ctxKeyReqID, reqID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FromContext — достать request id из контекста.
func FromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyReqID).(string)
	return v, ok
}
