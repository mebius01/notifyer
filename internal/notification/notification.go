package notification

import "fmt"

type Notification struct {
	ID        int    `json:"id"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	Channel   string `json:"channel"`
	IsUrgent  bool   `json:"is_urgent"`
}

func (n Notification) String() string {
	return fmt.Sprintf("Notification[ID=%d, Recipient=%s, Channel=%s,]",
		n.ID, n.Recipient, n.Channel)
}

// === Validate ===

func (n Notification) Validate() error {
	const op = "Notification.Validate"
	if n.Recipient == "" {
		return fmt.Errorf("%s: recipient cannot be empty", op)
	}
	if n.Channel == "" {
		return fmt.Errorf("%s: channel cannot be empty", op)
	}

	return nil
}
