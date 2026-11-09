package notification

import "testing"

func TestNotifyValidate(t *testing.T) {
	test := []struct {
		name         string
		notification Notification
		wantErr      bool
	}{
		{
			name: "Valid Notification",
			notification: Notification{
				Recipient: "test@example.com",
				Subject:   "Test Notification",
				Message:   "This is a test notification.",
				Channel:   "email",
				IsUrgent:  false,
			},
			wantErr: false,
		},
		{
			name: "Missing Recipient",
			notification: Notification{
				Recipient: "",
				Subject:   "Test Notification",
				Message:   "This is a test notification.",
				Channel:   "email",
				IsUrgent:  false,
			},
			wantErr: true,
		},
		{
			name: "Missing Channel",
			notification: Notification{
				Recipient: "test@gmail.com",
				Subject:   "Test Notification",
				Message:   "This is a test notification.",
				Channel:   "",
				IsUrgent:  false,
			},
			wantErr: true,
		},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.notification.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("Expected error: %v, got: %v", tt.wantErr, err)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected error got %v", err)
			}

		})
	}
}
