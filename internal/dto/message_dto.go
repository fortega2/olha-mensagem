package dto

import (
	"time"

	"github.com/fortega2/real-time-chat/internal/repository"
)

type MessageDTO struct {
	ID           int64  `json:"id"`
	ChannelID    int64  `json:"channelId"`
	UserID       int64  `json:"userId"`
	UserUsername string `json:"userUsername"`
	Content      string `json:"content"`
	Timestamp    string `json:"timestamp"`
}

func NewMessageDTO(repoMessage repository.GetHistoryMessagesByChannelRow) MessageDTO {
	return MessageDTO{
		ID:           repoMessage.ID,
		ChannelID:    repoMessage.ChannelID,
		UserID:       repoMessage.UserID,
		UserUsername: repoMessage.UserUsername,
		Content:      repoMessage.Content,
		Timestamp:    repoMessage.CreatedAt.Format(time.RFC3339),
	}
}
