package websocket

import "time"

const (
	chatType         = "Chat"
	notificationType = "Notification"
)

type Message struct {
	Type      string `json:"type"`
	UserID    *int   `json:"userId,omitempty"`
	Username  string `json:"username,omitempty"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Color     string `json:"color"`
}

func NewChatMessage(user *User, typeMsg, content string) Message {
	return Message{
		Type:      typeMsg,
		UserID:    &user.ID,
		Username:  user.Username,
		Content:   content,
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     user.Color,
	}
}

func NewNotificationMessage(content string) Message {
	return Message{
		Type:      notificationType,
		Content:   content,
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     "#666666",
	}
}
