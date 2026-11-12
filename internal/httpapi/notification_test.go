package httpapi_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"notifyer/internal/httpapi"
	"notifyer/internal/notification"
)

type okSender struct{}

func (okSender) Send(context.Context, notification.Notification) error { return nil }

// fakeRepo is a minimal Repository: IDs start at 1 and equal index+1.
type fakeRepo struct {
	items []notification.Notification
}

var _ notification.Repository = (*fakeRepo)(nil)

func (r *fakeRepo) Save(_ context.Context, n notification.Notification) (notification.Notification, error) {
	n.ID = len(r.items) + 1
	r.items = append(r.items, n)
	return n, nil
}

func (r *fakeRepo) List(context.Context) ([]notification.Notification, error) {
	return r.items, nil
}

func (r *fakeRepo) Get(_ context.Context, id int) (notification.Notification, error) {
	if id < 1 || id > len(r.items) {
		return notification.Notification{}, notification.ErrNotFound
	}
	return r.items[id-1], nil
}

func (r *fakeRepo) Update(_ context.Context, n notification.Notification) (notification.Notification, error) {
	if n.ID < 1 || n.ID > len(r.items) {
		return notification.Notification{}, notification.ErrNotFound
	}
	r.items[n.ID-1] = n
	return n, nil
}

func (r *fakeRepo) UpdateStatus(_ context.Context, id int, status notification.NotificationStatus) (notification.Notification, error) {
	if id < 1 || id > len(r.items) {
		return notification.Notification{}, notification.ErrNotFound
	}
	r.items[id-1].Status = status
	return r.items[id-1], nil
}

// noopSpawner drops background tasks: handler tests do not need them.
type noopSpawner struct{}

func (noopSpawner) Go(func()) {}

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newHandler builds the real router backed by fakeRepo.
func newHandler(t *testing.T) http.Handler {
	t.Helper()
	svc, err := notification.NewService(
		&fakeRepo{},
		map[string]notification.Sender{"console": okSender{}},
		discardLogger(),
		noopSpawner{},
	)
	if err != nil {
		t.Fatal(err)
	}
	return mustHandler(t, svc)
}

// mustHandler builds the router with a discarding logger.
func mustHandler(t *testing.T, svc *notification.Service) http.Handler {
	t.Helper()
	h, err := httpapi.NewHandler(httpapi.Deps{
		Notifications: svc,
		Logger:        slog.New(slog.NewTextHandler(io.Discard, nil)),
		Version:       "test",
	})
	if err != nil {
		t.Fatal(err)
	}
	return h
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
		{"valid", `{"recipient":"a@b.c","channel":"console"}`, http.StatusAccepted},
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

func TestErrorBody(t *testing.T) {
	h := newHandler(t)

	tests := []struct {
		name, method, path, body string
		wantStatus               int
		wantBody                 string
	}{
		{"not found", "GET", "/api/notifications/999", "",
			http.StatusNotFound, `{"error":"notification not found","code":"notification.not_found"}`},
		{"invalid", "POST", "/api/notifications", `{"recipient":"","channel":"console"}`,
			http.StatusBadRequest, `{"error":"invalid notification: recipient cannot be empty","code":"notification.invalid"}`},
		{"unknown channel", "POST", "/api/notifications", `{"recipient":"a@b.c","channel":"fax"}`,
			http.StatusBadRequest, `{"error":"unknown channel: \"fax\"","code":"notification.unknown_channel"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := do(h, tt.method, tt.path, tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if got := strings.TrimSpace(w.Body.String()); got != tt.wantBody {
				t.Errorf("body = %s, want %s", got, tt.wantBody)
			}
		})
	}
}

// failingRepo returns an internal error that must not leak to clients.
type failingRepo struct{ fakeRepo }

func (failingRepo) List(context.Context) ([]notification.Notification, error) {
	return nil, errors.New("pq: connection refused at 10.0.0.5")
}

func TestInternalErrorIsHidden(t *testing.T) {
	svc, err := notification.NewService(&failingRepo{}, map[string]notification.Sender{"console": okSender{}}, discardLogger(), noopSpawner{})
	if err != nil {
		t.Fatal(err)
	}
	h := mustHandler(t, svc)

	w := do(h, "GET", "/api/notifications", "")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", w.Code)
	}
	if want := `{"error":"internal server error","code":"internal"}`; strings.TrimSpace(w.Body.String()) != want {
		t.Errorf("body = %s, want %s", w.Body.String(), want)
	}
}

func TestNewHandlerValidatesDeps(t *testing.T) {
	svc, err := notification.NewService(&fakeRepo{}, map[string]notification.Sender{"console": okSender{}}, discardLogger(), noopSpawner{})
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name string
		deps httpapi.Deps
	}{
		{"no notifications", httpapi.Deps{Logger: logger}},
		{"no logger", httpapi.Deps{Notifications: svc}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := httpapi.NewHandler(tt.deps); err == nil {
				t.Error("want error, got nil")
			}
		})
	}
}
