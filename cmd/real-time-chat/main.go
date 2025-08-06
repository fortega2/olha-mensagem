package main

import (
	"github.com/fortega2/real-time-chat/internal/server"
)

func main() {
	if err := server.Start(); err != nil {
		panic(err)
	}
}
