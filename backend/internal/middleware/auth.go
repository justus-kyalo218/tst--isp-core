package middleware

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKey string

const (
	ctxRole  ctxKey = "role"
	ctxEmail ctxKey = "email"
)

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := map[string]struct{}{}
	for _, r := range allowedRoles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, email, err := parseToken(r)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if len(allowed) > 0 {
				if _, ok := allowed[role]; !ok {
					http.Error(w, "forbidden", http.StatusForbidden)
					return
				}
			}

			ctx := context.WithValue(r.Context(), ctxRole, role)
			ctx = context.WithValue(ctx, ctxEmail, email)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RoleFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxRole).(string); ok {
		return v
	}
	return ""
}

func EmailFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxEmail).(string); ok {
		return v
	}
	return ""
}

func parseToken(r *http.Request) (string, string, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", "", errors.New("missing authorization header")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", "", errors.New("invalid authorization header")
	}

	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		return "", "", errors.New("missing JWT_SECRET")
	}

	token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", errors.New("invalid claims")
	}

	role, _ := claims["role"].(string)
	email, _ := claims["sub"].(string)
	if role == "" || email == "" {
		return "", "", errors.New("missing claims")
	}
	return role, email, nil
}
