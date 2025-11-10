package util

import (
	"encoding/json"
	"net/http"
)

type H map[string]interface{}

// WriteJSON is a helper function for sending JSON responses.
// It sets the content type header and marshals the data to JSON.
func WriteJSON(w http.ResponseWriter, status int, data H) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}
