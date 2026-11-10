// Package sender implements notification.Sender for each delivery channel.
package sender

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"notifyer/internal/notification"
)

// Compile-time interface checks.
var (
	_ notification.Sender = (*Console)(nil)
	_ notification.Sender = (*Email)(nil)
	_ notification.Sender = (*Telegram)(nil)
	_ notification.Sender = (*Logging)(nil)
)

// writerOrStdout returns w, or os.Stdout if w is nil.
func writerOrStdout(w io.Writer) io.Writer {
	if w == nil {
		return os.Stdout
	}
	return w
}

// === Console ===

type Console struct {
	w io.Writer
}

func NewConsole(w io.Writer) *Console {
	return &Console{w: writerOrStdout(w)}
}

func (c *Console) Send(_ context.Context, n notification.Notification) error {
	_, err := fmt.Fprintf(c.w, "[Console] Sending notification to %s\n", n.Recipient)
	return err
}

// === Email ===

type Email struct {
	w io.Writer
}

func NewEmail(w io.Writer) *Email {
	return &Email{w: writerOrStdout(w)}
}

func (e *Email) Send(_ context.Context, n notification.Notification) error {
	_, err := fmt.Fprintf(e.w, "[Email] Sending email to %s\n", n.Recipient)
	return err
}

// === Telegram ===

type Telegram struct {
	w io.Writer
}

func NewTelegram(w io.Writer) *Telegram {
	return &Telegram{w: writerOrStdout(w)}
}

func (t *Telegram) Send(_ context.Context, n notification.Notification) error {
	_, err := fmt.Fprintf(t.w, "[Telegram] Sending message to %s\n", n.Recipient)
	return err
}

// === Logging ===

// Logging wraps a Sender and logs each delivery.
type Logging struct {
	next   notification.Sender
	logger *slog.Logger
}

func NewLogging(next notification.Sender, logger *slog.Logger) *Logging {
	return &Logging{next: next, logger: logger}
}

func (l *Logging) Send(ctx context.Context, n notification.Notification) error {
	l.logger.InfoContext(ctx, "Sending notification", "recipient", n.Recipient, "channel", n.Channel)

	if err := l.next.Send(ctx, n); err != nil {
		l.logger.ErrorContext(ctx, "Failed to send notification", "recipient", n.Recipient, "error", err)
		return err
	}

	l.logger.InfoContext(ctx, "Notification sent", "recipient", n.Recipient)
	return nil
}
