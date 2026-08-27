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

func AcceptTask(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		// Get logged-in user
		userID := r.Context().Value(middleware.UserIDKey)

		if userID == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "User not authenticated.",
			})
			return
		}

		// Get task ID from URL
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")

		if len(parts) != 5 ||
			parts[0] != "api" ||
			parts[1] != "v1" ||
			parts[2] != "tasks" ||
			parts[4] != "accept" {

			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid task ID.",
			})
			return
		}

		taskID, err := strconv.ParseInt(parts[3], 10, 64)

		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid task ID.",
			})
			return
		}
		ctx := context.Background()

		// Check that the task exists and has available slots.
		var slotsRemaining int

		err = db.QueryRow(
			ctx,
			`
			SELECT slots_remaining
			FROM tasks
			WHERE id = $1
			  AND status = 'active'
			`,
			taskID,
		).Scan(&slotsRemaining)

		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Task not found or inactive.",
			})
			return
		}

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve task.",
			})
			return
		}

		if slotsRemaining <= 0 {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"success": false,
				"message": "No slots remaining.",
			})
			return
		}

		// Check if this user already accepted the task.
		var existingSubmissionID int64

		err = db.QueryRow(
			ctx,
			`
			SELECT id
			FROM task_submissions
			WHERE task_id = $1
			  AND worker_id = $2
			LIMIT 1
			`,
			taskID,
			userID,
		).Scan(&existingSubmissionID)

		if err == nil {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"success": false,
				"message": "You have already accepted this task.",
			})
			return
		}

		if err != pgx.ErrNoRows {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to check task submission.",
			})
			return
		}

		// Create the submission and decrease the slot.
		tx, err := db.Begin(ctx)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to start transaction.",
			})
			return
		}

		defer tx.Rollback(ctx)

		var submissionID int64

		err = tx.QueryRow(
			ctx,
			`
			INSERT INTO task_submissions (
				task_id,
				worker_id,
				status
			)
			VALUES ($1, $2, 'pending')
			RETURNING id
			`,
			taskID,
			userID,
		).Scan(&submissionID)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to accept task.",
			})
			return
		}

		result, err := tx.Exec(
			ctx,
			`
			UPDATE tasks
			SET slots_remaining = slots_remaining - 1,
			    updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
			  AND status = 'active'
			  AND slots_remaining > 0
			`,
			taskID,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to reserve task slot.",
			})
			return
		}

		if result.RowsAffected() != 1 {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"success": false,
				"message": "No slots remaining.",
			})
			return
		}

		if err := tx.Commit(ctx); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to complete task acceptance.",
			})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"success": true,
			"message": "Task accepted successfully.",
			"data": map[string]interface{}{
				"task_id":       taskID,
				"submission_id": submissionID,
			},
		})
	}
}
