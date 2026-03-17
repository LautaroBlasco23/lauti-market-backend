package infrastructure

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const accountIDKey contextKey = "account_id"

func RequireAuth(validate func(token string) (accountID string, err error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if !strings.HasPrefix(auth, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(auth, "Bearer ")
			accountID, err := validate(token)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), accountIDKey, accountID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AccountIDFromContext(r *http.Request) (string, bool) {
	id, ok := r.Context().Value(accountIDKey).(string)
	return id, ok
}
