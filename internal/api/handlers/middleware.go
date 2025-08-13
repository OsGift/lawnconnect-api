package handlers

import (
	"context"
	"net/http"
	"os"
	"strings"

	httpresponse "lawnconnect-api/internal/api/http"
	"lawnconnect-api/internal/core/services"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// contextKey is a custom type to avoid context key collisions.
type contextKey string

const UserContextKey contextKey = "user"

// AuthMiddleware is a middleware to protect private routes.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			httpresponse.JSONError(w, http.StatusUnauthorized, "Authorization header is missing")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			httpresponse.JSONError(w, http.StatusUnauthorized, "Authorization header must be 'Bearer <token>'")
			return
		}

		tokenString := parts[1]
		claims := &services.Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				httpresponse.JSONError(w, http.StatusUnauthorized, "Invalid token signature")
				return
			}
			httpresponse.JSONError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		if !token.Valid {
			httpresponse.JSONError(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		// Add user ID and role to the request context
		ctx := context.WithValue(r.Context(), UserContextKey, claims.UserID)
		ctx = context.WithValue(ctx, "userRole", claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RoleMiddleware checks if the user has the required role to access a resource.
func RoleMiddleware(requiredRole string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value("userRole").(string)
			if !ok || userRole != requiredRole {
				httpresponse.JSONError(w, http.StatusForbidden, "Access denied: Insufficient privileges")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
