package sender

import (
	"context"
	"fmt"
	"io"
	"os"

	"notifyer/internal/notification"
)

// SMS is a stub: it writes to a writer instead of calling a real provider.
type SMS struct {
	w io.Writer
}

// NewSMS returns a SMS writing to w, or os.Stdout if w is nil.
func NewSMS(w io.Writer) *SMS {
	if w == nil {
		w = os.Stdout
	}
	return &SMS{w: w}
}

func (s *SMS) Send(_ context.Context, n notification.Notification) error {
	_, err := fmt.Fprintf(s.w, "[SMS] Sending SMS to %s\n", n.Recipient)
	return err
}
