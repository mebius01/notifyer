package sender

import (
	"context"
	"fmt"
	"io"
	"os"

	"notifyer/internal/notification"
)

// Email is a stub: it writes to a writer instead of calling a real provider.
type Email struct {
	w io.Writer
}

// NewEmail returns a Email writing to w, or os.Stdout if w is nil.
func NewEmail(w io.Writer) *Email {
	if w == nil {
		w = os.Stdout
	}
	return &Email{w: w}
}

func (s *Email) Send(_ context.Context, n notification.Notification) error {
	_, err := fmt.Fprintf(s.w, "[Email] Sending email to %s\n", n.Recipient)
	return err
}
