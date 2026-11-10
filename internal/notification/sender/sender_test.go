package sender_test

import (
	"bytes"
	"context"
	"testing"

	"notifyer/internal/notification"
	"notifyer/internal/notification/sender"
)

func TestSenders(t *testing.T) {
	n := notification.Notification{ID: 1, Recipient: "test@example.com"}

	tests := []struct {
		name string
		new  func(*bytes.Buffer) notification.Sender
		want string
	}{
		{"console", func(b *bytes.Buffer) notification.Sender { return sender.NewConsole(b) }, "[Console] Sending notification to test@example.com\n"},
		{"email", func(b *bytes.Buffer) notification.Sender { return sender.NewEmail(b) }, "[Email] Sending email to test@example.com\n"},
		{"telegram", func(b *bytes.Buffer) notification.Sender { return sender.NewTelegram(b) }, "[Telegram] Sending message to test@example.com\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := tt.new(&buf).Send(context.Background(), n); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
