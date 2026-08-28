package handlers

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"taskfund/backend/internal/middleware"
)

func WalletTransactions(db *pgxpool.Pool) http.HandlerFunc {
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

		rows, err := db.Query(
			context.Background(),
			`
			SELECT
				id,
				amount,
				type,
				description,
				task_id,
				submission_id,
				withdrawal_id,
				created_at
			FROM transactions
			WHERE user_id = $1
			ORDER BY created_at DESC
			`,
			userID,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve transactions.",
			})
			return
		}

		defer rows.Close()

		transactions := make([]map[string]interface{}, 0)

		for rows.Next() {

			var (
				id              int64
				amount          float64
				transactionType string
				description     *string
				taskID          *int64
				submissionID    *int64
				withdrawalID    *int64
				createdAt       interface{}
			)

			err := rows.Scan(
				&id,
				&amount,
				&transactionType,
				&description,
				&taskID,
				&submissionID,
				&withdrawalID,
				&createdAt,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to read transaction.",
				})
				return
			}

			transactions = append(transactions, map[string]interface{}{
				"id":            id,
				"amount":        amount,
				"type":          transactionType,
				"description":   description,
				"task_id":       taskID,
				"submission_id": submissionID,
				"withdrawal_id": withdrawalID,
				"created_at":    createdAt,
			})
		}

		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve transactions.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Transactions retrieved successfully.",
			"data":    transactions,
		})
	}
}
