package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fortega2/real-time-chat/internal/dto"
	"github.com/fortega2/real-time-chat/internal/handlers"
	"github.com/fortega2/real-time-chat/internal/logger"
	"github.com/fortega2/real-time-chat/internal/repository"
	"golang.org/x/crypto/bcrypt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	failedToDecodeResponseBody = "failed to decode response body: %v"
	expectedStatusErrMsg       = "expected status %v, got %v"
	expectedUsernameErrMsg     = "expected username %s, got %s"
	pathUsers                  = "/users"
	pathLogin                  = "/login"
	headerContentType          = "Content-Type"
	mimeApplicationJSON        = "application/json"
	contentTypeErrFmt          = "expected Content-Type application/json, got %q"
)

func TestNewHandler(t *testing.T) {
	h := handlers.NewHandler(getMockLogger(), nil)
	if h == nil {
		t.Fatal("Expected handler to be created, got nil")
	}
}

func TestCreateUser(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T) (*handlers.Handler, func())
		payload        interface{}
		expectedStatus int
	}{
		{
			name: "Successful Creation",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDB(t)
				queries := repository.New(db)
				h := handlers.NewHandler(getMockLogger(), queries)
				return h, func() { db.Close() }
			},
			payload:        map[string]string{"username": "newuser", "password": "password123"},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Duplicate Username",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db, h := setupUserTest(t)
				return h, func() { db.Close() }
			},
			payload:        map[string]string{"username": "testuser", "password": "password123"},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "Empty Username",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDB(t)
				h := handlers.NewHandler(getMockLogger(), repository.New(db))
				return h, func() { db.Close() }
			},
			payload:        map[string]string{"username": "", "password": "password123"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Empty Password",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDB(t)
				h := handlers.NewHandler(getMockLogger(), repository.New(db))
				return h, func() { db.Close() }
			},
			payload:        map[string]string{"username": "someone", "password": ""},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid JSON",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDB(t)
				h := handlers.NewHandler(getMockLogger(), repository.New(db))
				return h, func() { db.Close() }
			},
			payload:        "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, teardown := tc.setup(t)
			defer teardown()

			body, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, pathUsers, bytes.NewBuffer(body))
			req.Header.Set(headerContentType, mimeApplicationJSON)
			w := httptest.NewRecorder()

			h.CreateUser(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf(expectedStatusErrMsg, tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedStatus == http.StatusCreated {
				if resp.Header.Get(headerContentType) != mimeApplicationJSON {
					t.Errorf(contentTypeErrFmt, resp.Header.Get(headerContentType))
				}
				checkSuccessfulCreateUserResponse(t, w.Body, tc.payload.(map[string]string)["username"])
			}
		})
	}
}

func TestCreateUserInvalidJSONSyntax(t *testing.T) {
	db := initializeTestDB(t)
	defer db.Close()
	h := handlers.NewHandler(getMockLogger(), repository.New(db))

	req := httptest.NewRequest(http.MethodPost, pathUsers, bytes.NewBufferString("{"))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	w := httptest.NewRecorder()

	h.CreateUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf(expectedStatusErrMsg, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestCreateUserPasswordStoredHashed(t *testing.T) {
	db := initializeTestDB(t)
	defer db.Close()
	queries := repository.New(db)
	h := handlers.NewHandler(getMockLogger(), queries)

	body, _ := json.Marshal(map[string]string{"username": "hashuser", "password": "secret"})
	req := httptest.NewRequest(http.MethodPost, pathUsers, bytes.NewBuffer(body))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	w := httptest.NewRecorder()

	h.CreateUser(w, req)
	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf(expectedStatusErrMsg, http.StatusCreated, resp.StatusCode)
	}

	u, err := queries.GetUserByUsername(context.Background(), "hashuser")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if u.Password == "secret" {
		t.Fatal("password stored in plain text")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte("secret")); err != nil {
		t.Fatalf("invalid bcrypt hash: %v", err)
	}
}

func TestLoginUser(t *testing.T) {
	db, h := setupUserTest(t)
	defer db.Close()

	testCases := []struct {
		name           string
		payload        map[string]string
		expectedStatus int
	}{
		{"Successful Login", map[string]string{"username": "testuser", "password": "password123"}, http.StatusOK},
		{"User Not Found", map[string]string{"username": "nonexistent", "password": "password123"}, http.StatusNotFound},
		{"Invalid Password", map[string]string{"username": "testuser", "password": "wrongpassword"}, http.StatusUnauthorized},
		{"Empty Payload", map[string]string{}, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, pathLogin, bytes.NewBuffer(body))
			req.Header.Set(headerContentType, mimeApplicationJSON)
			w := httptest.NewRecorder()

			h.LoginUser(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf(expectedStatusErrMsg, tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedStatus == http.StatusOK {
				if resp.Header.Get(headerContentType) != mimeApplicationJSON {
					t.Errorf(contentTypeErrFmt, resp.Header.Get(headerContentType))
				}
				checkSuccessfulLoginResponse(t, w.Body, tc.payload["username"])
			}
		})
	}
}

func TestLoginUserInvalidJSONSyntax(t *testing.T) {
	db := initializeTestDB(t)
	defer db.Close()
	h := handlers.NewHandler(getMockLogger(), repository.New(db))

	req := httptest.NewRequest(http.MethodPost, pathLogin, bytes.NewBufferString("{"))
	req.Header.Set(headerContentType, mimeApplicationJSON)
	w := httptest.NewRecorder()

	h.LoginUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf(expectedStatusErrMsg, http.StatusBadRequest, resp.StatusCode)
	}
}

func initializeTestDB(t *testing.T) *sql.DB {
	t.Helper()
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

func setupUserTest(t *testing.T) (*sql.DB, *handlers.Handler) {
	t.Helper()
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

func checkSuccessfulCreateUserResponse(t *testing.T, body *bytes.Buffer, username string) {
	t.Helper()
	var userDto dto.UserDTO
	if err := json.NewDecoder(body).Decode(&userDto); err != nil {
		t.Fatalf(failedToDecodeResponseBody, err)
	}
	if userDto.Username != username {
		t.Errorf(expectedUsernameErrMsg, username, userDto.Username)
	}
	if userDto.ID == 0 {
		t.Error("Expected user ID to be non-zero")
	}
}

func checkSuccessfulLoginResponse(t *testing.T, body *bytes.Buffer, username string) {
	t.Helper()
	var userDto dto.UserDTO
	if err := json.NewDecoder(body).Decode(&userDto); err != nil {
		t.Fatalf(failedToDecodeResponseBody, err)
	}
	if userDto.Username != username {
		t.Errorf(expectedUsernameErrMsg, username, userDto.Username)
	}
	if userDto.ID == 0 {
		t.Error("expected user ID to be non-zero")
	}
}
