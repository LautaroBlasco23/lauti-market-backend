package infrastructure

import (
	"context"
	"net/http"
	"strings"

	apiDomain "github.com/LautaroBlasco23/lauti-market-backend/internal/api/domain"
	authUtils "github.com/LautaroBlasco23/lauti-market-backend/internal/auth/infrastructure/utils"
)

type contextKey string

const claimsKey contextKey = "claims"

type AuthMiddleware struct {
	jwtGen *authUtils.JWTGenerator
}

func NewAuthMiddleware(jwtGen *authUtils.JWTGenerator) *AuthMiddleware {
	return &AuthMiddleware{jwtGen: jwtGen}
}

func (m *AuthMiddleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := m.jwtGen.Validate(tokenStr)
		if err != nil {
			http.Error(w, apiDomain.ErrUnauthorized.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetClaims(ctx context.Context) (*authUtils.Claims, bool) {
	claims, ok := ctx.Value(claimsKey).(*authUtils.Claims)
	return claims, ok
}
