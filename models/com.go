package models

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func GetKey() string {
	file, err := os.Open("./.resources/.env")
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords) // use scanwords
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "apikey=") {
			return scanner.Text()[7:]
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
	}
	return ""
}

//Function helps with API Requests
func Send(method, URL string, data map[string]interface{}) (*http.Response,
	error) {
	client := &http.Client{}
	key := GetKey()
	dataJSON, _ := json.Marshal(data)

	req, _ := http.NewRequest(method, URL, bytes.NewBuffer(dataJSON))
	req.Header.Set("Authorization", "Bearer "+key)
	return client.Do(req)

}
