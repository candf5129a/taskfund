package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"taskfund/backend/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Profile(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		userID := r.Context().Value(middleware.UserIDKey)

		if userID == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "User not authenticated.",
			})
			return
		}

		var (
			id        int64
			firstName string
			lastName  string
			email     string
			username  *string
			phone     *string
			status    string
		)

		err := db.QueryRow(
			context.Background(),
			`
			SELECT
				id,
				first_name,
				last_name,
				email,
				username,
				phone,
				status
			FROM users
			WHERE id = $1
			`,
			userID,
		).Scan(
			&id,
			&firstName,
			&lastName,
			&email,
			&username,
			&phone,
			&status,
		)

		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "User profile not found.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Profile retrieved successfully.",
			"data": map[string]interface{}{
				"id":         id,
				"first_name": firstName,
				"last_name":  lastName,
				"email":      email,
				"username":   username,
				"phone":      phone,
				"status":     status,
			},
		})
	}
}

type UpdateProfileRequest struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Username  *string `json:"username"`
	Phone     *string `json:"phone"`
}

func UpdateProfile(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPut {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		userID := r.Context().Value(middleware.UserIDKey)

		if userID == nil {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "User not authenticated.",
			})
			return
		}

		var request UpdateProfileRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid request body.",
			})
			return
		}

		request.FirstName = strings.TrimSpace(request.FirstName)
		request.LastName = strings.TrimSpace(request.LastName)

		if request.FirstName == "" || request.LastName == "" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "First name and last name are required.",
			})
			return
		}

		if request.Username != nil {
			username := strings.TrimSpace(*request.Username)

			if username == "" {
				request.Username = nil
			} else {
				request.Username = &username
			}
		}

		if request.Phone != nil {
			phone := strings.TrimSpace(*request.Phone)

			if phone == "" {
				request.Phone = nil
			} else {
				request.Phone = &phone
			}
		}

		_, err = db.Exec(
			context.Background(),
			`
			UPDATE users
			SET
				first_name = $1,
				last_name = $2,
				username = $3,
				phone = $4,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $5
			`,
			request.FirstName,
			request.LastName,
			request.Username,
			request.Phone,
			userID,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to update profile.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Profile updated successfully.",
		})
	}
}
