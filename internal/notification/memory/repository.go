package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"notifyer/internal/notification"
)

type Repository struct {
	mu     sync.RWMutex
	items  map[int]notification.Notification
	nextID int
}

func NewRepository() *Repository {
	return &Repository{items: make(map[int]notification.Notification)}
}

var _ notification.Repository = (*Repository)(nil)

func (r *Repository) Save(_ context.Context, n notification.Notification) (notification.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	n.ID = r.nextID
	r.items[n.ID] = n
	return n, nil
}

func (r *Repository) List(_ context.Context) ([]notification.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	all := make([]notification.Notification, 0, len(r.items))
	for _, n := range r.items {
		all = append(all, n)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].ID < all[j].ID })
	return all, nil
}

func (r *Repository) Get(_ context.Context, id int) (notification.Notification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	n, ok := r.items[id]
	if !ok {
		return notification.Notification{}, fmt.Errorf("memory.Get: %w", notification.ErrNotFound)
	}
	return n, nil
}

func (r *Repository) Update(_ context.Context, n notification.Notification) (notification.Notification, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[n.ID]; !ok {
		return notification.Notification{}, fmt.Errorf("memory.Update: %w", notification.ErrNotFound)
	}
	r.items[n.ID] = n
	return n, nil
}
