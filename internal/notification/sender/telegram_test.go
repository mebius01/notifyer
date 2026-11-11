package sender_test

import (
	"bytes"
	"context"
	"testing"

	"notifyer/internal/notification"
	"notifyer/internal/notification/sender"
)

func TestTelegramSend(t *testing.T) {
	var buf bytes.Buffer
	n := notification.Notification{Recipient: "@user"}

	if err := sender.NewTelegram(&buf).Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := buf.String(), "[Telegram] Sending message to @user\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
