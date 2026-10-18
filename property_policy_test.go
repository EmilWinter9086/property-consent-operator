package main

import (
	"context"
	"testing"
)

type recordingConsent struct {
	operation string
	userID    string
	category  string
	requestID string
}

func (r *recordingConsent) Grant(_ context.Context, userID, category, requestID string) error {
	r.operation, r.userID, r.category, r.requestID = "grant", userID, category, requestID
	return nil
}

func (r *recordingConsent) Revoke(_ context.Context, userID, category, requestID string) error {
	r.operation, r.userID, r.category, r.requestID = "revoke", userID, category, requestID
	return nil
}

func TestApplyConsentRoutesPropertyScopes(t *testing.T) {
	tests := []struct {
		name, action, category, operation, state string
		resource                                 Resource
	}{
		{"grant maintenance handling", "grant", "essential", "grant", "granted", MaintenanceRequest},
		{"revoke document access", "revoke", "functional", "revoke", "revoked", TenantDocument},
		{"grant inspection reminders", "grant", "marketing", "grant", "granted", InspectionReminder},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &recordingConsent{}
			result, err := ApplyConsent(context.Background(), client, ConsentCommand{
				UserID: "tenant-42", Resource: tt.resource, Action: tt.action, OperationID: "cmd-9001",
			})
			if err != nil {
				t.Fatal(err)
			}
			if result.Category != tt.category || result.State != tt.state {
				t.Fatalf("result = %+v", result)
			}
			if client.operation != tt.operation || client.userID != "tenant-42" || client.category != tt.category || client.requestID != "cmd-9001" {
				t.Fatalf("write = %+v", client)
			}
		})
	}
}
