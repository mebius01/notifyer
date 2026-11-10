package notification

import "context"

// Repository persists notifications.
type Repository interface {
	Save(ctx context.Context, n Notification) (Notification, error)
	List(ctx context.Context) ([]Notification, error)
	Get(ctx context.Context, id int) (Notification, error)
	Update(ctx context.Context, n Notification) (Notification, error)
}

// Sender delivers a notification over a single channel.
type Sender interface {
	Send(ctx context.Context, n Notification) error
}
