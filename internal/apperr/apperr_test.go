package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"notifyer/internal/apperr"
)

var errThing = &apperr.Error{Kind: apperr.NotFound, Code: "thing.not_found", Message: "thing not found"}

func TestIsMatchesByCode(t *testing.T) {
	other := &apperr.Error{Kind: apperr.NotFound, Code: "other.not_found", Message: "other"}
	cause := errors.New("db down")

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"same", errThing, true},
		{"wrapped with fmt", fmt.Errorf("op: %w", errThing), true},
		{"with detail", errThing.WithDetail("id %d", 5), true},
		{"with cause", errThing.Wrap(cause), true},
		{"other code", other, false},
		{"plain error", cause, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := errors.Is(tt.err, errThing); got != tt.want {
				t.Errorf("errors.Is = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrapKeepsCause(t *testing.T) {
	cause := errors.New("db down")
	err := errThing.Wrap(cause)

	if !errors.Is(err, cause) {
		t.Error("cause is lost")
	}
	if want := "thing not found: db down"; err.Error() != want {
		t.Errorf("Error() = %q, want %q", err.Error(), want)
	}
}

func TestWithDetailDoesNotMutateSentinel(t *testing.T) {
	_ = errThing.WithDetail("x")
	if errThing.Message != "thing not found" {
		t.Errorf("sentinel mutated: %q", errThing.Message)
	}
}

func TestKindOf(t *testing.T) {
	if got := apperr.KindOf(fmt.Errorf("op: %w", errThing)); got != apperr.NotFound {
		t.Errorf("KindOf = %v, want NotFound", got)
	}
	if got := apperr.KindOf(errors.New("boom")); got != apperr.Internal {
		t.Errorf("KindOf = %v, want Internal", got)
	}
}
