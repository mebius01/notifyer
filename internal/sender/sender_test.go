package sender

import (
	"bytes"
	"testing"

	"order-platform/internal/notification"
)

func TestConsoleSender(t *testing.T) {
	// Anything who wants Write (bytes.Buffer, os.Stdout, etc.)
	var buf bytes.Buffer
	// Create a ConsoleSender with the buffer
	cs := NewConsoleSender(&buf)

	n := notification.Notification{
		ID:        1,
		Recipient: "test@example.com",
		Subject:   "Test Notification",
		Message:   "This is a test notification.",
	}

	err := cs.Send(n)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Check the output in the buffer
	got := buf.String()
	// The expected output should match the format used in ConsoleSender.Send
	want := "[ConsoleSender] Sending notification to " + n.Recipient + " \n"
	if got != want {
		t.Errorf("Expected %q, got %q", want, got)
	}
}
