package notification_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"notifyer/internal/notification"
)

// fakeSender records sent notifications and returns err.
type fakeSender struct {
	sent []notification.Notification
	err  error
}

func (f *fakeSender) Send(_ context.Context, n notification.Notification) error {
	f.sent = append(f.sent, n)
	return f.err
}

// fakeRepo is a minimal Repository: IDs start at 1 and equal index+1.
// If err is set, every method returns it.
type fakeRepo struct {
	items []notification.Notification
	err   error
}

var _ notification.Repository = (*fakeRepo)(nil)

func (r *fakeRepo) Save(_ context.Context, n notification.Notification) (notification.Notification, error) {
	if r.err != nil {
		return notification.Notification{}, r.err
	}
	n.ID = len(r.items) + 1
	r.items = append(r.items, n)
	return n, nil
}

func (r *fakeRepo) List(context.Context) ([]notification.Notification, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.items, nil
}

func (r *fakeRepo) Get(_ context.Context, id int) (notification.Notification, error) {
	if r.err != nil {
		return notification.Notification{}, r.err
	}
	if id < 1 || id > len(r.items) {
		return notification.Notification{}, notification.ErrNotFound
	}
	return r.items[id-1], nil
}

func (r *fakeRepo) Update(_ context.Context, n notification.Notification) (notification.Notification, error) {
	if r.err != nil {
		return notification.Notification{}, r.err
	}
	if n.ID < 1 || n.ID > len(r.items) {
		return notification.Notification{}, notification.ErrNotFound
	}
	r.items[n.ID-1] = n
	return n, nil
}

func newService(t *testing.T, repo *fakeRepo, snd *fakeSender) *notification.Service {
	t.Helper()
	svc, err := notification.NewService(
		repo,
		map[string]notification.Sender{"console": snd},
	)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

var errDB = errors.New("db down")

func TestServiceCreate(t *testing.T) {
	valid := notification.Notification{Recipient: "a@b.c", Channel: "console"}

	tests := []struct {
		name      string
		input     notification.Notification
		repoErr   error
		sendErr   error
		wantErr   error
		wantSends int
	}{
		{name: "valid", input: valid, wantSends: 1},
		{name: "empty recipient", input: notification.Notification{Channel: "console"}, wantErr: notification.ErrInvalid},
		{name: "unknown channel", input: notification.Notification{Recipient: "a@b.c", Channel: "fax"}, wantErr: notification.ErrUnknownChannel},
		{name: "repo fails", input: valid, repoErr: errDB, wantErr: errDB},
		{name: "send fails", input: valid, sendErr: errors.New("smtp down"), wantErr: notification.ErrSendFailed, wantSends: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snd := &fakeSender{err: tt.sendErr}
			svc := newService(t, &fakeRepo{err: tt.repoErr}, snd)

			got, err := svc.Create(context.Background(), tt.input)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if len(snd.sent) != tt.wantSends {
				t.Errorf("sends = %d, want %d", len(snd.sent), tt.wantSends)
			}
			if tt.wantSends > 0 && got.ID == 0 {
				t.Error("expected saved notification to have an ID")
			}
		})
	}
}

func TestServiceGetNotFound(t *testing.T) {
	svc := newService(t, &fakeRepo{}, &fakeSender{})

	_, err := svc.Get(context.Background(), 42)
	if !errors.Is(err, notification.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestServiceUpdateValidates(t *testing.T) {
	svc := newService(t, &fakeRepo{}, &fakeSender{})
	ctx := context.Background()

	saved, err := svc.Create(ctx, notification.Notification{Recipient: "a@b.c", Channel: "console"})
	if err != nil {
		t.Fatal(err)
	}

	saved.Recipient = ""
	if _, err := svc.Update(ctx, saved); !errors.Is(err, notification.ErrInvalid) {
		t.Fatalf("err = %v, want ErrInvalid", err)
	}
}

func TestServiceExport(t *testing.T) {
	svc := newService(t, &fakeRepo{}, &fakeSender{})
	ctx := context.Background()

	for range 2 {
		if _, err := svc.Create(ctx, notification.Notification{Recipient: "a@b.c", Channel: "console"}); err != nil {
			t.Fatal(err)
		}
	}

	var buf bytes.Buffer
	count, err := svc.Export(ctx, &buf)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("exported = %d, want 2", count)
	}
	want := "id:1 recipient:a@b.c channel:console\nid:2 recipient:a@b.c channel:console\n"
	if got := buf.String(); got != want {
		t.Errorf("export output = %q, want %q", got, want)
	}
}
