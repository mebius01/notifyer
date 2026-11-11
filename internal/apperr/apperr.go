// Package apperr defines transport-agnostic application errors.
// Domains declare their errors with a Kind; transports (HTTP, gRPC) map Kind to their codes.
// This package must not import any other project package.
package apperr

import (
	"errors"
	"fmt"
)

// Kind is the category of an error. Transports map it to status codes.
type Kind int

const (
	Internal Kind = iota // unknown or unexpected error; details are never shown to clients
	NotFound
	Invalid
	Conflict
	Unauthorized
	Forbidden
	Unavailable
)

// Error is an application error.
type Error struct {
	Kind    Kind
	Code    string // stable machine-readable code, e.g. "notification.not_found"
	Message string // safe to show to clients
	Err     error  // internal cause; logged, never shown to clients
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

// Is matches errors by Code, so errors.Is(err, ErrX) holds for copies made by WithDetail and Wrap.
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	return ok && e.Code != "" && e.Code == t.Code
}

// WithDetail returns a copy of e with a client-safe detail appended to Message.
func (e *Error) WithDetail(format string, args ...any) *Error {
	c := *e
	c.Message = e.Message + ": " + fmt.Sprintf(format, args...)
	return &c
}

// Wrap returns a copy of e with an internal cause attached.
func (e *Error) Wrap(cause error) *Error {
	c := *e
	c.Err = cause
	return &c
}

// KindOf returns the Kind of the first *Error in err's chain, or Internal.
func KindOf(err error) Kind {
	var e *Error
	if errors.As(err, &e) {
		return e.Kind
	}
	return Internal
}
