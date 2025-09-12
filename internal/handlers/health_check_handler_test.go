package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fortega2/real-time-chat/internal/dto"
	"github.com/fortega2/real-time-chat/internal/handlers"
	"github.com/fortega2/real-time-chat/internal/repository"

	_ "github.com/mattn/go-sqlite3"
)

const (
	pathHealthCheck = "/health"
)

func TestHealthCheck(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T) (*handlers.Handler, func())
		expectedStatus int
		expectedHealth string
	}{
		{
			name: "Healthy Database",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDB(t)
				queries := repository.New(db)
				h := handlers.NewHandler(getMockLogger(), queries, db)
				return h, func() { db.Close() }
			},
			expectedStatus: http.StatusOK,
			expectedHealth: "healthy",
		},
		{
			name: "Unhealthy Database - Closed Connection",
			setup: func(t *testing.T) (*handlers.Handler, func()) {
				db := initializeTestDB(t)
				queries := repository.New(db)
				h := handlers.NewHandler(getMockLogger(), queries, db)
				db.Close()
				return h, func() {
					// No-op teardown since DB is already closed
				}
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedHealth: "unhealthy",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h, teardown := tc.setup(t)
			defer teardown()

			req := httptest.NewRequest(http.MethodGet, pathHealthCheck, nil)
			w := httptest.NewRecorder()

			h.HealthCheck(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tc.expectedStatus {
				t.Errorf(expectedStatusErrMsg, tc.expectedStatus, resp.StatusCode)
			}

			if resp.Header.Get(headerContentType) != mimeApplicationJSON {
				t.Errorf(contentTypeErrFmt, resp.Header.Get(headerContentType))
			}

			checkHealthCheckResponse(t, w, tc.expectedHealth)
		})
	}
}

func TestHealthCheckWithCancelledContext(t *testing.T) {
	db := initializeTestDB(t)
	defer db.Close()
	queries := repository.New(db)
	h := handlers.NewHandler(getMockLogger(), queries, db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := httptest.NewRequest(http.MethodGet, pathHealthCheck, nil)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestTimeout {
		t.Errorf(expectedStatusErrMsg, http.StatusRequestTimeout, resp.StatusCode)
	}
}

func TestHealthCheckVersion(t *testing.T) {
	testCases := []struct {
		name            string
		envVersion      string
		expectedVersion string
	}{
		{
			name:            "With APP_VERSION environment variable",
			envVersion:      "2.1.0",
			expectedVersion: "2.1.0",
		},
		{
			name:            "Without APP_VERSION environment variable",
			envVersion:      "",
			expectedVersion: "1.0.0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalVersion := os.Getenv("APP_VERSION")
			defer os.Setenv("APP_VERSION", originalVersion)

			if tc.envVersion == "" {
				os.Unsetenv("APP_VERSION")
			} else {
				os.Setenv("APP_VERSION", tc.envVersion)
			}

			db := initializeTestDB(t)
			defer db.Close()
			queries := repository.New(db)
			h := handlers.NewHandler(getMockLogger(), queries, db)

			req := httptest.NewRequest(http.MethodGet, pathHealthCheck, nil)
			w := httptest.NewRecorder()

			h.HealthCheck(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf(expectedStatusErrMsg, http.StatusOK, resp.StatusCode)
			}

			checkHealthCheckVersionResponse(t, w, tc.expectedVersion)
		})
	}
}

func checkHealthCheckResponse(t *testing.T, recorder *httptest.ResponseRecorder, expectedStatus string) {
	t.Helper()

	if recorder == nil {
		t.Fatal("Expected non-nil *httptest.ResponseRecorder")
	}

	var response dto.HealthCheckResponseDTO
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode health check response: %v", err)
	}

	if response.Status != dto.HealthStatus(expectedStatus) {
		t.Errorf("Expected status %s, got %s", expectedStatus, response.Status)
	}

	if response.Version == "" {
		t.Error("Expected version to be non-empty")
	}

	if expectedStatus == "unhealthy" && response.Error == "" {
		t.Error("Expected error message for unhealthy status")
	}

	if expectedStatus == "healthy" && response.Error != "" {
		t.Errorf("Expected no error for healthy status, got: %s", response.Error)
	}
}

func checkHealthCheckVersionResponse(t *testing.T, body interface{}, expectedVersion string) {
	t.Helper()

	recorder, ok := body.(*httptest.ResponseRecorder)
	if !ok {
		t.Fatal("Expected *httptest.ResponseRecorder")
	}

	var response dto.HealthCheckResponseDTO
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode health check response: %v", err)
	}

	if response.Version != expectedVersion {
		t.Errorf("Expected version %s, got %s", expectedVersion, response.Version)
	}
}
