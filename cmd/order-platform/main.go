package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"order-platform/internal/notification"
	"order-platform/internal/sender"
)

const (
	AppVersion = "1.0.0"
	AppName    = "Order Platform"
)

type Server struct {
	notifications map[int]notification.Notification
	nextID        int
	logger        *slog.Logger
	senders       map[string]sender.Sender
}

// === Repository ===
type NotificationRepository interface {
	Save(n notification.Notification) (notification.Notification, error)
	List() ([]notification.Notification, error)
	Get(id int) (notification.Notification, error)
}

func (s *Server) Save(n notification.Notification) (notification.Notification, error) {
	// Prepare the notification for sending & for saving in the server's notifications map
	s.nextID++
	n.ID = s.nextID
	s.notifications[n.ID] = n
	return n, nil
}

func (s *Server) List() ([]notification.Notification, error) {
	var all []notification.Notification = []notification.Notification{}
	for _, n := range s.notifications {
		all = append(all, n)
	}
	return all, nil
}

func (s *Server) Get(id int) (notification.Notification, error) {
	const op = "Server.Get"

	n, ok := s.notifications[id]
	if !ok {
		return notification.Notification{}, fmt.Errorf("%s: %w", op, notification.ErrNotFound)
	}
	return n, nil
}

// === Handlers/Services ===
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"status": "ok", "version": ` + AppVersion + `" }`))

}

func (s *Server) createNotification(w http.ResponseWriter, r *http.Request) {
	const op = "Server.createNotification"

	// Parse the request body into a notification.Notification struct
	var n notification.Notification
	err := json.NewDecoder(r.Body).Decode(&n)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{error: invalid JSON}`))
		return
	}

	validateErr := n.Validate()
	if validateErr != nil {
		validateErr = fmt.Errorf("%s: %w", op, validateErr)
		http.Error(w, validateErr.Error(), http.StatusBadRequest)
		return
	}

	// Create a concrete sender (e.g., ConsoleSender, EmailSender, TelegramSender) based on the channel
	// If the channel is unknown, return an error response
	sender, ok := s.senders[n.Channel]
	if !ok {
		s.logger.Error("No sender found for channel", "channel", n.Channel)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"{"error":"unknown channel"}"`))
		return
	}

	// Save the notification to the server's notifications map
	savedNotify, notifyErr := s.Save(n)
	if notifyErr != nil {
		s.logger.Error("Failed to save notification", "error", notifyErr)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{error: failed to save notification}`))
		return
	}

	// Send the notification using the selected sender
	sendErr := sender.Send(n)
	if sendErr != nil {
		s.logger.Error("Failed to send notification", "error", sendErr)
		return
	}

	js, err := json.Marshal(savedNotify)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{error: invalid JSON}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(js)
}

func (s *Server) listNotification(w http.ResponseWriter, r *http.Request) {

	all, err := s.List()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{error: failed to list notifications}`))
		return
	}

	js, err := json.Marshal(all)
	if err != nil {

		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{error: invalid JSON}`))
		return
	}

	w.Write(js)
}

func (s *Server) getNotification(w http.ResponseWriter, r *http.Request) {
	const op = "Server.getNotification"

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, fmt.Sprintf("%s: invalid ID", op), http.StatusBadRequest)
	}

	// Get the notification from Repository
	item, err := s.Get(id)

	if err != nil {
		if errors.Is(err, notification.ErrNotFound) {
			http.Error(w, fmt.Errorf("%s: %w", op, err).Error(), http.StatusNotFound)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(item)
}

func (s *Server) exportNotification(w http.ResponseWriter, r *http.Request) {
	const op = "Server.exportNotification"

	if err := WriteAuditLog("audit.log", s.notifications); err != nil {
		s.logger.Error("Failed to write audit log", "error", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"exported": len(s.notifications)})
}

// === Middleware ===
func (s *Server) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Info("Request complied", "method", r.Method, "path", r.URL.Path, "time", time.Since(start))
	})
}

func contentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func NewServer(logger *slog.Logger, senders map[string]sender.Sender) (*Server, error) {
	const op = "NewServer"
	if logger == nil {
		return nil, fmt.Errorf("%s: logger cannot be nil", op)
	}
	if len(senders) == 0 {
		return nil, fmt.Errorf("%s: senders cannot be empty", op)
	}

	return &Server{
		notifications: make(map[int]notification.Notification),
		logger:        logger,
		senders:       senders,
	}, nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	const dsn = "host=localhost port=5433 user=postgres password=postgres dbname=order_platform_db sslmode=disable"
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}

	logger.Info("Connected to database successfully")

	server, serverError := NewServer(logger, map[string]sender.Sender{
		"console":  sender.LoggingSender{Sender: sender.NewConsoleSender(os.Stdout), Logger: logger},
		"email":    sender.LoggingSender{Sender: sender.NewEmailSender(os.Stdout), Logger: logger},
		"telegram": sender.LoggingSender{Sender: sender.TelegramSender{}, Logger: logger},
	})

	if serverError != nil {
		logger.Error("Failed to create server", "error", serverError)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", server.healthHandler)
	mux.HandleFunc("GET /api/notifications/export", server.exportNotification)
	mux.HandleFunc("POST /api/notifications", server.createNotification)
	mux.HandleFunc("GET /api/notifications", server.listNotification)
	mux.HandleFunc("GET /api/notifications/{id}", server.getNotification)

	logger.Info("Run serer on", "address", "8080")
	http.ListenAndServe(":8080", server.logRequest(contentType(mux)))
}
