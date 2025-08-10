package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fortega2/real-time-chat/internal/handlers"
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/websocket"
	"github.com/go-chi/chi/v5"
)

const (
	expectedHandlerCreation = "Expected handler to be created, got nil"
)

func TestNewHandler(t *testing.T) {
	h := handlers.NewHandler(getMockLogger())

	if h == nil {
		t.Error(expectedHandlerCreation)
	}
}

func TestLogin(t *testing.T) {
	h := handlers.NewHandler(getMockLogger())

	if h == nil {
		t.Error(expectedHandlerCreation)
	}

	fakeUsername := "testuser"

	usernameByteJson, err := json.Marshal(map[string]string{
		"username": fakeUsername,
	})

	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(usernameByteJson))
	w := httptest.NewRecorder()
	h.Login(w, req)

	resp := w.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var user websocket.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if user.Username != fakeUsername {
		t.Errorf("Expected username %s, got %s", fakeUsername, user.Username)
	}
}

func TestGetUserByID(t *testing.T) {
	h := handlers.NewHandler(getMockLogger())

	if h == nil {
		t.Error(expectedHandlerCreation)
	}

	user := websocket.NewUser("testuser")
	websocket.AddUser(user)

	userIdStr := strconv.Itoa(user.ID)

	req := httptest.NewRequest("GET", "/user/"+userIdStr, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", userIdStr)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GetUserByID(w, req)

	resp := w.Result()
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("Failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %v, got %v", http.StatusOK, resp.Status)
	}

	var returnedUser websocket.User
	if err := json.NewDecoder(resp.Body).Decode(&returnedUser); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if returnedUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, returnedUser.ID)
	}

	if returnedUser.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, returnedUser.Username)
	}

	if returnedUser.Color != user.Color {
		t.Errorf("Expected color %s, got %s", user.Color, returnedUser.Color)
	}

	if returnedUser.JoinedAt.IsZero() {
		t.Error("Expected JoinedAt to be set, got zero value")
	}
}

func getMockLogger() logger.Logger {
	return logger.NewMockLogger()
}
