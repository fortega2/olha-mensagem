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

func TestNewUserMultipleUsers(t *testing.T) {
	user1 := websocket.NewUser(1, "user1")
	user2 := websocket.NewUser(2, "user2")

	if user1.ID == user2.ID {
		t.Error("Expected different IDs for different users")
	}

	if user1.Username == user2.Username {
		t.Error("Expected different usernames for different users")
	}
}

func TestNewUserWithZeroID(t *testing.T) {
	user := websocket.NewUser(0, "zerouser")

	if user.ID != 0 {
		t.Errorf("Expected ID 0, got %d", user.ID)
	}

	if user.Username != "zerouser" {
		t.Errorf("Expected username 'zerouser', got '%s'", user.Username)
	}

	if user.Color == "" {
		t.Error("Expected a color even with zero ID")
	}
}

func TestNewUserWithEmptyUsername(t *testing.T) {
	user := websocket.NewUser(123, "")

	if user.Username != "" {
		t.Errorf("Expected empty username, got '%s'", user.Username)
	}

	if user.ID != 123 {
		t.Errorf("Expected ID 123, got %d", user.ID)
	}

	if user.Color == "" {
		t.Error("Expected a color even with empty username")
	}
}

func TestUserColorIsValid(t *testing.T) {
	user := websocket.NewUser(1, "colortest")

	if len(user.Color) == 0 || user.Color[0] != '#' {
		t.Errorf("Expected color to start with '#', got '%s'", user.Color)
	}

	if len(user.Color) != 7 {
		t.Errorf("Expected color length 7, got %d for color '%s'", len(user.Color), user.Color)
	}
}
