// Package httpx holds the HTTP helpers shared by all controller packages.
package httpx

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

// ErrorResponse is the JSON envelope returned by WriteError. It lives here
// (rather than in the view package) because it is an HTTP-transport concern,
// not a domain response shape.
type ErrorResponse struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, ErrorResponse{Error: msg})
}

func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func ParseID(r *http.Request, name string) (uuid.UUID, error) {
	return uuid.Parse(r.PathValue(name))
}
