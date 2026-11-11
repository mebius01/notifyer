// Package httpapi is the HTTP transport: routing, middleware, handlers and DTOs.
package httpapi

import (
	"log/slog"
	"net/http"

	"notifyer/internal/notification"
)

type Server struct {
	notifications *notification.Service
	logger        *slog.Logger
	version       string
}

// NewHandler returns the router with all routes and middleware.
func NewHandler(notifications *notification.Service, logger *slog.Logger, version string) http.Handler {
	s := &Server{notifications: notifications, logger: logger, version: version}

	mux := http.NewServeMux()
	s.routes(mux)

	return s.logRequest(mux)
}

// routes registers all API endpoints.
func (s *Server) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /api/notifications/export", s.exportNotifications)
	mux.HandleFunc("POST /api/notifications", s.createNotification)
	mux.HandleFunc("GET /api/notifications", s.listNotifications)
	mux.HandleFunc("GET /api/notifications/{id}", s.getNotification)
	mux.HandleFunc("PATCH /api/notifications/{id}", s.updateNotification)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": s.version})
}
