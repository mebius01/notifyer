package main

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"order-platform/internal/notification"
	"order-platform/internal/sender"
)

// MockSender is a mock implementation of the Sender
type MockSender struct {
	sentNotifications []notification.Notification
}

func (ms *MockSender) Send(n notification.Notification) error {
	ms.sentNotifications = append(ms.sentNotifications, n)
	return nil
}

func TestHealthHandler(t *testing.T) {
	// io.Discard is used to discard the log output during testing
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	senders := map[string]sender.Sender{"console": &MockSender{}}

	// Create a new server instance
	s, err := NewServer(logger, senders)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Create a new HTTP request to the /health endpoint
	r := httptest.NewRequest("GET", "/health", nil)
	// Create a ResponseRecorder to record the response
	w := httptest.NewRecorder()

	// Call the HealthHandler directly
	s.healthHandler(w, r)

	if w.Code != 200 {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	if !strings.Contains(w.Body.String(), "ok") {
		t.Errorf("Expected response body to contain 'ok', got %s", w.Body.String())
	}

}

func TestCreateNotification(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		wantStatus int
		wantCalls  int
	}{
		{
			name:       "Valid Notification",
			payload:    `{"recipient": "user@test.com", "subject": "Test", "message": "This is a test.", "channel": "console", "is_urgent": false}`,
			wantStatus: http.StatusCreated,
			wantCalls:  1,
		},
		{
			name:       "Invalid JSON",
			payload:    `{"recipient": "", "subject": "Test", "message": "This is a test.", "channel": "console", "is_urgent": false}`,
			wantStatus: http.StatusBadRequest,
			wantCalls:  0,
		},
		{
			name:       "Unknown Channel",
			payload:    `{"recipient": "user@test.com", "subject": "Test", "message": "This is a test.", "channel": "unknown", "is_urgent": false}`,
			wantStatus: http.StatusBadRequest,
			wantCalls:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockSender{}
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			senders := map[string]sender.Sender{"console": mock}

			s, err := NewServer(logger, senders)
			if err != nil {
				t.Fatalf("Failed to create server: %v", err)
			}

			r := httptest.NewRequest("POST", "/api/notifications", strings.NewReader(tt.payload))
			w := httptest.NewRecorder()

			s.createNotification(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("Expected status code %d, got %d", tt.wantStatus, w.Code)
			}

			if len(mock.sentNotifications) != tt.wantCalls {
				t.Errorf("Expected %d calls to Send, got %d", tt.wantCalls, len(mock.sentNotifications))
			}
		})
	}
}
