package http

import (
	"encoding/json"
	"io"
	"io/ioutil"
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

func receiveJSON(body io.ReadCloser, obj interface{}) error {
	b, err := ioutil.ReadAll(body)
	defer body.Close()

	if err != nil {
		return err
	}

	return json.Unmarshal(b, obj)
}
