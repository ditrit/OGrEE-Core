package models

import (
	"bytes"
	"encoding/json"
	"net/http"
)

//Function helps with API Requests
func Send(method, URL, key string, data map[string]interface{}) (*http.Response,
	error) {
	//Loop because sometimes a
	//Stream Error occurs
	//thus give max 400 attempts before returning error
	sender := func(method, URL, key string, data map[string]interface{}) (*http.Response, error) {
		client := &http.Client{}
		dataJSON, _ := json.Marshal(data)

		req, _ := http.NewRequest(method, URL, bytes.NewBuffer(dataJSON))
		req.Header.Set("Authorization", "Bearer "+key)
		return client.Do(req)
	}

	for i := 0; ; i++ {
		r, e := sender(method, URL, key, data)
		if e == nil {
			return r, e
		}

		if i == 400 {
			return r, e
		}
	}
}
