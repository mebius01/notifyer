package notification

import "notifyer/internal/apperr"

// Domain errors. Transports map their Kind to status codes.
var (
	ErrNotFound       = &apperr.Error{Kind: apperr.NotFound, Code: "notification.not_found", Message: "notification not found"}
	ErrInvalid        = &apperr.Error{Kind: apperr.Invalid, Code: "notification.invalid", Message: "invalid notification"}
	ErrUnknownChannel = &apperr.Error{Kind: apperr.Invalid, Code: "notification.unknown_channel", Message: "unknown channel"}
	ErrSendFailed     = &apperr.Error{Kind: apperr.Unavailable, Code: "notification.send_failed", Message: "failed to send notification"}
)
