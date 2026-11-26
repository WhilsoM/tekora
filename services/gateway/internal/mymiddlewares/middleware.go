package mymiddlewares

import (
	"context"
	"net/http"
	"strings"
	"tekora/services/gateway/internal/httpjson"
	"tekora/services/gateway/internal/myjwt"
)

type contextKey int

const userIDKey contextKey = iota

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func AuthMiddleware(
	tokenManager *myjwt.TokenManager,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			parts := strings.Fields(r.Header.Get("Authorization"))

			if len(parts) != 2 ||
				!strings.EqualFold(parts[0], "Bearer") {
				httpjson.WriteJSONError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
					"missing or invalid authorization header",
				)
				return
			}

			claims, err := tokenManager.VerifyAccessToken(parts[1])
			if err != nil {
				httpjson.WriteJSONError(
					w,
					http.StatusUnauthorized,
					"unauthorized",
					"invalid or expired access token",
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userIDKey,
				claims.Subject,
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
