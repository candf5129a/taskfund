package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateTask(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		var request struct {
			CampaignID  *int64  `json:"campaign_id"`
			CategoryID  *int64  `json:"category_id"`
			Title       string  `json:"title"`
			Description string  `json:"description"`
			Reward      float64 `json:"reward"`
			Slots       int     `json:"slots"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid request body.",
			})
			return
		}

		if request.Title == "" ||
			request.Description == "" ||
			request.Reward <= 0 ||
			request.Slots <= 0 {

			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Title, description, reward and slots are required.",
			})
			return
		}

		var taskID int64

		err = db.QueryRow(
			context.Background(),
			`
			INSERT INTO tasks (
				campaign_id,
				category_id,
				title,
				description,
				reward,
				slots,
				slots_remaining,
				status
			)
			VALUES ($1, $2, $3, $4, $5, $6, $6, 'active')
			RETURNING id
			`,
			request.CampaignID,
			request.CategoryID,
			request.Title,
			request.Description,
			request.Reward,
			request.Slots,
		).Scan(&taskID)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "failed to create task.",
			})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"success": true,
			"message": "Task created successfully.",
			"data": map[string]interface{}{
				"id": taskID,
			},
		})
	}
}

func Tasks(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			GetTasks(db).ServeHTTP(w, r)

		case http.MethodPost:
			CreateTask(db).ServeHTTP(w, r)

		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
		}
	}
}

func GetTasks(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		rows, err := db.Query(
			context.Background(),
			`
			SELECT
				id,
				campaign_id,
				category_id,
				title,
				description,
				reward,
				slots,
				slots_remaining,
				status
			FROM tasks
			WHERE status = 'active'
			ORDER BY created_at DESC
			`,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve tasks.",
			})
			return
		}
		defer rows.Close()

		tasks := []map[string]interface{}{}

		for rows.Next() {

			var (
				id             int64
				campaignID     *int64
				categoryID     *int64
				title          string
				description    string
				reward         float64
				slots          int
				slotsRemaining int
				status         string
			)

			err := rows.Scan(
				&id,
				&campaignID,
				&categoryID,
				&title,
				&description,
				&reward,
				&slots,
				&slotsRemaining,
				&status,
			)

			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
					"success": false,
					"message": "Failed to read tasks.",
				})
				return
			}

			tasks = append(tasks, map[string]interface{}{
				"id":              id,
				"campaign_id":     campaignID,
				"category_id":     categoryID,
				"title":           title,
				"description":     description,
				"reward":          reward,
				"slots":           slots,
				"slots_remaining": slotsRemaining,
				"status":          status,
			})
		}

		if err := rows.Err(); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve tasks.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Tasks retrieved successfully.",
			"data":    tasks,
		})
	}
}
