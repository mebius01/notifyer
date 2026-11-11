package sender_test

import (
	"bytes"
	"context"
	"testing"

	"notifyer/internal/notification"
	"notifyer/internal/notification/sender"
)

func TestConsoleSend(t *testing.T) {
	var buf bytes.Buffer
	n := notification.Notification{Recipient: "test@example.com"}

	if err := sender.NewConsole(&buf).Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := buf.String(), "[Console] Sending notification to test@example.com\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
