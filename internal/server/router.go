package server

import (
	"net/http"
	"time"

	"github.com/NetPo4ki/reward-system/internal/auth"
	"github.com/NetPo4ki/reward-system/internal/config"
	"github.com/NetPo4ki/reward-system/internal/handlers"
	"github.com/NetPo4ki/reward-system/internal/repo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(cfg config.Config, pool *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	authHandler := &handlers.AuthHandler{
		Users:     repo.NewUsersRepo(pool),
		JWTSecret: cfg.JWTSecret,
		TokenTTL:  24 * time.Hour,
	}

	r.Post("/auth/signup", authHandler.Signup)

	r.Group(func(pr chi.Router) {
		pr.Use(auth.Middleware(cfg.JWTSecret))
	})

	return r
}
