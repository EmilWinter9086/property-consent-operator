package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Resource string

const (
	MaintenanceRequest Resource = "maintenance_request"
	TenantDocument     Resource = "tenant_document"
	InspectionReminder Resource = "inspection_reminder"
)

var categories = map[Resource]string{
	MaintenanceRequest: "essential",
	TenantDocument:     "functional",
	InspectionReminder: "marketing",
}

type ConsentWriter interface {
	Grant(context.Context, string, string, string) error
	Revoke(context.Context, string, string, string) error
}

type ConsentCommand struct {
	UserID      string   `json:"user_id"`
	Resource    Resource `json:"resource"`
	Action      string   `json:"action"`
	OperationID string   `json:"operation_id"`
}

type ConsentResult struct {
	UserID   string   `json:"user_id"`
	Resource Resource `json:"resource"`
	Category string   `json:"category"`
	State    string   `json:"state"`
}

func ApplyConsent(ctx context.Context, client ConsentWriter, cmd ConsentCommand) (ConsentResult, error) {
	category, known := categories[cmd.Resource]
	if !known || strings.TrimSpace(cmd.UserID) == "" || strings.TrimSpace(cmd.OperationID) == "" {
		return ConsentResult{}, errors.New("user_id, operation_id, and a known resource are required")
	}

	result := ConsentResult{UserID: cmd.UserID, Resource: cmd.Resource, Category: category}
	switch cmd.Action {
	case "grant":
		if err := client.Grant(ctx, cmd.UserID, category, cmd.OperationID); err != nil {
			return ConsentResult{}, err
		}
		result.State = "granted"
	case "revoke":
		if err := client.Revoke(ctx, cmd.UserID, category, cmd.OperationID); err != nil {
			return ConsentResult{}, err
		}
		result.State = "revoked"
	default:
		return ConsentResult{}, fmt.Errorf("unknown action %q", cmd.Action)
	}
	return result, nil
}
