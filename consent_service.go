package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type ConsentService struct {
	Client ConsentWriter
}

func (s ConsentService) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /consent", s.apply)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	return mux
}

func (s ConsentService) apply(w http.ResponseWriter, r *http.Request) {
	var command ConsentCommand
	decoder := json.NewDecoder(io.LimitReader(r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&command); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
		return
	}
	result, err := ApplyConsent(r.Context(), s.Client, command)
	if err != nil {
		var apiErr *InfraiError
		if errors.As(err, &apiErr) && apiErr.HTTPStatus >= 400 && apiErr.HTTPStatus < 500 {
			writeJSON(w, apiErr.HTTPStatus, map[string]string{"error": apiErr.Code, "message": apiErr.Message})
			return
		}
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "consent request failed"})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
