package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"configaudit/internal/app"
)

func TestInfoEndpointReturnsStatus(t *testing.T) {
	handler := NewHandler(app.NewService())
	request := httptest.NewRequest(http.MethodGet, "/_info", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload["name"] != "configaudit" || payload["status"] != "ok" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestScanEndpointReturnsFindings(t *testing.T) {
	handler := NewHandler(app.NewService())
	request := httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader("storage:\n  digest-algorithm: MD5\n"))
	request.Header.Set("Content-Type", "application/yaml")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}

	var payload scanResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload.Problems) != 1 {
		t.Fatalf("expected 1 problem, got %d", len(payload.Problems))
	}
	if payload.Problems[0].Path != "storage.digest-algorithm" {
		t.Fatalf("unexpected problem path %q", payload.Problems[0].Path)
	}
}

func TestScanEndpointRejectsInvalidConfig(t *testing.T) {
	handler := NewHandler(app.NewService())
	request := httptest.NewRequest(http.MethodPost, "/scan", strings.NewReader("storage:\n  digest-algorithm: [\n"))
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", recorder.Code)
	}

	var payload errorResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Error == "" {
		t.Fatal("expected error message")
	}
}
