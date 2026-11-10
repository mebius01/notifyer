package httpapi_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"notifyer/internal/httpapi"
	"notifyer/internal/notification"
	"notifyer/internal/notification/memory"
)

type okSender struct{}

func (okSender) Send(context.Context, notification.Notification) error { return nil }

// newHandler builds the real router backed by an in-memory service.
func newHandler(t *testing.T) http.Handler {
	t.Helper()
	svc, err := notification.NewService(
		memory.NewRepository(),
		map[string]notification.Sender{"console": okSender{}},
	)
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return httpapi.NewHandler(svc, logger, "test")
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(method, path, strings.NewReader(body)))
	return w
}

func TestHealth(t *testing.T) {
	w := do(newHandler(t), "GET", "/health", "")

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	if want := `{"status":"ok","version":"test"}`; strings.TrimSpace(w.Body.String()) != want {
		t.Errorf("body = %s, want %s", w.Body.String(), want)
	}
}

func TestCreateNotification(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"valid", `{"recipient":"a@b.c","channel":"console"}`, http.StatusCreated},
		{"broken JSON", `{"recipient":`, http.StatusBadRequest},
		{"unknown field", `{"id":5,"recipient":"a@b.c","channel":"console"}`, http.StatusBadRequest},
		{"empty recipient", `{"recipient":"","channel":"console"}`, http.StatusBadRequest},
		{"unknown channel", `{"recipient":"a@b.c","channel":"fax"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := do(newHandler(t), "POST", "/api/notifications", tt.body)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d, body = %s", w.Code, tt.wantStatus, w.Body.String())
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q", ct)
			}
		})
	}
}

func TestGetNotification(t *testing.T) {
	h := newHandler(t)
	do(h, "POST", "/api/notifications", `{"recipient":"a@b.c","channel":"console"}`)

	tests := []struct {
		path       string
		wantStatus int
	}{
		{"/api/notifications/1", http.StatusOK},
		{"/api/notifications/999", http.StatusNotFound},
		{"/api/notifications/abc", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if w := do(h, "GET", tt.path, ""); w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestUpdateNotification(t *testing.T) {
	h := newHandler(t)
	do(h, "POST", "/api/notifications", `{"recipient":"a@b.c","channel":"console"}`)

	if w := do(h, "PATCH", "/api/notifications/1", `{"recipient":"new@b.c","channel":"console"}`); w.Code != http.StatusOK {
		t.Errorf("valid update: status = %d, body = %s", w.Code, w.Body.String())
	}
	if w := do(h, "PATCH", "/api/notifications/1", `{"recipient":"","channel":"console"}`); w.Code != http.StatusBadRequest {
		t.Errorf("empty recipient: status = %d, want 400", w.Code)
	}
	if w := do(h, "PATCH", "/api/notifications/999", `{"recipient":"a@b.c","channel":"console"}`); w.Code != http.StatusNotFound {
		t.Errorf("missing id: status = %d, want 404", w.Code)
	}
}

func TestExportNotifications(t *testing.T) {
	h := newHandler(t)
	do(h, "POST", "/api/notifications", `{"recipient":"a@b.c","channel":"console"}`)

	w := do(h, "GET", "/api/notifications/export", "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", w.Code, w.Body.String())
	}
	if got := w.Header().Get("X-Export-Count"); got != "1" {
		t.Errorf("X-Export-Count = %q, want 1", got)
	}
	if want := "id:1 recipient:a@b.c channel:console\n"; w.Body.String() != want {
		t.Errorf("body = %q, want %q", w.Body.String(), want)
	}
}
