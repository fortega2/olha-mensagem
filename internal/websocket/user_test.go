package websocket_test

import (
	"testing"

	"github.com/fortega2/real-time-chat/internal/websocket"
)

func TestNewUser(t *testing.T) {
	user := websocket.NewUser(1, "testuser")

	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}

	if user.Color == "" {
		t.Error("Expected a random color, got an empty string")
	}

	if user.ID <= 0 {
		t.Error("Expected a positive ID, got", user.ID)
	}

	if user.JoinedAt.IsZero() {
		t.Error("Expected a non-zero JoinedAt time")
	}
}
