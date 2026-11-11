package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

// maxBodyBytes limits request body size.
const maxBodyBytes = 1 << 20 // 1 MB

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// errorResponse is the JSON body of every error response.
type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

// decodeJSON decodes a single JSON object into dst, rejecting unknown fields.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if dec.More() {
		return errors.New("invalid JSON: body must contain a single object")
	}
	return nil
}

// parseID reads {id} from the path; on failure it writes 400 and returns false.
func parseID(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}
