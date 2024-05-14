package e2e

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"p3/app"
	"p3/models"
	"p3/router"
	_ "p3/test/integration"
	"testing"

	"github.com/elliotchance/pie/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var appRouter *mux.Router
var users map[string]any

const (
	REQUEST_WITH_USER = iota
	REQUEST_WITH_TOKEN
	REQUEST_WITH_HEADERS
	MANAGED_REQUEST
)

func init() {
	appRouter = router.Router(app.JwtAuthentication)
	users = map[string]any{}
	createUser("admin", map[string]models.Role{"*": "manager"})
	createUser("user", map[string]models.Role{"*": "user"})
	createUser("viewer", map[string]models.Role{"*": "viewer"})
}

func createUser(userType string, role map[string]models.Role) {
	user := models.Account{}
	user.Email = userType + "@" + userType + ".com"
	user.Password = userType + "123"
	user.Roles = role

	newAcc, err := user.Create(map[string]models.Role{"*": "manager"})
	if err != nil {
		log.Fatalln("Error while creating "+userType+"account:", err.Error())
	}

	if newAcc != nil {
		users[userType] = map[string]any{
			"id":    newAcc.ID,
			"token": newAcc.Token,
		}
	}
}

func MakeRequestWithHeaders(method, url string, requestBody []byte, header map[string]string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest(method, url, bytes.NewBuffer(requestBody))
	for key, value := range header {
		request.Header.Set(key, value)
	}
	appRouter.ServeHTTP(recorder, request)
	return recorder
}

func MakeRequestWithToken(method, url string, requestBody []byte, token string) *httptest.ResponseRecorder {
	header := map[string]string{
		"Authorization": "Bearer " + token,
	}
	return MakeRequestWithHeaders(method, url, requestBody, header)
}

func MakeRequestWithUser(method, url string, requestBody []byte, user string) *httptest.ResponseRecorder {
	token := users[user].(map[string]any)["token"].(string)
	return MakeRequestWithToken(method, url, requestBody, token)
}

func MakeRequest(method, url string, requestBody []byte) *httptest.ResponseRecorder {
	return MakeRequestWithUser(method, url, requestBody, "admin")
}

func GetObjects(queryParams string) (*httptest.ResponseRecorder, []map[string]any) {
	response := MakeRequest(http.MethodGet, router.GenericObjectsURL+"?"+queryParams, nil)

	var objects []map[string]any
	if response.Code == http.StatusOK {
		var responseBody map[string]interface{}
		json.Unmarshal(response.Body.Bytes(), &responseBody)
		objects = pie.Map(responseBody["data"].([]any), func(objectAny any) map[string]any {
			return objectAny.(map[string]any)
		})
	}

	return response, objects
}

func TestInvalidBody(t *testing.T, httpMethod string, endpoint string, errorMessage string) {
	invalidBody := []byte(`{`)

	recorder := MakeRequest(httpMethod, endpoint, invalidBody)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, errorMessage, message)
}

func validateRequest(t *testing.T, requestType int, httpMethod string, endpoint string, requestBody []byte, authId string, expectedStatus int, expectedMessage string) map[string]interface{} {
	// It executes a request and verifies the status code and the response message
	var recorder *httptest.ResponseRecorder
	if requestType == REQUEST_WITH_USER {
		// authId is the user name
		recorder = MakeRequestWithUser(httpMethod, endpoint, requestBody, authId)
	} else if requestType == MANAGED_REQUEST {
		// authId is ignored. The auth token is added by the method
		recorder = MakeRequest(httpMethod, endpoint, requestBody)
	} else if requestType == REQUEST_WITH_HEADERS {
		// authId is the header encoded in json format
		var headers map[string]string
		json.Unmarshal([]byte(authId), &headers)
		recorder = MakeRequestWithHeaders(httpMethod, endpoint, requestBody, headers)
	} else {
		// authId is the auth token
		recorder = MakeRequestWithToken(httpMethod, endpoint, requestBody, authId)
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
