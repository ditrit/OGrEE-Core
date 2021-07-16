package models

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

//Function helps with API Requests
func Send(method, URL string, data map[string]interface{}) (*http.Response,
	error) {
	client := &http.Client{}
	key := os.Getenv("apikey")
	dataJSON, _ := json.Marshal(data)

	req, _ := http.NewRequest(method, URL, bytes.NewBuffer(dataJSON))
	req.Header.Set("Authorization", "Bearer "+key)
	return client.Do(req)

}
