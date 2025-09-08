package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fortega2/real-time-chat/internal/handlers"
	"github.com/fortega2/real-time-chat/internal/repository"
	"github.com/go-chi/chi/v5"

	_ "github.com/mattn/go-sqlite3"
)

const (
	pathChannels          = "/channels"
	expectedChannelErrMsg = "expected channel %s, got %s"
	expectedCountErrMsg   = "expected count %d, got %d"
	failedCreateTestUser  = "Failed to create test user: %v"
	testChannelDesc       = "Test channel"
)

func TestGetAllChannels(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T) (*handlers.Handler, func())
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "Success - Empty Channels",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithChannels(t)
				queries := repository.New(db)
				h := handlers.NewHandler(getMockLogger(), queries, db)
				return h, func() { db.Close() }
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "Success - With Channels",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db, h := setupChannelTest(t)
				return h, func() { db.Close() }
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, teardown := tc.setup(t)
			defer teardown()

			req := httptest.NewRequest(http.MethodGet, pathChannels, nil)
			w := httptest.NewRecorder()

			h.GetAllChannels(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf(expectedStatusErrMsg, tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedStatus == http.StatusOK {
				if resp.Header.Get(headerContentType) != mimeApplicationJSON {
					t.Errorf(contentTypeErrFmt, resp.Header.Get(headerContentType))
				}
				checkChannelsResponse(t, w.Body, tc.expectedCount)
			}
		})
	}
}

func TestCreateChannel(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T) (*handlers.Handler, func())
		payload        interface{}
		expectedStatus int
	}{
		{
			name: "Successful Creation",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithChannels(t)
				queries := repository.New(db)

				_, err := queries.CreateUser(context.Background(), repository.CreateUserParams{
					Username: "channelcreator",
					Password: "password123",
				})
				if err != nil {
					t.Fatalf(failedCreateTestUser, err)
				}

				h := handlers.NewHandler(getMockLogger(), queries, db)
				return h, func() { db.Close() }
			},
			payload:        map[string]interface{}{"name": "general", "description": "General discussion", "userId": int64(1)},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Empty Channel Name",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithChannels(t)
				h := handlers.NewHandler(getMockLogger(), repository.New(db), db)
				return h, func() { db.Close() }
			},
			payload:        map[string]interface{}{"name": "", "description": "Empty name test", "userId": int64(1)},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid User ID - Zero",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithChannels(t)
				h := handlers.NewHandler(getMockLogger(), repository.New(db), db)
				return h, func() { db.Close() }
			},
			payload:        map[string]interface{}{"name": "test", "description": testChannelDesc, "userId": int64(0)},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing User ID",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithChannels(t)
				h := handlers.NewHandler(getMockLogger(), repository.New(db), db)
				return h, func() { db.Close() }
			},
			payload:        map[string]interface{}{"name": "test", "description": testChannelDesc},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid JSON",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithChannels(t)
				h := handlers.NewHandler(getMockLogger(), repository.New(db), db)
				return h, func() { db.Close() }
			},
			payload:        "invalid-json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "User Does Not Exist",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithChannels(t)
				h := handlers.NewHandler(getMockLogger(), repository.New(db), db)
				return h, func() { db.Close() }
			},
			payload:        map[string]interface{}{"name": "test", "description": testChannelDesc, "userId": int64(999)},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Channel Name Only (No Description)",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithChannels(t)
				queries := repository.New(db)

				_, err := queries.CreateUser(context.Background(), repository.CreateUserParams{
					Username: "testuser2",
					Password: "password123",
				})
				if err != nil {
					t.Fatalf(failedCreateTestUser, err)
				}

				h := handlers.NewHandler(getMockLogger(), queries, db)
				return h, func() { db.Close() }
			},
			payload:        map[string]interface{}{"name": "random", "userId": int64(1)},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "User Does Not Exist",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDBWithChannels(t)
				h := handlers.NewHandler(getMockLogger(), repository.New(db), db)
				return h, func() { db.Close() }
			},
			payload:        map[string]any{"name": "test", "description": testChannelDesc, "userId": int64(999)},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, teardown := tc.setup(t)
			defer teardown()

			body, _ := json.Marshal(tc.payload)
			req := httptest.NewRequest(http.MethodPost, pathChannels, bytes.NewBuffer(body))
			req.Header.Set(headerContentType, mimeApplicationJSON)

			w := httptest.NewRecorder()

			h.CreateChannel(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf(expectedStatusErrMsg, tc.expectedStatus, resp.StatusCode)
			}

			if tc.expectedStatus == http.StatusCreated {
				if resp.Header.Get(headerContentType) != mimeApplicationJSON {
					t.Errorf(contentTypeErrFmt, resp.Header.Get(headerContentType))
				}
				checkSuccessfulCreateChannelResponse(t, w.Body, tc.payload.(map[string]any))
			}
		})
	}
}

func TestCreateChannelInvalidJSONSyntax(t *testing.T) {
	db := initializeTestDBWithChannels(t)
	defer db.Close()
	h := handlers.NewHandler(getMockLogger(), repository.New(db), db)

	req := httptest.NewRequest(http.MethodPost, pathChannels, bytes.NewBufferString("{"))
	req.Header.Set(headerContentType, mimeApplicationJSON)

	w := httptest.NewRecorder()

	h.CreateChannel(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf(expectedStatusErrMsg, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestCreateChannelMissingUserID(t *testing.T) {
	db := initializeTestDBWithChannels(t)
	defer db.Close()
	h := handlers.NewHandler(getMockLogger(), repository.New(db), db)

	payload := map[string]string{"name": "test", "description": testChannelDesc}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, pathChannels, bytes.NewBuffer(body))
	req.Header.Set(headerContentType, mimeApplicationJSON)

	w := httptest.NewRecorder()

	h.CreateChannel(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf(expectedStatusErrMsg, http.StatusBadRequest, resp.StatusCode)
	}
}

func TestDeleteChannel(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(t *testing.T) (*handlers.Handler, int64, int64, func())
		channelID      string
		userID         string
		expectedStatus int
	}{
		{
			name:           "Successful Deletion",
			setup:          setupSuccessfulChannelDeletion,
			channelID:      "1",
			userID:         "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Channel Not Found",
			setup:          setupChannelNotFound,
			channelID:      "9999",
			userID:         "1",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "User Not Owner",
			setup:          setupUserNotOwner,
			channelID:      "1",
			userID:         "2",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Invalid Channel ID",
			setup:          setupBasicHandler,
			channelID:      "invalid",
			userID:         "1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid User ID",
			setup:          setupBasicHandler,
			channelID:      "1",
			userID:         "invalid",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing Channel ID",
			setup:          setupBasicHandler,
			channelID:      "",
			userID:         "1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing User ID",
			setup:          setupBasicHandler,
			channelID:      "1",
			userID:         "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			runDeleteChannelTest(t, tc.setup, tc.channelID, tc.userID, tc.expectedStatus, tc.name)
		})
	}
}

func setupSuccessfulChannelDeletion(t *testing.T) (*handlers.Handler, int64, int64, func()) {
	db := initializeTestDBWithChannels(t)
	queries := repository.New(db)

	user, err := queries.CreateUser(context.Background(), repository.CreateUserParams{
		Username: "channelowner",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf(failedCreateTestUser, err)
	}

	chId, err := queries.CreateChannel(context.Background(), repository.CreateChannelParams{
		Name:        "todelete",
		Description: sql.NullString{String: "Channel to delete", Valid: true},
		CreatedBy:   user.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create test channel: %v", err)
	}

	h := handlers.NewHandler(getMockLogger(), queries, db)
	return h, chId, user.ID, func() { db.Close() }
}

func setupChannelNotFound(t *testing.T) (*handlers.Handler, int64, int64, func()) {
	db := initializeTestDBWithChannels(t)
	queries := repository.New(db)

	user, err := queries.CreateUser(context.Background(), repository.CreateUserParams{
		Username: "testuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf(failedCreateTestUser, err)
	}

	h := handlers.NewHandler(getMockLogger(), queries, db)
	return h, 9999, user.ID, func() { db.Close() }
}

func setupUserNotOwner(t *testing.T) (*handlers.Handler, int64, int64, func()) {
	db := initializeTestDBWithChannels(t)
	queries := repository.New(db)

	owner, err := queries.CreateUser(context.Background(), repository.CreateUserParams{
		Username: "channelowner",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf(failedCreateTestUser, err)
	}

	notOwner, err := queries.CreateUser(context.Background(), repository.CreateUserParams{
		Username: "notowner",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to create not-owner user: %v", err)
	}

	chId, err := queries.CreateChannel(context.Background(), repository.CreateChannelParams{
		Name:        "ownedchannel",
		Description: sql.NullString{String: "Channel owned by first user", Valid: true},
		CreatedBy:   owner.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create test channel: %v", err)
	}

	h := handlers.NewHandler(getMockLogger(), queries, db)
	return h, chId, notOwner.ID, func() { db.Close() }
}

func setupBasicHandler(t *testing.T) (*handlers.Handler, int64, int64, func()) {
	db := initializeTestDBWithChannels(t)
	h := handlers.NewHandler(getMockLogger(), repository.New(db), db)
	return h, 0, 0, func() { db.Close() }
}

func runDeleteChannelTest(t *testing.T, setup func(t *testing.T) (*handlers.Handler, int64, int64, func()), channelID, userID string, expectedStatus int, testName string) {
	h, actualChannelID, actualUserID, teardown := setup(t)
	defer teardown()

	finalChannelID := channelID
	finalUserID := userID
	if actualChannelID != 0 && testName == "Successful Deletion" {
		finalChannelID = fmt.Sprintf("%d", actualChannelID)
		finalUserID = fmt.Sprintf("%d", actualUserID)
	}

	req := httptest.NewRequest(http.MethodDelete, "/channels/"+finalChannelID+"/users/"+finalUserID, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("channelId", finalChannelID)
	rctx.URLParams.Add("userId", finalUserID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.DeleteChannel(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Errorf(expectedStatusErrMsg, expectedStatus, resp.StatusCode)
	}

	if expectedStatus == http.StatusOK {
		validateSuccessfulDeletion(t, w, resp)
	}
}

func validateSuccessfulDeletion(t *testing.T, w *httptest.ResponseRecorder, resp *http.Response) {
	if resp.Header.Get(headerContentType) != mimeApplicationJSON {
		t.Errorf(contentTypeErrFmt, resp.Header.Get(headerContentType))
	}
	checkSuccessfulDeleteChannelResponse(t, w.Body)
}

func checkSuccessfulDeleteChannelResponse(t *testing.T, body *bytes.Buffer) {
	t.Helper()
	var response map[string]any
	if err := json.NewDecoder(body).Decode(&response); err != nil {
		t.Fatalf("failed to decode delete response: %v", err)
	}

	if message, exists := response["message"]; !exists || message == "" {
		t.Error("Expected response to have a non-empty 'message' field")
	}

	if channelID, exists := response["channelId"]; !exists || channelID == nil {
		t.Error("Expected response to have a non-nil 'channelId' field")
	}
}

func initializeTestDBWithChannels(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	createUsersTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        username TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`
	if _, err := db.Exec(createUsersTableSQL); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	createChannelsTableSQL := `
    CREATE TABLE IF NOT EXISTS channels (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL UNIQUE,
        description TEXT,
        created_by INTEGER NOT NULL,
        created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
    );`
	if _, err := db.Exec(createChannelsTableSQL); err != nil {
		t.Fatalf("Failed to create channels table: %v", err)
	}

	return db
}

func setupChannelTest(t *testing.T) (*sql.DB, *handlers.Handler) {
	t.Helper()
	db := initializeTestDBWithChannels(t)

	queries := repository.New(db)
	h := handlers.NewHandler(getMockLogger(), queries, db)

	user, err := queries.CreateUser(context.Background(), repository.CreateUserParams{
		Username: "channeltestuser",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf(failedCreateTestUser, err)
	}

	channels := []struct {
		name        string
		description string
	}{
		{"general", "General discussion"},
		{"random", "Random chat"},
	}

	for _, ch := range channels {
		_, err := queries.CreateChannel(context.Background(), repository.CreateChannelParams{
			Name: ch.name,
			Description: sql.NullString{
				String: ch.description,
				Valid:  ch.description != "",
			},
			CreatedBy: user.ID,
		})
		if err != nil {
			t.Fatalf("Failed to create test channel '%s': %v", ch.name, err)
		}
	}

	return db, h
}

func checkChannelsResponse(t *testing.T, body *bytes.Buffer, expectedCount int) {
	t.Helper()
	var channels []map[string]any
	if err := json.NewDecoder(body).Decode(&channels); err != nil {
		t.Fatalf("failed to decode channels response: %v", err)
	}

	if len(channels) != expectedCount {
		t.Errorf(expectedCountErrMsg, expectedCount, len(channels))
	}

	for _, channel := range channels {
		if _, ok := channel["id"]; !ok {
			t.Error("Expected channel to have 'id' field")
		}
		if _, ok := channel["name"]; !ok {
			t.Error("Expected channel to have 'name' field")
		}
		if _, ok := channel["description"]; !ok {
			t.Error("Expected channel to have 'description' field")
		}
		if _, ok := channel["createdBy"]; !ok {
			t.Error("Expected channel to have 'createdBy' field")
		}
		if _, ok := channel["createdAt"]; !ok {
			t.Error("Expected channel to have 'createdAt' field")
		}
	}
}

func checkSuccessfulCreateChannelResponse(t *testing.T, body *bytes.Buffer, expectedData map[string]interface{}) {
	t.Helper()
	var channel map[string]any
	if err := json.NewDecoder(body).Decode(&channel); err != nil {
		t.Fatalf("failed to decode channel response: %v", err)
	}

	if channel["name"] != expectedData["name"] {
		t.Errorf(expectedChannelErrMsg, expectedData["name"], channel["name"])
	}

	expectedDesc := ""
	if desc, exists := expectedData["description"]; exists {
		expectedDesc = desc.(string)
	}
	actualDesc := channel["description"].(string)
	if expectedDesc != actualDesc {
		t.Errorf("expected description %s, got %s", expectedDesc, actualDesc)
	}

	if channel["id"] == nil || channel["id"].(float64) == 0 {
		t.Error("Expected channel ID to be non-zero")
	}

	if channel["createdBy"] == nil || channel["createdBy"].(float64) == 0 {
		t.Error("Expected createdBy to be non-zero")
	}

	if channel["createdAt"] == nil || channel["createdAt"].(string) == "" {
		t.Error("Expected createdAt to be non-empty")
	}
}
