package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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

// decodeJSON decodes the request body into dst with strict unknown-field rejection.
// The signature carries w to keep call-sites uniform with handlers that may want to
// return early after detecting an oversize body.
func decodeJSON(_ http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// badJSONMessage produces a user-facing error string for JSON decode failures
// without leaking internal struct field names or stack traces.
func badJSONMessage(err error) string {
	if err == nil {
		return "invalid json body"
	}
	if errors.Is(err, io.EOF) {
		return "request body is empty"
	}
	msg := err.Error()
	if strings.Contains(msg, "unknown field") {
		return "unexpected field in request body"
	}
	if strings.Contains(msg, "cannot unmarshal") {
		return "wrong type for one of the request fields"
	}
	return "invalid json body"
}

// normalizeRequiredText trims the input, validates it is non-empty,
// and enforces an upper length bound. Returns a user-facing error.
func normalizeRequiredText(s, fieldName string, maxLen int) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", fmt.Errorf("%s is required", fieldName)
	}
	if maxLen > 0 && len(s) > maxLen {
		return "", fmt.Errorf("%s must be at most %d characters", fieldName, maxLen)
	}
	return s, nil
}

// normalizeOptionalText trims the input and enforces an upper length bound.
// Empty input is allowed.
func normalizeOptionalText(s string, maxLen int) (string, error) {
	s = strings.TrimSpace(s)
	if maxLen > 0 && len(s) > maxLen {
		return "", fmt.Errorf("text must be at most %d characters", maxLen)
	}
	return s, nil
}

// normalizeSearchQuery trims the search input and caps its length.
func normalizeSearchQuery(s string) (string, error) {
	s = strings.TrimSpace(s)
	if len(s) > 120 {
		return "", errors.New("search query is too long")
	}
	return s, nil
}

// parseLimitParam reads ?limit=N from the URL with a default and a soft cap.
func parseLimitParam(r *http.Request, def int) (int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return def, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n <= 0 {
		return 0, errors.New("invalid limit")
	}
	if n > 200 {
		n = 200
	}
	return n, nil
}
