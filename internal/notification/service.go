package notification

import (
	"context"
	"errors"
	"fmt"
	"io"
)

type Repository interface {
	Save(ctx context.Context, n Notification) (Notification, error)
	List(ctx context.Context) ([]Notification, error)
	Get(ctx context.Context, id int) (Notification, error)
	Update(ctx context.Context, n Notification) (Notification, error)
}

type Sender interface {
	Send(ctx context.Context, n Notification) error
}

type Service struct {
	repo    Repository
	senders map[string]Sender
}

// NewService creates a Service. senders maps channel name to Sender.
func NewService(repo Repository, senders map[string]Sender) (*Service, error) {
	if repo == nil {
		return nil, errors.New("notification.NewService: repository is nil")
	}
	if len(senders) == 0 {
		return nil, errors.New("notification.NewService: no senders")
	}
	return &Service{repo: repo, senders: senders}, nil
}

// senderFor validates n and returns the Sender for its channel.
func (s *Service) senderFor(n Notification) (Sender, error) {
	if err := n.Validate(); err != nil {
		return nil, err
	}
	sender, ok := s.senders[n.Channel]
	if !ok {
		return nil, ErrUnknownChannel.WithDetail("%q", n.Channel)
	}
	return sender, nil
}

// === Services ===
func (s *Service) Create(ctx context.Context, n Notification) (Notification, error) {
	const op = "notification.Service.Create"

	sender, err := s.senderFor(n)
	if err != nil {
		return Notification{}, err
	}

	saved, err := s.repo.Save(ctx, n)
	if err != nil {
		return Notification{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := sender.Send(ctx, saved); err != nil {
		return saved, fmt.Errorf("%s: %w", op, ErrSendFailed.Wrap(err))
	}
	return saved, nil
}

func (s *Service) List(ctx context.Context) ([]Notification, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("notification.Service.List: %w", err)
	}
	return all, nil
}

func (s *Service) Get(ctx context.Context, id int) (Notification, error) {
	n, err := s.repo.Get(ctx, id)
	if err != nil {
		return Notification{}, fmt.Errorf("notification.Service.Get: %w", err)
	}
	return n, nil
}

func (s *Service) Update(ctx context.Context, n Notification) (Notification, error) {
	const op = "notification.Service.Update"

	if _, err := s.senderFor(n); err != nil {
		return Notification{}, err
	}
	updated, err := s.repo.Update(ctx, n)
	if err != nil {
		return Notification{}, fmt.Errorf("%s: %w", op, err)
	}
	return updated, nil
}

func (s *Service) Export(ctx context.Context, w io.Writer) (int, error) {
	const op = "notification.Service.Export"

	all, err := s.repo.List(ctx)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	for _, n := range all {
		if _, err := fmt.Fprintf(w, "id:%d recipient:%s channel:%s\n", n.ID, n.Recipient, n.Channel); err != nil {
			return 0, fmt.Errorf("%s: %w", op, err)
		}
	}
	return len(all), nil
}
