package main

import (
	"fmt"
	"log"
	"net/http"

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
	mux.Handle("/api/v1/auth/login", handlers.Login(db))

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

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("TaskFunds API running on http://localhost:8080")

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
