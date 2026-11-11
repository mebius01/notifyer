package sender_test

import (
	"bytes"
	"context"
	"testing"

	"notifyer/internal/notification"
	"notifyer/internal/notification/sender"
)

func TestSMSSend(t *testing.T) {
	var buf bytes.Buffer
	n := notification.Notification{Recipient: "+380501234567"}

	if err := sender.NewSMS(&buf).Send(context.Background(), n); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := buf.String(), "[SMS] Sending SMS to +380501234567\n"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
