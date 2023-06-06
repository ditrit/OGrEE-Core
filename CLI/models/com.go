package models

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Function helps with API Requests
func Send(method, URL, key string, data map[string]any) (*http.Response, error) {
	client := &http.Client{}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, URL, bytes.NewBuffer(dataJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+key)
	return client.Do(req)
}
