package dto

import (
	"database/sql"
	"time"

	"github.com/fortega2/real-time-chat/internal/repository"
)

type CreateChannelRequestDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	UserID      int64  `json:"userId"`
}

func (ccr CreateChannelRequestDTO) IsValid() bool {
	return ccr.Name != "" && ccr.UserID != 0
}

type ChannelResponseDTO struct {
	ID                int64  `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	CreatedBy         int64  `json:"createdBy"`
	CreatedByUsername string `json:"createdByUsername"`
	CreatedAt         string `json:"createdAt"`
}

func NewChannelResponse[T repository.GetChannelByIDRow | repository.GetAllChannelsRow](channel T) ChannelResponseDTO {
	setDescription := func(description sql.NullString) string {
		if description.Valid {
			return description.String
		} else {
			return ""
		}
	}

	switch v := any(channel).(type) {
	case repository.GetChannelByIDRow:
		return ChannelResponseDTO{
			ID:                v.ID,
			Name:              v.Name,
			Description:       setDescription(v.Description),
			CreatedBy:         v.CreatedBy,
			CreatedByUsername: v.CreatedByUsername,
			CreatedAt:         v.CreatedAt.Format(time.RFC3339),
		}
	case repository.GetAllChannelsRow:
		return ChannelResponseDTO{
			ID:                v.ID,
			Name:              v.Name,
			Description:       setDescription(v.Description),
			CreatedBy:         v.CreatedBy,
			CreatedByUsername: v.CreatedByUsername,
			CreatedAt:         v.CreatedAt.Format(time.RFC3339),
		}
	default:
		return ChannelResponseDTO{}
	}
}

type DeleteChannelResponseDTO struct {
	Message   string `json:"message"`
	ChannelID int64  `json:"channelId"`
}
