// Filename: cmd/web/main_test.go

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"victortillett.net/basic/internal/ws"
)

func TestHandlerHome(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handlerHome(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v expected %v", status, http.StatusOK)
	}
	
	expected := "WebSockets!\n"
	if got := rr.Body.String(); got != expected {
		t.Errorf("handler returned unexpected body: got %q expected %q", got, expected)
	}
}

func TestWebSocketRoutes(t *testing.T) {
	// Test that routes are properly registered
	server := httptest.NewServer(http.HandlerFunc(ws.HandleWebSocket))
	defer server.Close()

	// Test GET request (should work)
	req1, _ := http.NewRequest("GET", server.URL, nil)
	req1.Header.Set("Origin", "http://localhost:4000")
	resp1, err := http.DefaultClient.Do(req1)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	// WebSocket upgrade fails in test environment but that's expected
	if resp1.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected BadRequest for test WebSocket, got %v", resp1.StatusCode)
	}

	// Test non-GET request (should be blocked)
	req2, _ := http.NewRequest("POST", server.URL, nil)
	resp2, _ := http.DefaultClient.Do(req2)
	if resp2.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected MethodNotAllowed for POST, got %v", resp2.StatusCode)
	}
}