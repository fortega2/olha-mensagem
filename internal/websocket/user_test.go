package websocket_test

import (
	"testing"

	"github.com/fortega2/real-time-chat/internal/websocket"
)

const (
	fakeUserID = -1
)

func TestNewUser(t *testing.T) {
	user := websocket.NewUser("testuser")

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

func TestAddUserFound(t *testing.T) {
	user := websocket.NewUser("testuser")
	websocket.AddUser(user)
	if retrivedUser := websocket.GetUserByID(user.ID); retrivedUser == nil {
		t.Errorf("Expected user with ID %d to be added, but it was not found", user.ID)
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	user := websocket.GetUserByID(fakeUserID)
	if user != nil {
		t.Error("Expected nil for non-existing user, got", user)
	}
}
