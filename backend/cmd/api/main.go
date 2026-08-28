package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"taskfund/backend/internal/database"
	"taskfund/backend/internal/handlers"

	"taskfund/backend/internal/middleware"
)

func main() {

	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, `{
			"success": true,
			"message": "TaskFunds API is running."
		}`)
	})

	mux.HandleFunc("/api/v1/auth/register", handlers.Register(db))
	mux.HandleFunc("/api/v1/auth/verify-email", handlers.VerifyEmail(db))
	mux.Handle("/api/v1/auth/login", handlers.Login(db))
	// mux.HandleFunc("/api/v1/tasks", handlers.GetTasks(db))

	mux.Handle(
		"/api/v1/profile",
		middleware.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			switch r.Method {

			case http.MethodGet:
				handlers.Profile(db).ServeHTTP(w, r)

			case http.MethodPut:
				handlers.UpdateProfile(db).ServeHTTP(w, r)

			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
			}

		})),
	)

	mux.Handle(
		"/api/v1/tasks/",
		middleware.Auth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if strings.HasSuffix(r.URL.Path, "/accept") {
				handlers.AcceptTask(db).ServeHTTP(w, r)
				return
			}

			if strings.HasSuffix(r.URL.Path, "/submit") {
				handlers.SubmitTask(db).ServeHTTP(w, r)
				return
			}

			http.NotFound(w, r)
		})),
	)

	mux.Handle(
		"/api/v1/admin/submissions/",
		middleware.Auth(middleware.Admin(db, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/approve") {
				handlers.ApproveSubmission(db).ServeHTTP(w, r)
				return
			}

			http.NotFound(w, r)
		}))),
	)

	mux.Handle(
		"/api/v1/tasks",
		handlers.Tasks(db),
	)

	mux.Handle(
		"/api/v1/admin/submissions",
		middleware.Auth(
			middleware.Admin(db, handlers.AdminSubmissions(db)),
		))

	mux.Handle(
		"/api/v1/wallet",
		middleware.Auth(handlers.Wallet(db)),
	)

	mux.Handle(
		"/api/v1/wallet/transactions",
		middleware.Auth(handlers.WalletTransactions(db)),
	)
	
	mux.Handle(
		"/api/v1/withdrawals",
		middleware.Auth(handlers.CreateWithdrawal(db)),
	)

	mux.Handle(
		"/api/v1/admin/withdrawals",
		middleware.Auth(
			middleware.Admin(db, handlers.AdminWithdrawals(db)),
		),
	)

	mux.Handle(
		"/api/v1/admin/withdrawals/",
		middleware.Auth(
			middleware.Admin(db, handlers.AdminReviewWithdrawal(db)),
		),
	)

	// server := &http.Server{
	// 	Addr:    ":8080",
	// 	Handler: mux,
	// }

	server := &http.Server{
		Addr:    ":8080",
		Handler: middleware.CORS(mux),
	}

	log.Println("TaskFunds API running on http://localhost:8080")

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
