package http

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/cwrk-planet/api-gateway/internal/service/auth"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
)

// RequestIDMiddleware generates and adds request ID to context
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			requestID := getRequestIDFromContext(r.Context())

			logger.Info("HTTP request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration),
				zap.String("request_id", requestID),
				zap.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

// authMiddleware validates JWT token and adds user ID to context
func authMiddleware(authService *auth.Service, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "Missing Authorization header", nil)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				respondError(w, http.StatusUnauthorized, "Invalid Authorization header format", nil)
				return
			}

			token := parts[1]
			valid, userID, err := authService.ValidateToken(r.Context(), token)
			if err != nil {
				logger.Error("Token validation error", zap.Error(err))
				respondError(w, http.StatusUnauthorized, "Invalid token", nil)
				return
			}

			if !valid {
				respondError(w, http.StatusUnauthorized, "Invalid token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// requireAuth is a helper that wraps authMiddleware for use with chi
func (h *AuthHandler) requireAuth(next http.Handler) http.Handler {
	return authMiddleware(h.authService, h.logger)(next)
}

// Helper functions

func getRequestIDFromContext(ctx context.Context) string {
	// Use chi's request ID if available
	if id := middleware.GetReqID(ctx); id != "" {
		return id
	}
	// Fall back to our custom key
	if id, ok := ctx.Value(requestIDKey).(string); ok {
		return id
	}
	return ""
}

func getUserIDFromContext(ctx context.Context) (int64, bool) {
	if id, ok := ctx.Value(userIDKey).(int64); ok {
		return id, true
	}
	return 0, false
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// decodeJSON decodes JSON request body
func decodeJSON(r *http.Request, v interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return err
	}
	return nil
}

// respondJSON sends JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondError sends error response
func respondError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]string{
		"error": message,
	}
	if err != nil {
		response["details"] = err.Error()
	}
	respondJSON(w, status, response)
}

// These are exported for use in auth_handler.go
func (h *AuthHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	respondJSON(w, status, data)
}

func (h *AuthHandler) respondError(w http.ResponseWriter, status int, message string, err error) {
	respondError(w, status, message, err)
}
