package sender

import (
	"context"
	"fmt"
	"io"
	"os"

	"notifyer/internal/notification"
)

// Telegram is a stub: it writes to a writer instead of calling a real provider.
type Telegram struct {
	w io.Writer
}

// NewTelegram returns a Telegram writing to w, or os.Stdout if w is nil.
func NewTelegram(w io.Writer) *Telegram {
	if w == nil {
		w = os.Stdout
	}
	return &Telegram{w: w}
}

func (s *Telegram) Send(_ context.Context, n notification.Notification) error {
	_, err := fmt.Fprintf(s.w, "[Telegram] Sending message to %s\n", n.Recipient)
	return err
}
