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
		return fmt.Errorf("%w: recipient cannot be empty", ErrInvalid)
	}
	if n.Channel == "" {
		return fmt.Errorf("%w: channel cannot be empty", ErrInvalid)
	}
	return nil
}
