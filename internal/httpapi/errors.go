package httpapi

import (
	"errors"
	"net/http"

	"notifyer/internal/apperr"
)

// httpStatus maps error kinds to HTTP status codes. apperr.Internal is absent on purpose.
var httpStatus = map[apperr.Kind]int{
	apperr.NotFound:     http.StatusNotFound,
	apperr.Invalid:      http.StatusBadRequest,
	apperr.Conflict:     http.StatusConflict,
	apperr.Unauthorized: http.StatusUnauthorized,
	apperr.Forbidden:    http.StatusForbidden,
	apperr.Unavailable:  http.StatusServiceUnavailable,
}

// writeServiceError maps an error from a domain service to an HTTP response.
// Known kinds expose only Code and Message; everything else becomes a generic 500.
// 5xx errors are logged with the full chain.
func (s *api) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		if status, ok := httpStatus[appErr.Kind]; ok {
			if status >= http.StatusInternalServerError {
				s.Logger.ErrorContext(r.Context(), "Request failed", "code", appErr.Code, "error", err)
			}
			writeJSON(w, status, errorResponse{Error: appErr.Message, Code: appErr.Code})
			return
		}
	}
	s.Logger.ErrorContext(r.Context(), "Request failed", "error", err)
	writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error", Code: "internal"})
}
