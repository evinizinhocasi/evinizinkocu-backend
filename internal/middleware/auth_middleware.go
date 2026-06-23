package middleware

import (
	"context"
	"net/http"
	"strings"

	"evinizinkocu-backend/internal/domain"
	"evinizinkocu-backend/internal/infrastructure/jwt"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	UserRoleKey   contextKey = "userRole"
	UserEmailKey  contextKey = "userEmail"
)

func Authenticate(jwtSecret string, userRepo domain.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "Authorization header missing"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error": "Invalid token format"}`, http.StatusUnauthorized)
				return
			}

			claims, err := jwt.ParseToken(parts[1], jwtSecret)
			if err != nil {
				http.Error(w, `{"error": "Invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			// Validate user is active in DB
			u, err := userRepo.GetByID(r.Context(), claims.UserID)
			if err != nil || !u.IsActive {
				http.Error(w, `{"error": "User account is inactive or not found"}`, http.StatusUnauthorized)
				return
			}

			// Attach to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roleVal := r.Context().Value(UserRoleKey)
			role, ok := roleVal.(string)
			if !ok {
				http.Error(w, `{"error": "Forbidden"}`, http.StatusForbidden)
				return
			}

			allowed := false
			for _, r := range allowedRoles {
				if r == role {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, `{"error": "Forbidden - Insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helpers to get values from context
func GetUserID(ctx context.Context) string {
	val := ctx.Value(UserIDKey)
	if id, ok := val.(string); ok {
		return id
	}
	return ""
}

func GetUserRole(ctx context.Context) string {
	val := ctx.Value(UserRoleKey)
	if role, ok := val.(string); ok {
		return role
	}
	return ""
}

func GetUserEmail(ctx context.Context) string {
	val := ctx.Value(UserEmailKey)
	if email, ok := val.(string); ok {
		return email
	}
	return ""
}
