package handlers

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"taskfund/backend/internal/middleware"
)

func Wallet(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		userIDValue := r.Context().Value(middleware.UserIDKey)

		userID, ok := userIDValue.(int64)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Invalid user identity.",
			})
			return
		}

		var (
			walletID int64
			balance  float64
		)

		err := db.QueryRow(
			context.Background(),
			`
			SELECT id, balance
			FROM wallets
			WHERE user_id = $1
			`,
			userID,
		).Scan(&walletID, &balance)

		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Wallet not found.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Wallet retrieved successfully.",
			"data": map[string]interface{}{
				"id":      walletID,
				"balance": balance,
			},
		})
	}
}
