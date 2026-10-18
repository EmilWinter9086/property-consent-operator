package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGrantRetriesAndKeepsWriteIdentity(t *testing.T) {
	var calls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/v1/auth/consent/grant/tenant-42" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatal("authorization header missing")
		}
		var body consentWrite
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body.UserID != "tenant-42" || body.Category != "essential" || body.IdempotencyKey != "cmd-9001" {
			t.Fatalf("body = %+v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"ok":false,"data":null,"error":{"code":"retry","message":"retry"},"metadata":{}}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true,"data":{},"error":null,"metadata":{}}`))
	}))
	defer server.Close()

	client := NewConsentClient("test-key")
	client.BaseURL = server.URL
	client.Sleep = func(context.Context, time.Duration) error { return nil }
	if err := client.Grant(context.Background(), "tenant-42", "essential", "cmd-9001"); err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}
