package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"notifyer/internal/notification"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

var _ notification.Repository = (*Repository)(nil)

const columns = "id, recipient, subject, message, channel, is_urgent, status"

func scan(row pgx.Row) (notification.Notification, error) {
	var n notification.Notification
	err := row.Scan(&n.ID, &n.Recipient, &n.Subject, &n.Message, &n.Channel, &n.IsUrgent, &n.Status)
	return n, err
}

// === Repository methods ===
func (r *Repository) Save(ctx context.Context, n notification.Notification) (notification.Notification, error) {
	const op = "notification.postgres.Save"

	row := r.pool.QueryRow(ctx,
		`INSERT INTO notifications (recipient, subject, message, channel, is_urgent)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING `+columns,
		n.Recipient, n.Subject, n.Message, n.Channel, n.IsUrgent)

	saved, err := scan(row)
	if err != nil {
		return notification.Notification{}, fmt.Errorf("%s: %w", op, err)
	}
	return saved, nil
}

func (r *Repository) List(ctx context.Context) ([]notification.Notification, error) {
	const op = "notification.postgres.List"

	rows, err := r.pool.Query(ctx, `SELECT `+columns+` FROM notifications ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	all := []notification.Notification{}
	for rows.Next() {
		n, err := scan(rows)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		all = append(all, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return all, nil
}

func (r *Repository) Get(ctx context.Context, id int) (notification.Notification, error) {
	const op = "notification.postgres.Get"

	n, err := scan(r.pool.QueryRow(ctx,
		`SELECT `+columns+` FROM notifications WHERE id = $1`, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notification.Notification{}, fmt.Errorf("%s: %w", op, notification.ErrNotFound)
		}
		return notification.Notification{}, fmt.Errorf("%s: %w", op, err)
	}
	return n, nil
}

func (r *Repository) Update(ctx context.Context, n notification.Notification) (notification.Notification, error) {
	const op = "notification.postgres.Update"
	row := r.pool.QueryRow(ctx,
		`UPDATE notifications SET recipient = $1, subject = $2, message = $3, channel = $4, is_urgent = $5, updated_at = now() WHERE id = $6 RETURNING `+columns,
		n.Recipient, n.Subject, n.Message, n.Channel, n.IsUrgent, n.ID)

	updated, err := scan(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notification.Notification{}, fmt.Errorf("%s: %w", op, notification.ErrNotFound)
		}
		return notification.Notification{}, fmt.Errorf("%s: %w", op, err)
	}
	return updated, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id int, status notification.NotificationStatus) (notification.Notification, error) {
	const op = "notification.postgres.UpdateStatus"

	updated, err := scan(r.pool.QueryRow(ctx,
		`UPDATE notifications SET status = $1, updated_at = now() WHERE id = $2 RETURNING `+columns,
		status, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return notification.Notification{}, fmt.Errorf("%s: %w", op, notification.ErrNotFound)
		}
		return notification.Notification{}, fmt.Errorf("%s: %w", op, err)
	}
	return updated, nil
}
