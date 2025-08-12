package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/fortega2/real-time-chat/internal/dto"
	"github.com/fortega2/real-time-chat/internal/handlers"
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/fortega2/real-time-chat/internal/websocket"
	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	expectedHandlerCreation    = "Expected handler to be created, got nil"
	failedToDecodeResponseBody = "Failed to decode response body: %v"
	expectedStatysErrMsg       = "Expected status %v, got %v"
	expectedUsernameErrMsg     = "Expected username %s, got %s"
)

func TestNewHandler(t *testing.T) {
	h := handlers.NewHandler(getMockLogger(), nil)
	if h == nil {
		t.Error(expectedHandlerCreation)
	}
}

func setupLoginTest(t *testing.T) (*sql.DB, *handlers.Handler) {
	db := initializeTestDB(t)

	queries := repository.New(db)
	h := handlers.NewHandler(getMockLogger(), queries)

	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	_, err = queries.CreateUser(context.Background(), repository.CreateUserParams{
		Username: "testuser",
		Password: string(hashedPassword),
	})
	if err != nil {
		t.Fatalf("Failed to create user for test: %v", err)
	}

	return db, h
}

func assertSuccessfulLogin(t *testing.T, resp *http.Response, payload map[string]string) {
	var userDto dto.UserDTO
	if err := json.NewDecoder(resp.Body).Decode(&userDto); err != nil {
		t.Fatalf(failedToDecodeResponseBody, err)
	}
	if userDto.Username != payload["username"] {
		t.Errorf(expectedUsernameErrMsg, payload["username"], userDto.Username)
	}
	if userDto.ID == 0 {
		t.Error("Expected user ID to be non-zero")
	}
}

func TestLoginUser(t *testing.T) {
	db, h := setupLoginTest(t)
	defer db.Close()

	testCases := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{
			name: "Successful Login",
			payload: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "User Not Found",
			payload: map[string]string{
				"username": "nonexistent",
				"password": "password123",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "Invalid Password",
			payload: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Empty Payload",
			payload:        map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing Password",
			payload: map[string]string{
				"username": "testuser",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			h.LoginUser(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf(expectedStatysErrMsg, tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedStatus == http.StatusOK {
				assertSuccessfulLogin(t, resp, tc.payload)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	h := handlers.NewHandler(getMockLogger(), nil)

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
		t.Errorf(expectedStatysErrMsg, http.StatusOK, resp.Status)
	}

	var returnedUser websocket.User
	if err := json.NewDecoder(resp.Body).Decode(&returnedUser); err != nil {
		t.Fatalf(failedToDecodeResponseBody, err)
	}

	if returnedUser.ID != user.ID {
		t.Errorf("Expected user ID %d, got %d", user.ID, returnedUser.ID)
	}

	if returnedUser.Username != user.Username {
		t.Errorf(expectedUsernameErrMsg, user.Username, returnedUser.Username)
	}

	if returnedUser.Color != user.Color {
		t.Errorf("Expected color %s, got %s", user.Color, returnedUser.Color)
	}

	if returnedUser.JoinedAt.IsZero() {
		t.Error("Expected JoinedAt to be set, got zero value")
	}
}

func TestCreateUser(t *testing.T) {
	db := initializeTestDB(t)
	defer db.Close()

	mockLogger := getMockLogger()
	queries := repository.New(db)
	h := handlers.NewHandler(mockLogger, queries)

	userData := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	body, _ := json.Marshal(userData)
	req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf(expectedStatysErrMsg, http.StatusCreated, resp.Status)
	}

	var userDto dto.UserDTO
	if err := json.NewDecoder(resp.Body).Decode(&userDto); err != nil {
		t.Fatalf(failedToDecodeResponseBody, err)
	}

	if userDto.Username != userData["username"] {
		t.Errorf(expectedUsernameErrMsg, userData["username"], userDto.Username)
	}

	if userDto.ID == 0 {
		t.Error("Expected user ID to be non-zero")
	}
}

func initializeTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}
	return db
}

func getMockLogger() logger.Logger {
	return logger.NewMockLogger()
}
