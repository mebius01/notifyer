package notification

import "fmt"

type Notification struct {
	ID        int
	Recipient string
	Subject   string
	Message   string
	Channel   string
	IsUrgent  bool
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
