package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"taskfund/backend/internal/auth"
	"taskfund/backend/internal/models"
)

type RegisterRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func Register(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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
		request.Email = strings.ToLower(strings.TrimSpace(request.Email))

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

		passwordHash, err := bcrypt.GenerateFromPassword(
			[]byte(request.Password),
			bcrypt.DefaultCost,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to secure password.",
			})
			return
		}

		var userID int64

		err = db.QueryRow(
			context.Background(),
			`
			INSERT INTO users
				(first_name, last_name, email, password_hash)
			VALUES
				($1, $2, $3, $4)
			RETURNING id
			`,
			request.FirstName,
			request.LastName,
			request.Email,
			string(passwordHash),
		).Scan(&userID)

		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Failed to create account.",
			})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"success": true,
			"message": "Registration successful.",
			"data": map[string]interface{}{
				"id":    userID,
				"email": request.Email,
			},
		})
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		var request LoginRequest

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid request body.",
			})
			return
		}

		request.Email = strings.ToLower(strings.TrimSpace(request.Email))

		if request.Email == "" || request.Password == "" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Email and password are required.",
			})
			return
		}

		var user models.User

		err = db.QueryRow(
			context.Background(),
			`
			SELECT id, first_name, last_name, email, password_hash
			FROM users
			WHERE email = $1
			`,
			request.Email,
		).Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.PasswordHash,
		)

		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Invalid email or password.",
			})
			return
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(user.PasswordHash),
			[]byte(request.Password),
		)

		if err != nil {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Invalid email or password.",
			})
			return
		}

		token, err := auth.GenerateToken(user.ID, user.Email)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to create authentication token.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Login successful.",
			"data": map[string]interface{}{
				"id":           user.ID,
				"first_name":   user.FirstName,
				"last_name":    user.LastName,
				"email":        user.Email,
				"access_token": token,
			},
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(data)
}
