package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"taskfund/backend/internal/middleware"
)

type SubmitTaskRequest struct {
	Content  string  `json:"content"`
	ProofURL *string `json:"proof_url"`
}

func SubmitTask(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
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

		var userID int64

		switch id := userIDValue.(type) {
		case int64:
			userID = id

		case float64:
			userID = int64(id)

		default:
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Invalid user identity.",
			})
			return
		}

		taskIDText := strings.TrimPrefix(r.URL.Path, "/api/v1/tasks/")
		taskIDText = strings.TrimSuffix(taskIDText, "/submit")

		taskID, err := strconv.ParseInt(taskIDText, 10, 64)

		if err != nil || taskID <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid task ID.",
			})
			return
		}

		var request SubmitTaskRequest

		err = json.NewDecoder(r.Body).Decode(&request)

		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid request body.",
			})
			return
		}

		request.Content = strings.TrimSpace(request.Content)

		if request.ProofURL != nil {
			proofURL := strings.TrimSpace(*request.ProofURL)

			if proofURL == "" {
				request.ProofURL = nil
			} else {
				request.ProofURL = &proofURL
			}
		}

		if request.Content == "" && request.ProofURL == nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Proof content or proof URL is required.",
			})
			return
		}

		var submissionID int64

		err = db.QueryRow(
			context.Background(),
			`
	UPDATE task_submissions
	SET
		content = $1,
		proof_url = $2,
		status = 'pending',
		updated_at = CURRENT_TIMESTAMP
	WHERE task_id = $3
	  AND worker_id = $4
	  AND status = 'pending'
	RETURNING id
	`,
			request.Content,
			request.ProofURL,
			taskID,
			userID,
		).Scan(&submissionID)

		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Unable to submit proof. Make sure you have accepted this task and have a pending submission.",
			})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"success": true,
			"message": "Proof submitted successfully.",
			"data": map[string]interface{}{
				"submission_id": submissionID,
				"task_id":       taskID,
				"status":        "pending",
			},
		})
	}
}
