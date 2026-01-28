package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPRunHandler_Plan(t *testing.T) {
	s, err := NewServer()
	if err != nil {
		t.Fatalf("new server: %v", err)
	}

	body, _ := json.Marshal(map[string]any{"input": "my location weather"})
	req := httptest.NewRequest(http.MethodPost, "/run", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	s.RunHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["error"] != nil {
		t.Fatalf("unexpected error: %#v", resp["error"])
	}
	if resp["result"] == nil {
		t.Fatalf("missing result: %s", rr.Body.String())
	}
}
