package models

import (
	"bytes"
	"encoding/json"
	"net/http"
)

//Function helps with API Requests
func Send(method, URL, key string, data map[string]interface{}) (*http.Response,
	error) {
	client := &http.Client{}
	dataJSON, _ := json.Marshal(data)

	req, _ := http.NewRequest(method, URL, bytes.NewBuffer(dataJSON))
	req.Header.Set("Authorization", "Bearer "+key)
	return client.Do(req)

}

//Function communicates with Unity
func ContactUnity(method, URL string, data map[string]interface{}) (*http.Response, error) {
	client := &http.Client{}
	dataJSON, _ := json.Marshal(data)

	req, _ := http.NewRequest(method, URL, bytes.NewBuffer(dataJSON))
	return client.Do(req)
}
