package main

import (
	"fmt"
	"order-platform/internal/notification"
	"os"
)

func WriteAuditLog(path string, notifications map[int]notification.Notification) error {
	const op = "WriteAuditLog"

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	defer file.Close()

	for _, n := range notifications {
		_, err = fmt.Fprintf(file, "id:%d recipient:%s channel:%s\n", n.ID, n.Recipient, n.Channel)
		if err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	return nil
}
