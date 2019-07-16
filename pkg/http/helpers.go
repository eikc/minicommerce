package http

import (
	"encoding/json"
	"net/http"
)

func sendJSON(w http.ResponseWriter, status int, obj interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	w.WriteHeader(status)
	_, err = w.Write(b)
	return err
}
