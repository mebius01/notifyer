// Command notifyer runs the notification HTTP service.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"notifyer/internal/config"
	"notifyer/internal/httpapi"
	"notifyer/internal/notification"
	notifypg "notifyer/internal/notification/postgres"
	"notifyer/internal/notification/sender"
	"notifyer/internal/platform/postgres"
)

const AppVersion = "1.0.0"

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("Application stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// ctx is cancelled on SIGINT/SIGTERM for graceful shutdown.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// === Infrastructure ===
	pool, err := postgres.NewPool(ctx, postgres.Config{
		URL:      cfg.DB.URL,
		MaxConns: cfg.DB.MaxConns,
		MinConns: cfg.DB.MinConns,
	})
	if err != nil {
		return err
	}
	defer pool.Close()
	logger.Info("Connected to database")

	if err := postgres.Migrate(ctx, pool, logger); err != nil {
		return err
	}

	// === Features ===
	senders := map[string]notification.Sender{
		"console":  sender.NewLogging(sender.NewConsole(os.Stdout), logger),
		"email":    sender.NewLogging(sender.NewEmail(os.Stdout), logger),
		"telegram": sender.NewLogging(sender.NewTelegram(os.Stdout), logger),
		"sms":      sender.NewLogging(sender.NewSMS(os.Stdout), logger),
	}

	notifications, err := notification.NewService(notifypg.NewRepository(pool), senders)
	if err != nil {
		return err
	}

	// === HTTP ===
	handler, err := httpapi.NewHandler(httpapi.Deps{
		Notifications: notifications,
		Logger:        logger,
		Version:       AppVersion,
	})
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("Run server on", "address", cfg.HTTPAddr)
		errCh <- httpServer.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		logger.Info("Shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return httpServer.Shutdown(shutdownCtx)
}
