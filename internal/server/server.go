// Package server provides HTTP server setup and routing.
package server

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ttani03/gotha-boilerplate/internal/config"
	"github.com/ttani03/gotha-boilerplate/internal/db/generated"
	"github.com/ttani03/gotha-boilerplate/internal/handler"
	"github.com/ttani03/gotha-boilerplate/internal/middleware"
)

// New creates and configures the HTTP server.
func New(db *pgxpool.Pool, cfg *config.Config) http.Handler {
	queries := generated.New(db)

	// Handlers
	homeHandler := handler.NewHomeHandler()
	authHandler := handler.NewAuthHandler(db, queries, cfg)
	todoHandler := handler.NewTodoHandler(db, queries)

	// Auth middleware
	requireAuth := middleware.Auth(cfg.JWTSecret)

	mux := http.NewServeMux()

	// Static files
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Public routes
	mux.HandleFunc("GET /{$}", homeHandler.Index)
	mux.HandleFunc("GET /login", authHandler.LoginPage)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("GET /register", authHandler.RegisterPage)
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /logout", authHandler.Logout)
	mux.HandleFunc("POST /auth/refresh", authHandler.RefreshToken)

	// Protected routes (require authentication)
	mux.Handle("GET /todos", requireAuth(http.HandlerFunc(todoHandler.ListPage)))
	mux.Handle("POST /todos", requireAuth(http.HandlerFunc(todoHandler.Create)))
	mux.Handle("PUT /todos/{id}", requireAuth(http.HandlerFunc(todoHandler.Update)))
	mux.Handle("PATCH /todos/{id}/toggle", requireAuth(http.HandlerFunc(todoHandler.Toggle)))
	mux.Handle("DELETE /todos/{id}", requireAuth(http.HandlerFunc(todoHandler.Delete)))

	// Apply global middleware
	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	return h
}
