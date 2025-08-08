# Real-Time Chat Application

A simple real-time chat application built with Go using WebSockets. This project demonstrates WebSocket implementation for real-time communication between multiple clients.

## Features

- Real-time messaging using WebSockets
- Multiple client support
- Simple web interface
- Hot reload development environment with Air
- Structured logging

## Project Structure

```
.
├── cmd/
│   └── real-time-chat/
│       └── main.go              # Application entry point
├── internal/
│   ├── logger/
│   │   ├── logger.go            # Logger interface
│   │   └── slog.go              # Structured logger implementation
│   ├── server/
│   │   └── server.go            # HTTP server setup
│   ├── templates/
│   │   └── index.html           # Web interface
│   └── websocket/
│       ├── client.go            # WebSocket client management
│       ├── handlers.go          # WebSocket handlers
│       └── hub.go               # Connection hub
├── tmp/                         # Build artifacts (ignored)
├── .air.toml                    # Air configuration for hot reload
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Prerequisites

- Go 1.21 or higher
- Air (for development with hot reload) - optional but recommended

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd real-time-chat
```

2. Install dependencies:
```bash
go mod download
```

3. (Optional) Install Air for hot reload during development:
```bash
go install github.com/cosmtrek/air@latest
```

## Running the Application

### Development Mode (with hot reload)

If you have Air installed, you can run the application with hot reload:

```bash
air
```

This will:
- Build the application to `./tmp/main`
- Start the server on `http://localhost:8080`
- Watch for file changes and automatically rebuild/restart

### Production Mode

To run the application directly:

```bash
go run ./cmd/real-time-chat
```

Or build and run:

```bash
go build -o ./tmp/main ./cmd/real-time-chat
./tmp/main
```

## Usage

1. Start the application using one of the methods above
2. Open your web browser and navigate to `http://localhost:8080`
3. Open multiple browser tabs/windows to simulate multiple users
4. Type messages and press Enter or click "Enviar" to send
5. Messages will appear in real-time across all connected clients

## Configuration

The application uses Air for development configuration. You can modify [.air.toml](.air.toml) to change:

- Build commands
- Watch directories
- File extensions to monitor
- Port and other settings

## WebSocket Endpoint

- **URL**: `ws://localhost:8080/ws`
- **Protocol**: WebSocket
- **Usage**: Connect to this endpoint to send/receive real-time messages

## Building for Production

To create a production build:

```bash
go build -ldflags="-w -s" -o ./bin/chat-server ./cmd/real-time-chat
```

## License

This project is open source and available under the [MIT License](LICENSE).