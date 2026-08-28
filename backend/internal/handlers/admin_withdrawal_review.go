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

func AdminReviewWithdrawal(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		// Expected:
		// /api/v1/admin/withdrawals/{id}/approve
		// /api/v1/admin/withdrawals/{id}/reject
		if len(parts) != 6 {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Invalid withdrawal endpoint.",
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
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Invalid withdrawal action.",
			})
			return
		}

		adminIDValue := r.Context().Value(middleware.UserIDKey)

		adminID, ok := adminIDValue.(int64)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Invalid user identity.",
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
			userID   int64
			walletID int64
			amount   float64
			status   string
		)

		err = tx.QueryRow(
			ctx,
			`
			SELECT
				w.user_id,
				w.amount,
				w.status,
				wa.id
			FROM withdrawals w
			JOIN wallets wa ON wa.user_id = w.user_id
			WHERE w.id = $1
			FOR UPDATE OF w, wa
			`,
			withdrawalID,
		).Scan(
			&userID,
			&amount,
			&status,
			&walletID,
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

		if action == "approve" {

			var pendingBalance float64

			err = tx.QueryRow(
				ctx,
				`
				SELECT pending_balance
				FROM wallets
				WHERE id = $1
				FOR UPDATE
				`,
				walletID,
			).Scan(&pendingBalance)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to retrieve wallet.",
				})
				return
			}

			if pendingBalance < amount {
				writeJSON(w, http.StatusConflict, map[string]interface{}{
					"success": false,
					"message": "Wallet pending balance is inconsistent with withdrawal.",
				})
				return
			}

			_, err = tx.Exec(
				ctx,
				`
				UPDATE wallets
				SET
					pending_balance = pending_balance - $1,
					updated_at = CURRENT_TIMESTAMP
				WHERE id = $2
				`,
				amount,
				walletID,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to finalize wallet reservation.",
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
					"message": "Failed to record returned funds.",
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

		_ = adminID

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
