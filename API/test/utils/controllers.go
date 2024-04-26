package utils

import (
	"encoding/json"
	"net/http/httptest"
	"p3/models"
	"p3/test/e2e"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	REQUEST_WITH_USER = iota
	REQUEST_WITH_TOKEN
	REQUEST_WITH_HEADERS
	MANAGED_REQUEST
)

func GetUserToken(email string, password string) string {
	// It executes the user login and returns tha auth token
	acc, e := models.Login(email, password)
	if e != nil {
		return ""
	}
	return acc.Token
}

func validateRequest(t *testing.T, requestType int, httpMethod string, endpoint string, requestBody []byte, authId string, expectedStatus int, expectedMessage string) map[string]interface{} {
	// It executes a request and verifies the status code and the response message
	var recorder *httptest.ResponseRecorder
	if requestType == REQUEST_WITH_USER {
		// authId is the user name
		recorder = e2e.MakeRequestWithUser(httpMethod, endpoint, requestBody, authId)
	} else if requestType == MANAGED_REQUEST {
		// authId is ignored. The auth token is added by the method
		recorder = e2e.MakeRequest(httpMethod, endpoint, requestBody)
	} else if requestType == REQUEST_WITH_HEADERS {
		// authId is the header encoded in json format
		var headers map[string]string
		json.Unmarshal([]byte(authId), &headers)
		recorder = e2e.MakeRequestWithHeaders(httpMethod, endpoint, requestBody, headers)
	} else {
		// authId is the auth token
		recorder = e2e.MakeRequestWithToken(httpMethod, endpoint, requestBody, authId)
	}
	assert.Equal(t, expectedStatus, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, expectedMessage, message)
	return response
}

func ValidateManagedRequest(t *testing.T, httpMethod string, endpoint string, requestBody []byte, expectedStatus int, expectedMessage string) map[string]interface{} {
	return validateRequest(t, MANAGED_REQUEST, httpMethod, endpoint, requestBody, "", expectedStatus, expectedMessage)
}

func ValidateRequestWithUser(t *testing.T, httpMethod string, endpoint string, requestBody []byte, user string, expectedStatus int, expectedMessage string) map[string]interface{} {
	return validateRequest(t, REQUEST_WITH_USER, httpMethod, endpoint, requestBody, user, expectedStatus, expectedMessage)
}

func ValidateRequestWithToken(t *testing.T, httpMethod string, endpoint string, requestBody []byte, token string, expectedStatus int, expectedMessage string) map[string]interface{} {
	return validateRequest(t, REQUEST_WITH_TOKEN, httpMethod, endpoint, requestBody, token, expectedStatus, expectedMessage)
}

func ValidateRequestWithHeaders(t *testing.T, httpMethod string, endpoint string, requestBody []byte, header string, expectedStatus int, expectedMessage string) map[string]interface{} {
	return validateRequest(t, REQUEST_WITH_HEADERS, httpMethod, endpoint, requestBody, header, expectedStatus, expectedMessage)
}
