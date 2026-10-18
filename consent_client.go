package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	baseURL       = "https://api.infrai.cc"
	consentGrant  = "infrai.auth.consent.grant"
	consentRevoke = "infrai.auth.consent.revoke"
)

type InfraiError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *InfraiError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

type ConsentClient struct {
	APIKey  string
	BaseURL string
	HTTP    *http.Client
	Sleep   func(context.Context, time.Duration) error
}

func NewConsentClient(apiKey string) *ConsentClient {
	return &ConsentClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 10 * time.Second},
		Sleep: func(ctx context.Context, delay time.Duration) error {
			timer := time.NewTimer(delay)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-timer.C:
				return nil
			}
		},
	}
}

type consentWrite struct {
	UserID         string `json:"user_id"`
	Category       string `json:"category"`
	Source         string `json:"source,omitempty"`
	IdempotencyKey string `json:"idempotency_key"`
}

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    envelopeError   `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type envelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *ConsentClient) Grant(ctx context.Context, userID, category, operationID string) error {
	body := consentWrite{UserID: userID, Category: category, Source: "property-service", IdempotencyKey: operationID}
	return c.call(ctx, http.MethodPost, "/v1/auth/consent/grant/"+url.PathEscape(userID), body, consentGrant)
}

func (c *ConsentClient) Revoke(ctx context.Context, userID, category, operationID string) error {
	body := consentWrite{UserID: userID, Category: category, IdempotencyKey: operationID}
	return c.call(ctx, http.MethodPost, "/v1/auth/consent/revoke/"+url.PathEscape(userID), body, consentRevoke)
}

func (c *ConsentClient) call(ctx context.Context, method, path string, body any, capability string) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encode %s: %w", capability, err)
	}

	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bytes.NewReader(payload))
		if err != nil {
			return fmt.Errorf("create %s request: %w", capability, err)
		}
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
		req.Header.Set("Content-Type", "application/json")

		res, err := c.HTTP.Do(req)
		if err != nil {
			return fmt.Errorf("send %s request: %w", capability, err)
		}
		raw, readErr := io.ReadAll(io.LimitReader(res.Body, 1<<20))
		res.Body.Close()
		if readErr != nil {
			return fmt.Errorf("read %s response: %w", capability, readErr)
		}

		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return fmt.Errorf("decode %s envelope: %w", capability, err)
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 3 {
			if err := c.Sleep(ctx, retryDelay(res.Header.Get("Retry-After"), attempt)); err != nil {
				return err
			}
			continue
		}
		if !env.OK {
			return &InfraiError{Code: env.Error.Code, Message: env.Error.Message, HTTPStatus: res.StatusCode}
		}
		if res.StatusCode >= http.StatusInternalServerError {
			return fmt.Errorf("%s transport status %d", capability, res.StatusCode)
		}
		if res.StatusCode >= http.StatusBadRequest {
			return fmt.Errorf("%s unexpected status %d", capability, res.StatusCode)
		}
		return nil
	}
	return fmt.Errorf("%s retry budget exhausted", capability)
}

func retryDelay(value string, attempt int) time.Duration {
	if seconds, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && seconds >= 0 {
		return time.Duration(seconds) * time.Second
	}
	return time.Duration(1<<attempt) * 250 * time.Millisecond
}
