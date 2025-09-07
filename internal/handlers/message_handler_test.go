package handlers_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/fortega2/real-time-chat/internal/handlers"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/go-chi/chi/v5"
)

func TestGetHistoryMessagesByChannel(t *testing.T) {
	testCases := []struct {
		name           string
		channelID      string
		setup          func(t *testing.T) (*handlers.Handler, func())
		expectedStatus int
	}{
		{
			name:      "Successful Fetch",
			channelID: "1",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithMessages(t)
				queries := repository.New(db)
				h := handlers.NewHandler(getMockLogger(), queries)
				return h, func() { db.Close() }
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "Empty Channel ID",
			channelID: "",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDB(t)
				queries := repository.New(db)
				h := handlers.NewHandler(getMockLogger(), queries)
				return h, func() { db.Close() }
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Invalid Channel ID",
			channelID: "invalid",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDB(t)
				queries := repository.New(db)
				h := handlers.NewHandler(getMockLogger(), queries)
				return h, func() { db.Close() }
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "Non-existent Channel",
			channelID: "999",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithMessages(t)
				queries := repository.New(db)
				h := handlers.NewHandler(getMockLogger(), queries)
				return h, func() { db.Close() }
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, teardown := tc.setup(t)
			defer teardown()

			req := httptest.NewRequest(http.MethodGet, "/channels/"+tc.channelID+"/messages", nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("channelId", tc.channelID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			h.GetHistoryMessagesByChannel(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf(expectedStatusErrMsg, tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedStatus == http.StatusOK {
				if resp.Header.Get(headerContentType) != mimeApplicationJSON {
					t.Errorf(contentTypeErrFmt, resp.Header.Get(headerContentType))
				}
			}
		})
	}
}

func TestGetHistoryMessagesByChannelWithCustomLimit(t *testing.T) {
	originalLimit := os.Getenv("MESSAGES_LIMIT")
	defer os.Setenv("MESSAGES_LIMIT", originalLimit)

	os.Setenv("MESSAGES_LIMIT", "10")

	db := initializeTestDBWithMessages(t)
	defer db.Close()
	queries := repository.New(db)
	h := handlers.NewHandler(getMockLogger(), queries)

	req := httptest.NewRequest(http.MethodGet, "/channels/1/messages", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("channelId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetHistoryMessagesByChannel(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(expectedStatusErrMsg, http.StatusOK, resp.StatusCode)
	}
}

func TestGetHistoryMessagesByChannelInvalidLimit(t *testing.T) {
	originalLimit := os.Getenv("MESSAGES_LIMIT")
	defer os.Setenv("MESSAGES_LIMIT", originalLimit)

	os.Setenv("MESSAGES_LIMIT", "invalid")

	db := initializeTestDBWithMessages(t)
	defer db.Close()
	queries := repository.New(db)
	h := handlers.NewHandler(getMockLogger(), queries)

	req := httptest.NewRequest(http.MethodGet, "/channels/1/messages", nil)
	w := httptest.NewRecorder()

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("channelId", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	h.GetHistoryMessagesByChannel(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(expectedStatusErrMsg, http.StatusOK, resp.StatusCode)
	}
}

func initializeTestDBWithMessages(t *testing.T) *sql.DB {
	t.Helper()
	db := initializeTestDB(t)

	createChannelsTableSQL := `
	CREATE TABLE IF NOT EXISTS channels (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err := db.Exec(createChannelsTableSQL); err != nil {
		t.Fatalf("Failed to create channels table: %v", err)
	}

	createMessagesTableSQL := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		channel_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		user_color VARCHAR(7) NOT NULL,
		content TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (channel_id) REFERENCES channels(id) ON DELETE CASCADE,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`
	if _, err := db.Exec(createMessagesTableSQL); err != nil {
		t.Fatalf("Failed to create messages table: %v", err)
	}

	_, err := db.Exec("INSERT INTO channels (id, name) VALUES (1, 'test-channel')")
	if err != nil {
		t.Fatalf("Failed to insert test channel: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, username, password) VALUES (1, 'testuser', 'hashedpassword')")
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	for i := 1; i <= 5; i++ {
		_, err = db.Exec("INSERT INTO messages (channel_id, user_id, user_color, content) VALUES (1, 1, ?, ?)", "#3498db", "Test message "+strconv.Itoa(i))
		if err != nil {
			t.Fatalf("Failed to insert test message: %v", err)
		}
	}

	return db
}
