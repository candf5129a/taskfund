package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

var jwtSecret = []byte("CHANGE_THIS_TO_A_LONG_RANDOM_SECRET")

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, `{"success":false,"message":"Authorization token required."}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"success":false,"message":"Invalid authorization format."}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenSignatureInvalid
			}

			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"success":false,"message":"Invalid or expired token."}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			http.Error(w, `{"success":false,"message":"Invalid token claims."}`, http.StatusUnauthorized)
			return
		}

		userIDValue, ok := claims["user_id"]

		if !ok {
			http.Error(w, `{"success":false,"message":"User ID missing from token."}`, http.StatusUnauthorized)
			return
		}

		userIDFloat, ok := userIDValue.(float64)

		if !ok {
			http.Error(w, `{"success":false,"message":"Invalid user ID in token."}`, http.StatusUnauthorized)
			return
		}

		userID := int64(userIDFloat)

		ctx := context.WithValue(
			r.Context(),
			UserIDKey,
			userID,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
