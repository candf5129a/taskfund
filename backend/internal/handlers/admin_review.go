package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ApproveSubmission(db *pgxpool.Pool) http.HandlerFunc {
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
			parts[3] != "submissions" ||
			parts[5] != "approve" {

			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid submission ID.",
			})
			return
		}

		submissionID, err := strconv.ParseInt(parts[4], 10, 64)

		if err != nil || submissionID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid submission ID.",
			})
			return
		}

		ctx := context.Background()

		tx, err := db.Begin(ctx)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to start transaction.",
			})
			return
		}

		defer tx.Rollback(ctx)

		// Get submission, worker and reward.
		var (
			taskID   int64
			workerID int64
			reward   float64
		)

		err = tx.QueryRow(
			ctx,
			`
			SELECT
				ts.task_id,
				ts.worker_id,
				t.reward
			FROM task_submissions ts
			JOIN tasks t ON t.id = ts.task_id
			WHERE ts.id = $1
			  AND ts.status = 'pending'
			FOR UPDATE OF ts
			`,
			submissionID,
		).Scan(
			&taskID,
			&workerID,
			&reward,
		)

		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Pending submission not found.",
			})
			return
		}

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve submission.",
			})
			return
		}

		// Create wallet if the worker does not have one.
		var walletID int64

		err = tx.QueryRow(
			ctx,
			`
			INSERT INTO wallets (user_id, balance)
			VALUES ($1, 0)
			ON CONFLICT (user_id)
			DO UPDATE SET updated_at = CURRENT_TIMESTAMP
			RETURNING id
			`,
			workerID,
		).Scan(&walletID)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to create worker wallet.",
			})
			return
		}

		// Add reward to wallet.
		_, err = tx.Exec(
			ctx,
			`
			UPDATE wallets
			SET
				balance = balance + $1,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
			`,
			reward,
			walletID,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to update wallet balance.",
			})
			return
		}

		// Record the reward transaction.
		var transactionID int64

		err = tx.QueryRow(
			ctx,
			`
			INSERT INTO transactions (
				user_id,
				wallet_id,
				amount,
				type,
				description,
				task_id,
				submission_id
			)
			VALUES (
				$1,
				$2,
				$3,
				'credit',
				$4,
				$5,
				$6
			)
			RETURNING id
			`,
			workerID,
			walletID,
			reward,
			"Reward for completed task",
			taskID,
			submissionID,
		).Scan(&transactionID)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		// Mark submission as approved.
		_, err = tx.Exec(
			ctx,
			`
			UPDATE task_submissions
			SET
				status = 'approved',
				reviewed_at = CURRENT_TIMESTAMP,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
			`,
			submissionID,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to approve submission.",
			})
			return
		}

		// Commit everything.
		if err := tx.Commit(ctx); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to complete reward payment.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Submission approved and reward credited successfully.",
			"data": map[string]interface{}{
				"submission_id":  submissionID,
				"task_id":        taskID,
				"worker_id":      workerID,
				"wallet_id":      walletID,
				"transaction_id": transactionID,
				"reward":         reward,
				"status":         "approved",
			},
		})
	}
}
