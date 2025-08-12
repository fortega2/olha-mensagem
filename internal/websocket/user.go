package websocket

import (
	"math/rand"
	"time"
)

type User struct {
	ID       int       `json:"id"`
	Username string    `json:"username"`
	Color    string    `json:"color"`
	JoinedAt time.Time `json:"joinedAt"`
}

var (
	users        = make(map[int]*User)
	randomColors = []string{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7",
		"#DDA0DD", "#98D8C8", "#F7DC6F", "#BB8FCE", "#85C1E9",
		"#F8C471", "#82E0AA", "#F1948A", "#85C1E9", "#D7DBDD",
	}
)

func NewUser(username string) *User {
	return &User{
		ID:       generateRandomID(),
		Username: username,
		Color:    generateRandomColor(),
		JoinedAt: time.Now(),
	}
}

func AddUser(user *User) {
	users[user.ID] = user
}

func GetUserByID(id int) *User {
	if user, exists := users[id]; exists {
		return user
	}
	return nil
}

func generateRandomID() int {
	return int(time.Now().UnixNano() / int64(time.Millisecond))
}

func generateRandomColor() string {
	return randomColors[rand.Intn(len(randomColors))]
}
