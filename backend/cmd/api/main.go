package main

import (
	"fmt"
	"log"
	"net/http"

	"taskfund/backend/internal/handlers"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		fmt.Fprint(w, `{
			"success": true,
			"message": "TaskFunds API is running."
		}`)
	})
		mux.HandleFunc("/api/v1/auth/register", handlers.Register)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("TaskFunds API running on http://localhost:8080")

	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
