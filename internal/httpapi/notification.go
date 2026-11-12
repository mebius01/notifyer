package httpapi

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"

	"notifyer/internal/notification"
)

// === DTOs ===
type notificationRequest struct {
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	Channel   string `json:"channel"`
	IsUrgent  bool   `json:"is_urgent"`
}

func (req notificationRequest) toDomain() notification.Notification {
	return notification.Notification{
		Recipient: req.Recipient,
		Subject:   req.Subject,
		Message:   req.Message,
		Channel:   req.Channel,
		IsUrgent:  req.IsUrgent,
	}
}

type notificationResponse struct {
	ID        int    `json:"id"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	Channel   string `json:"channel"`
	IsUrgent  bool   `json:"is_urgent"`
	Status    string `json:"status"`
}

func toNotificationResponse(n notification.Notification) notificationResponse {
	return notificationResponse{
		ID:        n.ID,
		Recipient: n.Recipient,
		Subject:   n.Subject,
		Message:   n.Message,
		Channel:   n.Channel,
		IsUrgent:  n.IsUrgent,
		Status:    string(n.Status),
	}
}

// === Routs ===
func (s *api) registerNotificationRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/notifications/export", s.exportNotifications)
	mux.HandleFunc("POST /api/notifications", s.createNotification)
	mux.HandleFunc("GET /api/notifications", s.listNotifications)
	mux.HandleFunc("GET /api/notifications/{id}", s.getNotification)
	mux.HandleFunc("PATCH /api/notifications/{id}", s.updateNotification)
}

// === Handlers ===
func (s *api) createNotification(w http.ResponseWriter, r *http.Request) {
	var req notificationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	saved, err := s.Notifications.Create(r.Context(), req.toDomain())
	if errors.Is(err, notification.ErrSendFailed) {
		// Saved but not delivered: return 502 with the id of the stored record.
		s.Logger.ErrorContext(r.Context(), "Notification saved but not sent", "id", saved.ID, "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"error": "notification saved but sending failed",
			"code":  notification.ErrSendFailed.Code,
			"id":    saved.ID,
		})
		return
	}
	if err != nil {
		s.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusAccepted, toNotificationResponse(saved))
}

func (s *api) listNotifications(w http.ResponseWriter, r *http.Request) {
	all, err := s.Notifications.List(r.Context())
	if err != nil {
		s.writeServiceError(w, r, err)
		return
	}
	resp := make([]notificationResponse, 0, len(all))
	for _, n := range all {
		resp = append(resp, toNotificationResponse(n))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *api) getNotification(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	n, err := s.Notifications.Get(r.Context(), id)
	if err != nil {
		s.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toNotificationResponse(n))
}

func (s *api) updateNotification(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var req notificationRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	n := req.toDomain()
	n.ID = id // id comes from the path, not the body

	updated, err := s.Notifications.Update(r.Context(), n)
	if err != nil {
		s.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toNotificationResponse(updated))
}

func (s *api) exportNotifications(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	count, err := s.Notifications.Export(r.Context(), &buf)
	if err != nil {
		s.writeServiceError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="notifications.txt"`)
	w.Header().Set("X-Export-Count", strconv.Itoa(count))
	w.WriteHeader(http.StatusOK)
	_, _ = buf.WriteTo(w)
}
