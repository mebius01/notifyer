package sender

import (
	"context"
	"log/slog"

	"notifyer/internal/notification"
)

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
