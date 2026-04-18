package handlers

import (
	"encoding/json"
	"net/http"
)

// errorResponse is the standard error envelope.
// swagger:model
type errorResponse struct {
	Error string `json:"error" example:"something went wrong"`
}

// listResponse is a generic list envelope used by list endpoints.
// swagger:model
type listResponse struct {
	Items any `json:"items"`
	Count int `json:"count" example:"10"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]any{"error": msg})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}