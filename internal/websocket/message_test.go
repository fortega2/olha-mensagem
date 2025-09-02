package websocket_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/fortega2/real-time-chat/internal/websocket"
)

const (
	expectedContentErrMsg  = "Expected content %s, got %s"
	expectedTypeErrMsg     = "Expected type %s, got %s"
	expectedUsernameErrMsg = "Expected username %s, got %s"
	expectedColorErrMsg    = "Expected color %s, got %s"
	testNotification       = "test notification"
	channelID              = 1
)

func TestNewChatMessage(t *testing.T) {
	user := websocket.NewUser(1, "testuser")
	content := "Hello, world!"
	typeMsg := "Chat"

	message := websocket.NewChatMessage(user, typeMsg, content, channelID)

	if message.Type != typeMsg {
		t.Errorf(expectedTypeErrMsg, typeMsg, message.Type)
	}
	if message.UserID == nil {
		t.Fatal("Expected UserID to be non-nil")
	}
	if *message.UserID != user.ID {
		t.Errorf("Expected UserID %d, got %d", user.ID, *message.UserID)
	}
	if message.Username != user.Username {
		t.Errorf(expectedUsernameErrMsg, user.Username, message.Username)
	}
	if message.Content != content {
		t.Errorf(expectedContentErrMsg, content, message.Content)
	}
	if message.Color != user.Color {
		t.Errorf(expectedColorErrMsg, user.Color, message.Color)
	}
	if message.Timestamp == "" {
		t.Error("Expected timestamp to be non-empty")
	}

	_, err := time.Parse(time.RFC3339, message.Timestamp)
	if err != nil {
		t.Errorf("Invalid timestamp format: %v", err)
	}
}

func TestNewNotificationMessage(t *testing.T) {
	content := "User joined the chat"

	message := websocket.NewNotificationMessage(content, channelID)

	if message.Type != "Notification" {
		t.Errorf("Expected type 'Notification', got %s", message.Type)
	}
	if message.UserID != nil {
		t.Error("Expected UserID to be nil for notification message")
	}
	if message.Username != "" {
		t.Errorf("Expected empty username, got %s", message.Username)
	}
	if message.Content != content {
		t.Errorf(expectedContentErrMsg, content, message.Content)
	}
	if message.Color != "#666666" {
		t.Errorf("Expected color #666666, got %s", message.Color)
	}
	if message.Timestamp == "" {
		t.Error("Expected timestamp to be non-empty")
	}

	_, err := time.Parse(time.RFC3339, message.Timestamp)
	if err != nil {
		t.Errorf("Invalid timestamp format: %v", err)
	}
}

func TestMessageJSONMarshaling(t *testing.T) {
	user := websocket.NewUser(123, "jsonuser")
	message := websocket.NewChatMessage(user, "Chat", "test content", channelID)

	jsonData, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var unmarshaled websocket.Message
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if unmarshaled.Type != message.Type {
		t.Errorf(expectedTypeErrMsg, message.Type, unmarshaled.Type)
	}
	if unmarshaled.Content != message.Content {
		t.Errorf(expectedContentErrMsg, message.Content, unmarshaled.Content)
	}
	if unmarshaled.Username != message.Username {
		t.Errorf(expectedUsernameErrMsg, message.Username, unmarshaled.Username)
	}
	if unmarshaled.Color != message.Color {
		t.Errorf(expectedColorErrMsg, message.Color, unmarshaled.Color)
	}
}

func TestNotificationMessageJSONMarshaling(t *testing.T) {
	message := websocket.NewNotificationMessage(testNotification, channelID)

	jsonData, err := json.Marshal(message)
	if err != nil {
		t.Fatalf("Failed to marshal notification message: %v", err)
	}

	var unmarshaled websocket.Message
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal notification message: %v", err)
	}

	if unmarshaled.Type != "Notification" {
		t.Errorf(expectedTypeErrMsg, "Notification", unmarshaled.Type)
	}
	if unmarshaled.Content != testNotification {
		t.Errorf(expectedContentErrMsg, testNotification, unmarshaled.Content)
	}
	if unmarshaled.UserID != nil {
		t.Error("Expected UserID to be nil")
	}
}
