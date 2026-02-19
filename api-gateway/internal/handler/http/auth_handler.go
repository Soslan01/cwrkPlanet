package http

import (
	"net/http"

	"github.com/cwrk-planet/api-gateway/internal/domain"
	"github.com/cwrk-planet/api-gateway/internal/service/auth"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService *auth.Service
	logger      *zap.Logger
}

func NewAuthHandler(authService *auth.Service, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
	}
}

// RegisterRoutes registers auth routes
func (h *AuthHandler) RegisterRoutes(r chi.Router) {
	r.Route("/v1/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)
		r.Post("/logout", h.Logout)
		r.With(h.requireAuth).Get("/me", h.Me)
	})
}

// Register handles POST /v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// Login handles POST /v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// Refresh handles POST /v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req domain.RefreshRequest
	if err := decodeJSON(r, &req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	resp, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, resp)
}

// Logout handles POST /v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req domain.LogoutRequest
	if err := decodeJSON(r, &req); err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, domain.LogoutResponse{Success: true})
}

// Me handles GET /v1/auth/me (requires authentication)
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		h.respondError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.authService.GetUser(r.Context(), userID)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.respondJSON(w, http.StatusOK, domain.MeResponse{User: user})
}

// handleServiceError maps domain errors to HTTP status codes
func (h *AuthHandler) handleServiceError(w http.ResponseWriter, err error) {
	switch err {
	case domain.ErrInvalidRequest:
		h.respondError(w, http.StatusBadRequest, err.Error(), err)
	case domain.ErrUnauthorized:
		h.respondError(w, http.StatusUnauthorized, err.Error(), err)
	case domain.ErrForbidden:
		h.respondError(w, http.StatusForbidden, err.Error(), err)
	case domain.ErrNotFound:
		h.respondError(w, http.StatusNotFound, err.Error(), err)
	case domain.ErrServiceUnavailable:
		h.respondError(w, http.StatusServiceUnavailable, err.Error(), err)
	default:
		h.logger.Error("Internal server error", zap.Error(err))
		h.respondError(w, http.StatusInternalServerError, "Internal server error", nil)
	}
}
