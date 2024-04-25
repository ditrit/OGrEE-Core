package controllers_test

import (
	"encoding/json"
	"net/http"
	"p3/models"
	"p3/test/e2e"
	test_utils "p3/test/utils"
	u "p3/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateBulkUsers(t *testing.T) {
	// Test create two separate users
	requestBody := []byte(`[
		{
			"name": "User With No Passsword",
			"roles": {
				"*": "manager"
			},
			"email": "user_no_password@test.com"
		},
		{
			"name": "User With Passsword",
			"password": "fake_password",
			"roles": {
				"*": "user"
			},
			"email": "user_with_password@test.com"
		}
	]`)

	recorder := e2e.MakeRequest("POST", test_utils.GetEndpoint("usersBulk"), requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	userWithoutPassword, exists := response["user_no_password@test.com"].(map[string]interface{})
	assert.True(t, exists)
	status, exists := userWithoutPassword["status"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created", status)
	// A random password should be created and passed in the response
	password, exists := userWithoutPassword["password"].(string)
	assert.True(t, exists)
	assert.True(t, len(password) > 0)

	userWithPassword, exists := response["user_with_password@test.com"].(map[string]interface{})
	assert.True(t, exists)
	status, exists = userWithPassword["status"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created", status)
	_, exists = userWithPassword["password"]
	assert.False(t, exists)

	// we delete the created users
	for _, userEmail := range []string{"user_with_password@test.com", "user_no_password@test.com"} {
		models.DeleteUser(models.GetUserByEmail(userEmail).ID)
	}
}

// Tests Login
func TestLoginWrongPassword(t *testing.T) {
	email, _ := test_utils.CreateTestUser(t, "manager")
	requestBody := []byte(`{
		"email": "` + email + `",
		"password": "wrong_password"
	}`)

	recorder := e2e.MakeRequest("POST", test_utils.GetEndpoint("login"), requestBody)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Invalid login credentials", message)
}

func TestLoginSuccess(t *testing.T) {
	userEmail, password := test_utils.CreateTestUser(t, "manager")
	requestBody := []byte(`{
		"email": "` + userEmail + `",
		"password": "` + password + `"
	}`)

	recorder := e2e.MakeRequest("POST", test_utils.GetEndpoint("login"), requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Login succesful", message)

	account, exists := response["account"].(map[string]interface{})
	assert.True(t, exists)
	email, exists := account["email"].(string)
	assert.True(t, exists)
	assert.Equal(t, userEmail, email)
	token, exists := account["token"].(string)
	assert.True(t, exists)
	assert.NotEmpty(t, token)
}

func TestVerifyToken(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/token/valid", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "working", message)
}

func TestRequestWithEmptyAuthorizationHeader(t *testing.T) {
	header := map[string]string{
		"Authorization": "",
	}
	recorder := e2e.MakeRequestWithHeaders("GET", test_utils.GetEndpoint("users"), nil, header)
	assert.Equal(t, http.StatusForbidden, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Missing auth token", message)
}

func TestRequestWithNoToken(t *testing.T) {
	header := map[string]string{
		"Authorization": "Basic",
	}
	recorder := e2e.MakeRequestWithHeaders("GET", test_utils.GetEndpoint("users"), nil, header)
	assert.Equal(t, http.StatusForbidden, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Invalid/Malformed auth token", message)
}

func TestRequestWithInvalidToken(t *testing.T) {
	recorder := e2e.MakeRequestWithToken("GET", test_utils.GetEndpoint("users"), nil, "invalid")
	assert.Equal(t, http.StatusForbidden, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Malformed authentication token", message)
}

func TestGetAllUsers(t *testing.T) {
	// As admin, we get all users

	recorder := e2e.MakeRequest("GET", test_utils.GetEndpoint("users"), nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully got users", message)

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 3, len(data))
}

func TestGetUsersWithNormalUser(t *testing.T) {
	userEmail, password := test_utils.CreateTestUser(t, "user")
	userToken := test_utils.GetUserToken(userEmail, password)
	assert.NotEmpty(t, userToken)

	recorder := e2e.MakeRequestWithToken("GET", test_utils.GetEndpoint("users"), nil, userToken)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully got users", message)

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 0, len(data))
}

func TestDeleteWithoutEnoughPermissions(t *testing.T) {
	userEmail, password := test_utils.CreateTestUser(t, "user")
	userId := models.GetUserByEmail("admin@admin.com").ID.Hex()
	assert.NotEmpty(t, userId)
	userToken := test_utils.GetUserToken(userEmail, password)
	assert.NotEmpty(t, userToken)

	recorder := e2e.MakeRequestWithToken("DELETE", test_utils.GetEndpoint("usersInstance", userId), nil, userToken)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Caller does not have permission to delete this user", message)
}

func TestDeleteUser(t *testing.T) {
	// we get the user ID
	userEmail, _ := test_utils.CreateTestUser(t, "user")
	userId := models.GetUserByEmail(userEmail).ID.Hex()
	assert.NotEmpty(t, userId)

	recorder := e2e.MakeRequest("DELETE", test_utils.GetEndpoint("usersInstance", userId), nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully removed user", message)

	// We get a Not Found if we try to delete again
	recorder = e2e.MakeRequest("DELETE", test_utils.GetEndpoint("usersInstance", userId), nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists = response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "User not found", message)

}

func TestDeleteWithInvalidIdReturnsError(t *testing.T) {
	recorder := e2e.MakeRequest("DELETE", test_utils.GetEndpoint("usersInstance", "unknown"), nil)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "User ID is not valid", message)
}

// Tests modify user role
func TestModifyRole(t *testing.T) {
	email, password := test_utils.CreateTestUser(t, "manager")
	userId := models.GetUserByEmail(email).ID.Hex()
	userToken := test_utils.GetUserToken(email, password)
	tests := []struct {
		name        string
		requestType int
		authId      string
		userId      string
		requestBody string
		statusCode  int
		message     string
	}{
		{"ExtraDataReturnsError", test_utils.REQUEST_WITH_TOKEN, userToken, userId, `{"roles": {"*": "user"},"name": "other name"}`, http.StatusBadRequest, "Only 'roles' should be provided to patch"},
		{"InvalidRole", test_utils.REQUEST_WITH_TOKEN, userToken, userId, `{"roles": {"*": "invalid"}}`, http.StatusInternalServerError, "Role assigned is not valid: "},
		{"InvalidId", test_utils.REQUEST_WITH_TOKEN, userToken, "invalid", `{"roles": {"*": "user"}}`, http.StatusBadRequest, "User ID is not valid"},
		{"ModifyRoleWithNormalUser", test_utils.REQUEST_WITH_USER, "user", userId, `{"roles": {"*": "manager"}}`, http.StatusUnauthorized, "Caller does not have permission to modify this user"},
		{"Success", test_utils.REQUEST_WITH_TOKEN, userToken, userId, `{"roles": {"*": "viewer"}}`, http.StatusOK, "successfully updated user roles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changePasswordEndpoint := test_utils.GetEndpoint("usersInstance", tt.userId)
			test_utils.ValidateRequest(t, tt.requestType, "PATCH", changePasswordEndpoint, []byte(tt.requestBody), tt.authId, tt.statusCode, tt.message)
		})
	}
}

// Tests modify and reset user password
func TestModifyPassword(t *testing.T) {
	email, password := test_utils.CreateTestUser(t, "manager")
	userToken := test_utils.GetUserToken(email, password)
	correctRequestBody := `{
		"currentPassword": "` + password + `",
		"newPassword": "fake_password2"
	}`
	tests := []struct {
		name        string
		requestBody string
		statusCode  int
		message     string
	}{
		{"NotEnoughArguments", `{"newPassword": "fake_password"}`, http.StatusBadRequest, "Invalid request: wrong body format"},
		{"Success", correctRequestBody, http.StatusOK, "successfully updated user password"},
	}

	changePasswordEndpoint := test_utils.GetEndpoint("changePassword")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test_utils.ValidateRequest(t, test_utils.REQUEST_WITH_TOKEN, "POST", changePasswordEndpoint, []byte(tt.requestBody), userToken, tt.statusCode, tt.message)
		})
	}
}

func TestResetPassword(t *testing.T) {
	email, password := test_utils.CreateTestUser(t, "manager")
	userId := models.GetUserByEmail(email).ID
	correctRequestBody := `{"newPassword": "fake_password"}`
	tests := []struct {
		name        string
		token       string
		requestBody string
		statusCode  int
		message     string
	}{
		{"InvalidResetToken", test_utils.GetUserToken(email, password), correctRequestBody, http.StatusForbidden, "Token is not valid."}, // User token is not a reset token
		{"NotEnoughArguments", models.GenerateToken(u.RESET_TAG, userId, time.Minute), `{}`, http.StatusBadRequest, "Invalid request: wrong body format"},
		{"Success", models.GenerateToken(u.RESET_TAG, userId, time.Minute), correctRequestBody, http.StatusOK, "successfully updated user password"},
	}

	resetPasswordEndpoint := test_utils.GetEndpoint("resetPassword")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test_utils.ValidateRequest(t, test_utils.REQUEST_WITH_TOKEN, "POST", resetPasswordEndpoint, []byte(tt.requestBody), tt.token, tt.statusCode, tt.message)
		})
	}
}

// Tests with invalid body
func TestRequestsWithInvalidBody(t *testing.T) {
	email, _ := test_utils.CreateTestUser(t, "manager")
	userId := models.GetUserByEmail(email).ID.Hex()
	tests := []struct {
		name          string
		requestMethod string
		endpoint      string
		message       string
	}{
		{"CreateUser", "POST", test_utils.GetEndpoint("users"), "Invalid request: wrong format body"},
		{"CreateBulkUsers", "POST", test_utils.GetEndpoint("usersBulk"), "Invalid request"},
		{"Login", "POST", test_utils.GetEndpoint("login"), "Invalid request"},
		{"ModifyUser", "PATCH", test_utils.GetEndpoint("usersInstance", userId), "Invalid request"},
		{"ModifyPassword", "POST", test_utils.GetEndpoint("changePassword"), "Invalid request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e2e.TestInvalidBody(t, tt.requestMethod, tt.endpoint, tt.message)
		})
	}

}
