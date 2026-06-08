// Package middleware provides HTTP middleware for e-Dossier.
package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/edossier/api/pkg/apperror"
	"github.com/edossier/api/pkg/token"
	"github.com/google/uuid"
)

// Context keys
type ctxKey string

const (
	CtxClaims   ctxKey = "claims"
	CtxReqID    ctxKey = "request_id"
)

// ─────────────────────────────────────────────────────────────────────────────
// REQUEST ID
// ─────────────────────────────────────────────────────────────────────────────

// RequestID injects a unique request ID into each request context and response header.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		ctx := context.WithValue(r.Context(), CtxReqID, id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID retrieves the request ID from a context.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(CtxReqID).(string); ok {
		return id
	}
	return ""
}

// ─────────────────────────────────────────────────────────────────────────────
// STRUCTURED LOGGER
// ─────────────────────────────────────────────────────────────────────────────

// Logger logs method, path, status code, and latency for every request.
func Logger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			log.Info("http",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"latency_ms", time.Since(start).Milliseconds(),
				"request_id", GetRequestID(r.Context()),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.status = code
		rw.written = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// RECOVERY
// ─────────────────────────────────────────────────────────────────────────────

// Recovery catches panics, logs the stack trace, and returns a 500.
func Recovery(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered",
						"error", rec,
						"stack", string(debug.Stack()),
						"request_id", GetRequestID(r.Context()),
					)
					http.Error(w, `{"code":"INTERNAL_ERROR","message":"an unexpected error occurred"}`,
						http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// CORS
// ─────────────────────────────────────────────────────────────────────────────

// CORS sets permissive CORS headers suitable for development.
// In production, restrict AllowedOrigins.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, o := range allowedOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// JWT AUTHENTICATION
// ─────────────────────────────────────────────────────────────────────────────

// Authenticate validates the Bearer token and stores claims in context.
// Returns 401 for missing/invalid tokens.
func Authenticate(maker *token.Maker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorized(w, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				writeUnauthorized(w, "invalid authorization header format")
				return
			}

			claims, err := maker.Verify(parts[1])
			if err != nil {
				writeUnauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), CtxClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromCtx retrieves JWT claims from context.
func ClaimsFromCtx(ctx context.Context) *token.Claims {
	if c, ok := ctx.Value(CtxClaims).(*token.Claims); ok {
		return c
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// RBAC AUTHORIZATION
// ─────────────────────────────────────────────────────────────────────────────

// UserRoleChecker is satisfied by the UserRoleRepository.HasPermission method.
type UserRoleChecker interface {
	HasPermission(ctx context.Context, userID, resource, action string) (bool, error)
}

// Authorize checks that the authenticated user has resource:action permission.
func Authorize(checker UserRoleChecker, resource, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromCtx(r.Context())
			if claims == nil {
				writeUnauthorized(w, "authentication required")
				return
			}

			ok, err := checker.HasPermission(r.Context(), claims.UserID, resource, action)
			if err != nil || !ok {
				writeForbidden(w, "you do not have permission to perform this action")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoles checks that the JWT contains at least one of the given role codes.
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromCtx(r.Context())
			if claims == nil {
				writeUnauthorized(w, "authentication required")
				return
			}
			for _, role := range claims.Roles {
				if _, ok := allowed[role]; ok {
					next.ServeHTTP(w, r)
					return
				}
			}
			writeForbidden(w, "insufficient role")
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// HELPERS
// ─────────────────────────────────────────────────────────────────────────────

func writeUnauthorized(w http.ResponseWriter, msg string) {
	ae := apperror.Unauthorized(msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ae.HTTPStatus())
	w.Write([]byte(`{"code":"` + string(ae.Code) + `","message":"` + ae.Message + `"}`)) //nolint:errcheck
}

func writeForbidden(w http.ResponseWriter, msg string) {
	ae := apperror.Forbidden(msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ae.HTTPStatus())
	w.Write([]byte(`{"code":"` + string(ae.Code) + `","message":"` + ae.Message + `"}`)) //nolint:errcheck
}
