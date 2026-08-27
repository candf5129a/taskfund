package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"taskfund/backend/internal/verification"

	"github.com/jackc/pgx/v5/pgxpool"
)

func VerifyEmail(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		token := strings.TrimSpace(r.URL.Query().Get("token"))

		if token == "" {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Verification token is required.",
			})
			return
		}

		var userID int64
		var verified bool

		err := db.QueryRow(
			context.Background(),
			`
			SELECT id, email_verified
			FROM users
			WHERE email_verification_token = $1
			  AND email_verification_expires_at > $2
			`,
			token,
			time.Now(),
		).Scan(&userID, &verified)

		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid or expired verification token.",
			})
			return
		}

		if verified {
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"success": true,
				"message": "Email is already verified.",
			})
			return
		}

		_, err = db.Exec(
			context.Background(),
			`
			UPDATE users
			SET
				email_verified = TRUE,
				email_verification_token = NULL,
				email_verification_expires_at = NULL,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
			`,
			userID,
		)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to verify email.",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Email verified successfully.",
		})
	}
}

// Keep this reference temporarily so the compiler
// confirms the verification package is available.
var _ = verification.GenerateToken
