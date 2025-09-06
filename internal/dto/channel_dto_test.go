package dto_test

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/fortega2/real-time-chat/internal/dto"
	"github.com/fortega2/real-time-chat/internal/repository"
)

const (
	channelName = "Test Channel"
	description = "Test Description"
	userID      = 123
)

type ChannelResponseTestCase struct {
	name     string
	input    repository.GetChannelByIDRow
	expected dto.ChannelResponseDTO
}

func TestCreateChannelRequestDTOIsValid(t *testing.T) {
	tests := []struct {
		name     string
		request  dto.CreateChannelRequestDTO
		expected bool
	}{
		{
			name: "valid request",
			request: dto.CreateChannelRequestDTO{
				Name:        channelName,
				Description: description,
				UserID:      userID,
			},
			expected: true,
		},
		{
			name: "empty name",
			request: dto.CreateChannelRequestDTO{
				Name:        "",
				Description: description,
				UserID:      userID,
			},
			expected: false,
		},
		{
			name: "zero user ID",
			request: dto.CreateChannelRequestDTO{
				Name:        channelName,
				Description: description,
				UserID:      0,
			},
			expected: false,
		},
		{
			name: "empty name and zero user ID",
			request: dto.CreateChannelRequestDTO{
				Name:        "",
				Description: description,
				UserID:      0,
			},
			expected: false,
		},
		{
			name: "valid without description",
			request: dto.CreateChannelRequestDTO{
				Name:        channelName,
				Description: "",
				UserID:      userID,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.request.IsValid(); got != tt.expected {
				t.Errorf("IsValid() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewChannelResponseGetChannelByIDRow(t *testing.T) {
	tests := setupNewChannelResponseTestData()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dto.NewChannelResponse(tt.input)
			validateChannelResponse(t, result, tt)
		})
	}
}

func setupNewChannelResponseTestData() []ChannelResponseTestCase {
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []ChannelResponseTestCase{
		{
			name: "with description",
			input: repository.GetChannelByIDRow{
				ID:                1,
				Name:              channelName,
				Description:       sql.NullString{String: description, Valid: true},
				CreatedBy:         userID,
				CreatedByUsername: "testuser",
				CreatedAt:         testTime,
			},
			expected: dto.ChannelResponseDTO{
				ID:                1,
				Name:              channelName,
				Description:       description,
				CreatedBy:         userID,
				CreatedByUsername: "testuser",
				CreatedAt:         testTime.Format(time.RFC3339),
			},
		},
		{
			name: "without description",
			input: repository.GetChannelByIDRow{
				ID:                2,
				Name:              "Test Channel 2",
				Description:       sql.NullString{String: "", Valid: false},
				CreatedBy:         456,
				CreatedByUsername: "testuser2",
				CreatedAt:         testTime,
			},
			expected: dto.ChannelResponseDTO{
				ID:                2,
				Name:              "Test Channel 2",
				Description:       "",
				CreatedBy:         456,
				CreatedByUsername: "testuser2",
				CreatedAt:         testTime.Format(time.RFC3339),
			},
		},
	}
	return tests
}

func validateChannelResponse(t *testing.T, result dto.ChannelResponseDTO, tt ChannelResponseTestCase) {
	if result.ID != tt.expected.ID {
		t.Errorf("ID = %v, want %v", result.ID, tt.expected.ID)
	}
	if result.Name != tt.expected.Name {
		t.Errorf("Name = %v, want %v", result.Name, tt.expected.Name)
	}
	if result.Description != tt.expected.Description {
		t.Errorf("Description = %v, want %v", result.Description, tt.expected.Description)
	}
	if result.CreatedBy != tt.expected.CreatedBy {
		t.Errorf("CreatedBy = %v, want %v", result.CreatedBy, tt.expected.CreatedBy)
	}
	if result.CreatedByUsername != tt.expected.CreatedByUsername {
		t.Errorf("CreatedByUsername = %v, want %v", result.CreatedByUsername, tt.expected.CreatedByUsername)
	}
	if result.CreatedAt != tt.expected.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", result.CreatedAt, tt.expected.CreatedAt)
	}
}

func TestNewChannelResponseGetAllChannelsRow(t *testing.T) {
	testTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	input := repository.GetAllChannelsRow{
		ID:                1,
		Name:              channelName,
		Description:       sql.NullString{String: description, Valid: true},
		CreatedBy:         userID,
		CreatedByUsername: "testuser",
		CreatedAt:         testTime,
	}

	result := dto.NewChannelResponse(input)

	expected := dto.ChannelResponseDTO{
		ID:                1,
		Name:              channelName,
		Description:       description,
		CreatedBy:         userID,
		CreatedByUsername: "testuser",
		CreatedAt:         testTime.Format(time.RFC3339),
	}

	if result.ID != expected.ID {
		t.Errorf("ID = %v, want %v", result.ID, expected.ID)
	}
	if result.Name != expected.Name {
		t.Errorf("Name = %v, want %v", result.Name, expected.Name)
	}
	if result.Description != expected.Description {
		t.Errorf("Description = %v, want %v", result.Description, expected.Description)
	}
	if result.CreatedBy != expected.CreatedBy {
		t.Errorf("CreatedBy = %v, want %v", result.CreatedBy, expected.CreatedBy)
	}
	if result.CreatedByUsername != expected.CreatedByUsername {
		t.Errorf("CreatedByUsername = %v, want %v", result.CreatedByUsername, expected.CreatedByUsername)
	}
	if result.CreatedAt != expected.CreatedAt {
		t.Errorf("CreatedAt = %v, want %v", result.CreatedAt, expected.CreatedAt)
	}
}

func TestChannelResponseDTOJSONSerialization(t *testing.T) {
	channelDTO := dto.ChannelResponseDTO{
		ID:                1,
		Name:              channelName,
		Description:       description,
		CreatedBy:         userID,
		CreatedByUsername: "testuser",
		CreatedAt:         "2023-01-01T12:00:00Z",
	}

	jsonData, err := json.Marshal(channelDTO)
	if err != nil {
		t.Fatalf("Failed to marshal ChannelResponseDTO: %v", err)
	}

	var unmarshaled dto.ChannelResponseDTO
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal ChannelResponseDTO: %v", err)
	}

	if unmarshaled.ID != channelDTO.ID {
		t.Errorf("Expected ID %d, got %d", channelDTO.ID, unmarshaled.ID)
	}
	if unmarshaled.Name != channelDTO.Name {
		t.Errorf("Expected Name %s, got %s", channelDTO.Name, unmarshaled.Name)
	}
	if unmarshaled.Description != channelDTO.Description {
		t.Errorf("Expected Description %s, got %s", channelDTO.Description, unmarshaled.Description)
	}
}

func TestDeleteChannelResponseDTOJSONSerialization(t *testing.T) {
	deleteDTO := dto.DeleteChannelResponseDTO{
		Message:   "Channel deleted successfully",
		ChannelID: 1,
	}

	jsonData, err := json.Marshal(deleteDTO)
	if err != nil {
		t.Fatalf("Failed to marshal DeleteChannelResponseDTO: %v", err)
	}

	var unmarshaled dto.DeleteChannelResponseDTO
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal DeleteChannelResponseDTO: %v", err)
	}

	if unmarshaled.Message != deleteDTO.Message {
		t.Errorf("Expected Message %s, got %s", deleteDTO.Message, unmarshaled.Message)
	}
	if unmarshaled.ChannelID != deleteDTO.ChannelID {
		t.Errorf("Expected ChannelID %d, got %d", deleteDTO.ChannelID, unmarshaled.ChannelID)
	}
}
