package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHealthHandler_Health(t *testing.T) {
	// Create a new health handler
	handler := NewHealthHandler("1.0.0")

	// Create a test request
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Call the handler
	handler.Health()(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected content type 'application/json', got '%s'", contentType)
	}

	// Parse response body
	var response HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check response fields
	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}

	if response.Service != "conduit-api" {
		t.Errorf("Expected service 'conduit-api', got '%s'", response.Service)
	}

	if response.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", response.Version)
	}

	// Check that timestamp is recent (within last 5 seconds)
	now := time.Now().UTC()
	if response.Timestamp.After(now) || response.Timestamp.Before(now.Add(-5*time.Second)) {
		t.Errorf("Timestamp %v is not recent (current time: %v)", response.Timestamp, now)
	}
}
