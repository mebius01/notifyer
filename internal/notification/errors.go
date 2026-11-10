package notification

import "errors"

// Domain errors. Transport layers map them to status codes.
var (
	ErrNotFound       = errors.New("notification not found")
	ErrInvalid        = errors.New("invalid notification")
	ErrUnknownChannel = errors.New("unknown channel")
	ErrSendFailed     = errors.New("failed to send notification")
)
