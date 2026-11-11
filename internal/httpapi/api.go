// Package httpapi is the HTTP transport: routing, middleware, handlers and DTOs.
package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"notifyer/internal/notification"
)

// Deps are the dependencies of the HTTP API. Filled by main and tests.
type Deps struct {
	Notifications *notification.Service
	Logger        *slog.Logger
	Version       string
}

// api holds what handlers need. Deps is embedded so a new domain adds one field only;
// api-only state (caches, counters) goes here next to Deps.
type api struct {
	Deps
}

func NewHandler(d Deps) (http.Handler, error) {
	if d.Notifications == nil {
		return nil, errors.New("httpapi.NewHandler: notifications service is nil")
	}
	if d.Logger == nil {
		return nil, errors.New("httpapi.NewHandler: logger is nil")
	}

	a := &api{Deps: d}

	mux := http.NewServeMux()
	a.routes(mux)

	return a.logRequest(mux), nil
}

// routes registers all API endpoints.
func (a *api) routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", a.health)
	a.registerNotificationRoutes(mux)
}

func (a *api) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": a.Version})
}
