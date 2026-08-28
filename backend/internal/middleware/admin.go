package middleware

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Admin(db *pgxpool.Pool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userIDValue := r.Context().Value(UserIDKey)

		userID, ok := userIDValue.(int64)
		if !ok {
			http.Error(w, `{"success":false,"message":"Invalid user identity."}`, http.StatusUnauthorized)
			return
		}

		var roleName string
		var status string

		err := db.QueryRow(
			context.Background(),
			`
			SELECT
				r.name,
				u.status
			FROM users u
			JOIN roles r ON r.id = u.role_id
			WHERE u.id = $1
			`,
			userID,
		).Scan(&roleName, &status)

		if err != nil {
			http.Error(w, `{"success":false,"message":"Unable to verify user authorization."}`, http.StatusInternalServerError)
			return
		}

		if status != "active" {
			http.Error(w, `{"success":false,"message":"User account is not active."}`, http.StatusForbidden)
			return
		}

		if roleName != "admin" {
			http.Error(w, `{"success":false,"message":"Administrator access required."}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
