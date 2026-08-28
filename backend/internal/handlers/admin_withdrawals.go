package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskfund/backend/internal/middleware"
)

func AdminWithdrawals(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		userIDValue := r.Context().Value(middleware.UserIDKey)

		if userIDValue == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "User not authenticated.",
			})
			return
		}

		rows, err := db.Query(
			context.Background(),
			`
			SELECT
				w.id,
				w.user_id,
				u.first_name,
				u.last_name,
				w.amount,
				w.account_name,
				w.account_number,
				w.bank_name,
				w.status,
				w.created_at,
				w.reviewed_at
			FROM withdrawals w
			JOIN users u ON u.id = w.user_id
			ORDER BY w.created_at DESC
			`,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve withdrawals.",
			})
			return
		}

		defer rows.Close()

		withdrawals := make([]map[string]interface{}, 0)

		for rows.Next() {
			var (
				id            int64
				userID        int64
				firstName     string
				lastName      string
				amount        float64
				accountName   string
				accountNumber string
				bankName      string
				status        string
				createdAt     interface{}
				reviewedAt    interface{}
			)

			err := rows.Scan(
				&id,
				&userID,
				&firstName,
				&lastName,
				&amount,
				&accountName,
				&accountNumber,
				&bankName,
				&status,
				&createdAt,
				&reviewedAt,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to read withdrawal.",
				})
				return
			}

			withdrawals = append(withdrawals, map[string]interface{}{
				"id":             id,
				"user_id":        userID,
				"worker_name":    firstName + " " + lastName,
				"amount":         amount,
				"account_name":   accountName,
				"account_number": accountNumber,
				"bank_name":      bankName,
				"status":         status,
				"created_at":     createdAt,
				"reviewed_at":    reviewedAt,
			})
		}

		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve withdrawals.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Withdrawals retrieved successfully.",
			"data":    withdrawals,
		})
	}
}

func ReviewWithdrawal(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(parts) != 6 ||
			parts[0] != "api" ||
			parts[1] != "v1" ||
			parts[2] != "admin" ||
			parts[3] != "withdrawals" {

			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid withdrawal request.",
			})
			return
		}

		withdrawalID, err := strconv.ParseInt(parts[4], 10, 64)

		if err != nil || withdrawalID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid withdrawal ID.",
			})
			return
		}

		action := parts[5]

		if action != "approve" && action != "reject" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid withdrawal action.",
			})
			return
		}

		ctx := context.Background()

		tx, err := db.Begin(ctx)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to start withdrawal review.",
			})
			return
		}

		defer tx.Rollback(ctx)

		var (
			userID int64
			amount float64
			status string
		)

		err = tx.QueryRow(
			ctx,
			`
			SELECT user_id, amount, status
			FROM withdrawals
			WHERE id = $1
			FOR UPDATE
			`,
			withdrawalID,
		).Scan(
			&userID,
			&amount,
			&status,
		)

		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Withdrawal not found.",
			})
			return
		}

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve withdrawal.",
			})
			return
		}

		if status != "pending" {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"success": false,
				"message": "Withdrawal has already been reviewed.",
			})
			return
		}

		var walletID int64

		err = tx.QueryRow(
			ctx,
			`
			SELECT id
			FROM wallets
			WHERE user_id = $1
			FOR UPDATE
			`,
			userID,
		).Scan(&walletID)

		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Wallet not found.",
			})
			return
		}

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve wallet.",
			})
			return
		}

		if action == "approve" {

			_, err = tx.Exec(
				ctx,
				`
				UPDATE wallets
				SET
					pending_balance = pending_balance - $1,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = $2
				  AND pending_balance >= $1
				`,
				amount,
				walletID,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to finalize withdrawal.",
				})
				return
			}

			_, err = tx.Exec(
				ctx,
				`
				UPDATE withdrawals
				SET
					status = 'approved',
					reviewed_at = CURRENT_TIMESTAMP,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = $1
				`,
				withdrawalID,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to approve withdrawal.",
				})
				return
			}

		} else {

			_, err = tx.Exec(
				ctx,
				`
				UPDATE wallets
				SET
					balance = balance + $1,
					pending_balance = pending_balance - $1,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = $2
				  AND pending_balance >= $1
				`,
				amount,
				walletID,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to return withdrawal funds.",
				})
				return
			}

			_, err = tx.Exec(
				ctx,
				`
				INSERT INTO transactions (
					user_id,
					wallet_id,
					amount,
					type,
					description,
					withdrawal_id
				)
				VALUES (
					$1,
					$2,
					$3,
					'credit',
					'Withdrawal rejected - funds returned',
					$4
				)
				`,
				userID,
				walletID,
				amount,
				withdrawalID,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to record returned withdrawal funds.",
				})
				return
			}

			_, err = tx.Exec(
				ctx,
				`
				UPDATE withdrawals
				SET
					status = 'rejected',
					reviewed_at = CURRENT_TIMESTAMP,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = $1
				`,
				withdrawalID,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to reject withdrawal.",
				})
				return
			}
		}

		if err := tx.Commit(ctx); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to complete withdrawal review.",
			})
			return
		}

		finalStatus := "approved"
		if action == "reject" {
			finalStatus = "rejected"
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Withdrawal reviewed successfully.",
			"data": map[string]interface{}{
				"withdrawal_id": withdrawalID,
				"amount":        amount,
				"status":        finalStatus,
			},
		})
	}
}
