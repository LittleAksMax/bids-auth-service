package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// NewRouter constructs the main API router.
func NewRouter(db *sql.DB) http.Handler {
	r := chi.NewRouter()

	// Health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("db not ready"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Placeholder auth routes group
	r.Route("/auth", func(r chi.Router) {
		// r.Post("/login", loginHandler)
		// r.Post("/register", registerHandler)
	})

	return r
}
