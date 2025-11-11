package http

import (
	"encoding/json"
	"net/http"
	"strings"

	appauth "github.com/cwrk-planet/api-gateway/internal/app/auth"
	"github.com/cwrk-planet/api-gateway/pkg/errs"
	"github.com/cwrk-planet/api-gateway/pkg/httputil"
)

type AuthHandlers struct {
	Auth appauth.Client
}

func (h *AuthHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var in appauth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid JSON", nil)
		return
	}
	if strings.TrimSpace(in.Email) == "" || strings.TrimSpace(in.Password) == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "email and password are required", nil)
		return
	}
	out, err := h.Auth.Login(r.Context(), in)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "login failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, out)
}

func (h *AuthHandlers) Register(w http.ResponseWriter, r *http.Request) {
	var in appauth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid JSON", nil)
		return
	}
	if strings.TrimSpace(in.Email) == "" || strings.TrimSpace(in.Password) == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "email and password are required", nil)
		return
	}
	out, err := h.Auth.Register(r.Context(), in)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "register failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, out)
}

func (h *AuthHandlers) Refresh(w http.ResponseWriter, r *http.Request) {
	var in appauth.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "invalid JSON", nil)
		return
	}
	rt := strings.TrimSpace(in.RefreshToken)
	if rt == "" {
		httputil.Error(r.Context(), w, http.StatusBadRequest, "refreshToken is required", nil)
		return
	}
	out, err := h.Auth.Refresh(r.Context(), rt)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "refresh failed", map[string]any{"reason": err.Error()})
		return
	}

	httputil.OK(w, out)
}

func (h *AuthHandlers) Me(w http.ResponseWriter, r *http.Request) {
	// Достаём Bearer токен из заголовка Authorization и прокидываем в gRPC metadata.
	authz := r.Header.Get("Authorization")
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") || strings.TrimSpace(parts[1]) == "" {
		httputil.Error(r.Context(), w, http.StatusUnauthorized, "missing or invalid Authorization header", nil)
		return
	}
	ctx := appauth.WithBearer(r.Context(), strings.TrimSpace(parts[1]))

	out, err := h.Auth.Me(ctx)
	if err != nil {
		status := errs.ToHTTP(err)
		httputil.Error(r.Context(), w, status, "me failed", map[string]any{"reason": err.Error()})
		return
	}
	httputil.OK(w, out)
}
