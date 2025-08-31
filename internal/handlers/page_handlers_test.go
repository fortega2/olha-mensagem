package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fortega2/real-time-chat/internal/handlers"
	"github.com/fortega2/real-time-chat/internal/logger"
)

func TestRootPage(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "index.html")
	testContent := "<html><body>Test Page</body></html>"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() { _ = os.Chdir(originalDir) }()

	templatesDir := filepath.Join(tempDir, "internal", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates directory: %v", err)
	}

	indexFile := filepath.Join(templatesDir, "index.html")
	if err := os.WriteFile(indexFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create index.html: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	h := handlers.NewHandler(logger.NewMockLogger(), nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.RootPage(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status %v, got %v", http.StatusOK, resp.StatusCode)
	}
}

func TestRootPageFileNotFound(t *testing.T) {
	h := handlers.NewHandler(logger.NewMockLogger(), nil)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	h.RootPage(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status %v, got %v", http.StatusNotFound, resp.StatusCode)
	}
}
