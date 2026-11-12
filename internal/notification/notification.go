package notification

import "fmt"

type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "Pending"
	NotificationStatusSent    NotificationStatus = "Sent"
	NotificationStatusFail    NotificationStatus = "Fail"
)

func (s NotificationStatus) Valid() bool {
	switch s {
	case NotificationStatusPending, NotificationStatusSent, NotificationStatusFail:
		return true
	}
	return false
}

type Notification struct {
	ID        int
	Recipient string
	Subject   string
	Message   string
	Channel   string
	IsUrgent  bool
	Status    NotificationStatus
}

func (n Notification) String() string {
	return fmt.Sprintf("Notification[ID=%d, Recipient=%s, Channel=%s]",
		n.ID, n.Recipient, n.Channel)
}

// Validate checks required fields.
func (n Notification) Validate() error {
	if n.Recipient == "" {
		return ErrInvalid.WithDetail("recipient cannot be empty")
	}
	if n.Channel == "" {
		return ErrInvalid.WithDetail("channel cannot be empty")
	}
	return nil
}
