package websocket

import "time"

const (
	chatType = "Chat"
)

type Message struct {
	Type      string `json:"type"`
	UserID    int    `json:"userId"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
	Color     string `json:"color"`
}

func newMessage(user *User, typeMsg, content string) Message {
	return Message{
		Type:      typeMsg,
		UserID:    user.ID,
		Username:  user.Username,
		Content:   content,
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     user.Color,
	}
}
