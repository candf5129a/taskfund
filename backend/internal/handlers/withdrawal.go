package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskfund/backend/internal/middleware"
)

type WithdrawalRequest struct {
	Amount        float64 `json:"amount"`
	AccountName   string  `json:"account_name"`
	AccountNumber string  `json:"account_number"`
	BankName      string  `json:"bank_name"`
}

func CreateWithdrawal(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]interface{}{
				"success": false,
				"message": "Method not allowed.",
			})
			return
		}

		userIDValue := r.Context().Value(middleware.UserIDKey)

		userID, ok := userIDValue.(int64)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"message": "Invalid user identity.",
			})
			return
		}

		var request WithdrawalRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Invalid request body.",
			})
			return
		}

		request.AccountName = strings.TrimSpace(request.AccountName)
		request.AccountNumber = strings.TrimSpace(request.AccountNumber)
		request.BankName = strings.TrimSpace(request.BankName)

		if request.Amount <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Withdrawal amount must be greater than zero.",
			})
			return
		}

		if request.AccountName == "" ||
			request.AccountNumber == "" ||
			request.BankName == "" {

			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Bank account details are required.",
			})
			return
		}

		ctx := context.Background()

		tx, err := db.Begin(ctx)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to start withdrawal.",
			})
			return
		}

		defer tx.Rollback(ctx)

		// Lock the wallet so two withdrawals cannot spend
		// the same balance simultaneously.
		var walletID int64
		var balance float64
		var pendingBalance float64

		err = tx.QueryRow(
			ctx,
			`
			SELECT id, balance, pending_balance
			FROM wallets
			WHERE user_id = $1
			FOR UPDATE
			`,
			userID,
		).Scan(&walletID, &balance, &pendingBalance)

		if err == pgx.ErrNoRows {
			writeJSON(w, http.StatusNotFound, map[string]interface{}{
				"success": false,
				"message": "Wallet not found.",
			})
			return
		}

		if err != nil {
			log.Printf("withdrawal wallet query failed: user_id=%d error=%v", userID, err)

			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to retrieve wallet.",
			})
			return
		}

		if request.Amount > balance {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"message": "Insufficient wallet balance.",
			})
			return
		}

		// Reserve the withdrawal amount immediately.
		_, err = tx.Exec(
			ctx,
			`
        UPDATE wallets
        SET
                balance = balance - $1,
                pending_balance = pending_balance + $1,
                updated_at = CURRENT_TIMESTAMP
        WHERE id = $2
        `,
			request.Amount,
			walletID,
		)

		if err != nil {
			log.Printf("withdrawal wallet update failed: user_id=%d wallet_id=%d amount=%f error=%v",
				userID, walletID, request.Amount, err)

			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to reserve withdrawal amount.",
			})
			return
		}

		var withdrawalID int64

		err = tx.QueryRow(
			ctx,
			`
			INSERT INTO withdrawals (
				user_id,
				amount,
				account_name,
				account_number,
				bank_name,
				status
			)
			VALUES ($1, $2, $3, $4, $5, 'pending')
			RETURNING id
			`,
			userID,
			request.Amount,
			request.AccountName,
			request.AccountNumber,
			request.BankName,
		).Scan(&withdrawalID)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to create withdrawal request.",
			})
			return
		}

		// Record the wallet debit.
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
				withdrawal_id
			)
				VALUES (
					$1,
					$2,
					$3,
					'debit',
					'Withdrawal request',
					$4
			)
			RETURNING id
			`,
			userID,
			walletID,
			request.Amount,
			withdrawalID,
		).Scan(&transactionID)

		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to record withdrawal transaction.",
			})
			return
		}

		if err := tx.Commit(ctx); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"message": "Failed to complete withdrawal request.",
			})
			return
		}

		writeJSON(w, http.StatusCreated, map[string]interface{}{
			"success": true,
			"message": "Withdrawal request created successfully.",
			"data": map[string]interface{}{
				"withdrawal_id":  withdrawalID,
				"transaction_id": transactionID,
				"amount":         request.Amount,
				"status":         "pending",
			},
		})
	}
}
