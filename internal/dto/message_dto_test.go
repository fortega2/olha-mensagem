package dto_test

import (
	"testing"
	"time"

	"github.com/fortega2/real-time-chat/internal/dto"
	"github.com/fortega2/real-time-chat/internal/repository"
)

func assertMessageDTOEquals(t *testing.T, expected, actual dto.MessageDTO) {
	if actual.ID != expected.ID {
		t.Errorf("Expected ID %d, got %d", expected.ID, actual.ID)
	}

	if actual.ChannelID != expected.ChannelID {
		t.Errorf("Expected ChannelID %d, got %d", expected.ChannelID, actual.ChannelID)
	}

	if actual.UserID != expected.UserID {
		t.Errorf("Expected UserID %d, got %d", expected.UserID, actual.UserID)
	}

	if actual.UserUsername != expected.UserUsername {
		t.Errorf("Expected UserUsername %s, got %s", expected.UserUsername, actual.UserUsername)
	}

	if actual.Content != expected.Content {
		t.Errorf("Expected Content %s, got %s", expected.Content, actual.Content)
	}

	if actual.CreatedAt != expected.CreatedAt {
		t.Errorf("Expected CreatedAt %s, got %s", expected.CreatedAt, actual.CreatedAt)
	}
}

func TestNewMessageDTO(t *testing.T) {
	testCases := []struct {
		name        string
		repoMessage repository.GetHistoryMessagesByChannelRow
		expectedDTO dto.MessageDTO
	}{
		{
			name: "Valid message conversion",
			repoMessage: repository.GetHistoryMessagesByChannelRow{
				ID:           1,
				ChannelID:    10,
				UserID:       5,
				UserUsername: "testuser",
				Content:      "Hello, world!",
				CreatedAt:    time.Date(2023, 12, 25, 10, 30, 0, 0, time.UTC),
			},
			expectedDTO: dto.MessageDTO{
				ID:           1,
				ChannelID:    10,
				UserID:       5,
				UserUsername: "testuser",
				Content:      "Hello, world!",
				CreatedAt:    "2023-12-25T10:30:00Z",
			},
		},
		{
			name: "Empty content message",
			repoMessage: repository.GetHistoryMessagesByChannelRow{
				ID:           2,
				ChannelID:    20,
				UserID:       3,
				UserUsername: "emptyuser",
				Content:      "",
				CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			expectedDTO: dto.MessageDTO{
				ID:           2,
				ChannelID:    20,
				UserID:       3,
				UserUsername: "emptyuser",
				Content:      "",
				CreatedAt:    "2024-01-01T00:00:00Z",
			},
		},
		{
			name: "Special characters in content",
			repoMessage: repository.GetHistoryMessagesByChannelRow{
				ID:           3,
				ChannelID:    15,
				UserID:       7,
				UserUsername: "specialuser",
				Content:      "Special chars: @#$%^&*()_+{}|:<>?[]\\;'\",./ áéíóú",
				CreatedAt:    time.Date(2024, 6, 15, 14, 45, 30, 123456789, time.UTC),
			},
			expectedDTO: dto.MessageDTO{
				ID:           3,
				ChannelID:    15,
				UserID:       7,
				UserUsername: "specialuser",
				Content:      "Special chars: @#$%^&*()_+{}|:<>?[]\\;'\",./ áéíóú",
				CreatedAt:    "2024-06-15T14:45:30Z",
			},
		},
		{
			name: "Zero values",
			repoMessage: repository.GetHistoryMessagesByChannelRow{
				ID:           0,
				ChannelID:    0,
				UserID:       0,
				UserUsername: "",
				Content:      "",
				CreatedAt:    time.Time{},
			},
			expectedDTO: dto.MessageDTO{
				ID:           0,
				ChannelID:    0,
				UserID:       0,
				UserUsername: "",
				Content:      "",
				CreatedAt:    "0001-01-01T00:00:00Z",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := dto.NewMessageDTO(tc.repoMessage)
			assertMessageDTOEquals(t, tc.expectedDTO, result)
		})
	}
}

func TestNewMessageDTOTimeFormat(t *testing.T) {
	testTime := time.Date(2024, 3, 15, 9, 45, 30, 500000000, time.UTC)

	repoMessage := repository.GetHistoryMessagesByChannelRow{
		ID:           1,
		ChannelID:    1,
		UserID:       1,
		UserUsername: "testuser",
		Content:      "test message",
		CreatedAt:    testTime,
	}

	result := dto.NewMessageDTO(repoMessage)

	expectedTimeStr := "2024-03-15T09:45:30Z"
	if result.CreatedAt != expectedTimeStr {
		t.Errorf("Expected time format %s, got %s", expectedTimeStr, result.CreatedAt)
	}

	_, err := time.Parse(time.RFC3339, result.CreatedAt)
	if err != nil {
		t.Errorf("Failed to parse formatted time: %v", err)
	}
}
