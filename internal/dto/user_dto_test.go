package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/fortega2/real-time-chat/internal/dto"
)

func TestNewUserDTO(t *testing.T) {
	id := int64(123)
	username := "testuser"

	userDTO := dto.NewUserDTO(id, username)

	if userDTO.ID != id {
		t.Errorf("Expected ID %d, got %d", id, userDTO.ID)
	}
	if userDTO.Username != username {
		t.Errorf("Expected username %s, got %s", username, userDTO.Username)
	}
}

func TestUserDTOJSONSerialization(t *testing.T) {
	userDTO := dto.NewUserDTO(456, "jsonuser")

	jsonData, err := json.Marshal(userDTO)
	if err != nil {
		t.Fatalf("Failed to marshal UserDTO: %v", err)
	}

	var unmarshaled dto.UserDTO
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal UserDTO: %v", err)
	}

	if unmarshaled.ID != userDTO.ID {
		t.Errorf("Expected ID %d, got %d", userDTO.ID, unmarshaled.ID)
	}
	if unmarshaled.Username != userDTO.Username {
		t.Errorf("Expected username %s, got %s", userDTO.Username, unmarshaled.Username)
	}
}

func TestUserDTOZeroValues(t *testing.T) {
	userDTO := dto.NewUserDTO(0, "")

	if userDTO.ID != 0 {
		t.Errorf("Expected ID 0, got %d", userDTO.ID)
	}
	if userDTO.Username != "" {
		t.Errorf("Expected empty username, got %s", userDTO.Username)
	}
}

func TestUserDTONegativeID(t *testing.T) {
	userDTO := dto.NewUserDTO(-1, "neguser")

	if userDTO.ID != -1 {
		t.Errorf("Expected ID -1, got %d", userDTO.ID)
	}
	if userDTO.Username != "neguser" {
		t.Errorf("Expected username 'neguser', got %s", userDTO.Username)
	}
}
