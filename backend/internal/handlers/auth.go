package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"taskfund/backend/internal/models"
)

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
			"success": false,
			"message": "Method not allowed.",
		})
		return
	}

	var request RegisterRequest

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
	request.Email = strings.TrimSpace(request.Email)

	if request.FirstName == "" ||
		request.LastName == "" ||
		request.Email == "" ||
		request.Password == "" {

		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"message": "All fields are required.",
		})
		return
	}

	user := models.User{
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Email:     request.Email,
		Password:  request.Password,
	}

	// Temporary response.
	// Database and password hashing will be added next.
	_ = user

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"message": "Registration request received.",
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}
