package notification_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"notifyer/internal/notification"
	"notifyer/internal/notification/memory"
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

func newService(t *testing.T, snd *fakeSender) *notification.Service {
	t.Helper()
	svc, err := notification.NewService(
		memory.NewRepository(),
		map[string]notification.Sender{"console": snd},
	)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func TestServiceCreate(t *testing.T) {
	valid := notification.Notification{Recipient: "a@b.c", Channel: "console"}

	tests := []struct {
		name      string
		input     notification.Notification
		sendErr   error
		wantErr   error
		wantSends int
	}{
		{name: "valid", input: valid, wantSends: 1},
		{name: "empty recipient", input: notification.Notification{Channel: "console"}, wantErr: notification.ErrInvalid},
		{name: "unknown channel", input: notification.Notification{Recipient: "a@b.c", Channel: "fax"}, wantErr: notification.ErrUnknownChannel},
		{name: "send fails", input: valid, sendErr: errors.New("smtp down"), wantErr: notification.ErrSendFailed, wantSends: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			snd := &fakeSender{err: tt.sendErr}
			svc := newService(t, snd)

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
	svc := newService(t, &fakeSender{})

	_, err := svc.Get(context.Background(), 42)
	if !errors.Is(err, notification.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestServiceUpdateValidates(t *testing.T) {
	svc := newService(t, &fakeSender{})
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
	svc := newService(t, &fakeSender{})
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
