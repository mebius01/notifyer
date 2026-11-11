package sender

import (
	"context"
	"fmt"
	"io"
	"os"

	"notifyer/internal/notification"
)

// Console prints notifications to a writer. For development only.
type Console struct {
	w io.Writer
}

// NewConsole returns a Console writing to w, or os.Stdout if w is nil.
func NewConsole(w io.Writer) *Console {
	if w == nil {
		w = os.Stdout
	}
	return &Console{w: w}
}

func (c *Console) Send(_ context.Context, n notification.Notification) error {
	_, err := fmt.Fprintf(c.w, "[Console] Sending notification to %s\n", n.Recipient)
	return err
}
