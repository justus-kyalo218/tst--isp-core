package handlers

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"tst-isp/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.SetLevel(logger.ERROR) // Reduce noise in tests
	os.Exit(m.Run())
}

func TestLogin_InvalidMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	w := httptest.NewRecorder()

	Login(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

// Note: Full integration tests require DB setup. For now, test structure only.
