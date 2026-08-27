package handlers

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func AdminSubmissions(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		rows, err := db.Query(
			context.Background(),
			`
			SELECT
				ts.id,
				ts.task_id,
				t.title,
				ts.worker_id,
				u.first_name,
				u.last_name,
				ts.content,
				ts.proof_url,
				t.reward,
				ts.status,
				ts.submitted_at
			FROM task_submissions ts
			JOIN tasks t ON t.id = ts.task_id
			JOIN users u ON u.id = ts.worker_id
			ORDER BY ts.submitted_at DESC
			`,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve submissions.",
			})
			return
		}

		defer rows.Close()

		submissions := []map[string]interface{}{}

		for rows.Next() {

			var (
				id          int64
				taskID      int64
				taskTitle   string
				workerID    int64
				firstName   string
				lastName    string
				content     *string
				proofURL    *string
				reward      float64
				status      string
				submittedAt interface{}
			)

			err := rows.Scan(
				&id,
				&taskID,
				&taskTitle,
				&workerID,
				&firstName,
				&lastName,
				&content,
				&proofURL,
				&reward,
				&status,
				&submittedAt,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to read submissions.",
				})
				return
			}

			submissions = append(submissions, map[string]interface{}{
				"id":           id,
				"task_id":      taskID,
				"task_title":   taskTitle,
				"worker_id":    workerID,
				"worker_name":  firstName + " " + lastName,
				"content":      content,
				"proof_url":    proofURL,
				"reward":       reward,
				"status":       status,
				"submitted_at": submittedAt,
			})
		}

		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve submissions.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Submissions retrieved successfully.",
			"data":    submissions,
		})
	}
}
