package mymiddlewares

import (
	"context"
	"net/http"
	"strings"
	"tekora/services/gateway/internal/myjwt"
)

type contextKey string

const userIDCtxKey contextKey = "userID"

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDCtxKey).(string)
	return userID, ok && userID != ""
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		userID, err := myjwt.VerifyToken(parts[1])
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(
			r.Context(),
			userIDCtxKey,
			userID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
