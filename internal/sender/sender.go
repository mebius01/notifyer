package sender

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"order-platform/internal/notification"
)

// Sender sends a notification to its recipient over a particular channel.
type Sender interface {
	Send(n notification.Notification) error
}

// === ConsoleSender ===
type ConsoleSender struct {
	w io.Writer
}

func (cs ConsoleSender) Send(n notification.Notification) error {
	_, err := fmt.Fprintf(cs.w, "[ConsoleSender] Sending notification to %s \n", n.Recipient)
	return err
}

func NewConsoleSender(w io.Writer) *ConsoleSender {
	if w == nil {
		w = os.Stdout
	}

	return &ConsoleSender{w: w}
}

// === EmailSender ===
type EmailSender struct {
	w io.Writer
}

func (e EmailSender) Send(n notification.Notification) error {
	_, err := fmt.Fprintf(e.w, "[EmailSender] Sending email to %s \n", n.Recipient)
	return err
}

func NewEmailSender(w io.Writer) *EmailSender {
	if w == nil {
		w = os.Stdout
	}

	return &EmailSender{w: w}
}

// === TelegramSender ===
type TelegramSender struct{}

func (t TelegramSender) Send(n notification.Notification) error {
	fmt.Printf("[TelegramSender] .......>\n")
	return nil
}

// === LoggingSender ===
type LoggingSender struct {
	Sender
	Logger *slog.Logger
}

func (l LoggingSender) Send(n notification.Notification) error {
	l.Logger.Info("Sending to >>>> ", "recipient", n.Recipient)

	err := l.Sender.Send(n)
	if err != nil {
		l.Logger.Error("Failed to send notification", "error", err)
		return err
	}

	l.Logger.Info("Notification sent successfully", "recipient", n.Recipient)
	return nil
}
