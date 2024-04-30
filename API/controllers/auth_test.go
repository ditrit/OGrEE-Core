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
	userNoPasswordEmail := "user_no_password@test.com"
	userWithPasswordEmail := "user_with_password@test.com"
	requestBody := []byte(`[
		{
			"name": "User With No Passsword",
			"roles": {
				"*": "manager"
			},
			"email": "` + userNoPasswordEmail + `"
		},
		{
			"name": "User With Passsword",
			"password": "fake_password",
			"roles": {
				"*": "user"
			},
			"email": "` + userWithPasswordEmail + `"
		}
	]`)

	recorder := e2e.MakeRequest("POST", test_utils.GetEndpoint("usersBulk"), requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	userWithoutPassword, exists := response[userNoPasswordEmail].(map[string]interface{})
	assert.True(t, exists)
	status, exists := userWithoutPassword["status"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created", status)
	// A random password should be created and passed in the response
	password, exists := userWithoutPassword["password"].(string)
	assert.True(t, exists)
	assert.True(t, len(password) > 0)

	userWithPassword, exists := response[userWithPasswordEmail].(map[string]interface{})
	assert.True(t, exists)
	status, exists = userWithPassword["status"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created", status)
	_, exists = userWithPassword["password"]
	assert.False(t, exists)

	// we delete the created users
	for _, userEmail := range []string{userWithPasswordEmail, userNoPasswordEmail} {
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

	e2e.ValidateManagedRequest(t, "POST", test_utils.GetEndpoint("login"), requestBody, http.StatusUnauthorized, "Invalid login credentials")
}

func TestLoginSuccess(t *testing.T) {
	userEmail, password := test_utils.CreateTestUser(t, "manager")
	requestBody := []byte(`{
		"email": "` + userEmail + `",
		"password": "` + password + `"
	}`)

	response := e2e.ValidateManagedRequest(t, "POST", test_utils.GetEndpoint("login"), requestBody, http.StatusOK, "Login succesful")

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
	e2e.ValidateManagedRequest(t, "GET", test_utils.GetEndpoint("tokenValid"), nil, http.StatusOK, "working")
}

func TestRequestWithInvalidAuthorizationHeader(t *testing.T) {
	endpoint := test_utils.GetEndpoint("users")
	tests := []struct {
		name    string
		header  string
		message string
	}{
		{"EmptyAuthorizationHeader", `{"Authorization": ""}`, "Missing auth token"},
		{"NoToken", `{"Authorization": "Basic"}`, "Invalid/Malformed auth token"},
		{"InvalidToken", `{"Authorization": "Basic invalid"}`, "Malformed authentication token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e2e.ValidateRequestWithHeaders(t, "GET", endpoint, nil, tt.header, http.StatusForbidden, tt.message)
		})
	}
}

func TestGetAllUsers(t *testing.T) {
	// As admin, we get all users
	response := e2e.ValidateManagedRequest(t, "GET", test_utils.GetEndpoint("users"), nil, http.StatusOK, "successfully got users")

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 3, len(data))
}

func TestGetUsersWithNormalUser(t *testing.T) {
	userEmail, password := test_utils.CreateTestUser(t, "user")
	userToken := test_utils.GetUserToken(userEmail, password)
	assert.NotEmpty(t, userToken)

	response := e2e.ValidateRequestWithToken(t, "GET", test_utils.GetEndpoint("users"), nil, userToken, http.StatusOK, "successfully got users")

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

	endpoint := test_utils.GetEndpoint("usersInstance", userId)
	e2e.ValidateRequestWithToken(t, "DELETE", endpoint, nil, userToken, http.StatusUnauthorized, "Caller does not have permission to delete this user")
}

func TestDeleteUser(t *testing.T) {
	// we get the user ID
	userEmail, _ := test_utils.CreateTestUser(t, "user")
	userId := models.GetUserByEmail(userEmail).ID.Hex()
	assert.NotEmpty(t, userId)

	endpoint := test_utils.GetEndpoint("usersInstance", userId)
	e2e.ValidateManagedRequest(t, "DELETE", endpoint, nil, http.StatusOK, "successfully removed user")

	// We get a Not Found if we try to delete again
	e2e.ValidateManagedRequest(t, "DELETE", endpoint, nil, http.StatusNotFound, "User not found")
}

func TestDeleteWithInvalidIdReturnsError(t *testing.T) {
	e2e.ValidateManagedRequest(t, "DELETE", test_utils.GetEndpoint("usersInstance", "unknown"), nil, http.StatusBadRequest, "User ID is not valid")
}

// Tests modify user role
func TestModifyRole(t *testing.T) {
	email, password := test_utils.CreateTestUser(t, "manager")
	userId := models.GetUserByEmail(email).ID.Hex()
	userToken := test_utils.GetUserToken(email, password)
	tests := []struct {
		name               string
		validationFunction func(*testing.T, string, string, []byte, string, int, string) map[string]interface{}
		authId             string
		userId             string
		requestBody        string
		statusCode         int
		message            string
	}{
		{"ExtraDataReturnsError", e2e.ValidateRequestWithToken, userToken, userId, `{"roles": {"*": "user"},"name": "other name"}`, http.StatusBadRequest, "Only 'roles' should be provided to patch"},
		{"InvalidRole", e2e.ValidateRequestWithToken, userToken, userId, `{"roles": {"*": "invalid"}}`, http.StatusInternalServerError, "Role assigned is not valid: "},
		{"InvalidId", e2e.ValidateRequestWithToken, userToken, "invalid", `{"roles": {"*": "user"}}`, http.StatusBadRequest, "User ID is not valid"},
		{"ModifyRoleWithNormalUser", e2e.ValidateRequestWithUser, "user", userId, `{"roles": {"*": "manager"}}`, http.StatusUnauthorized, "Caller does not have permission to modify this user"},
		{"Success", e2e.ValidateRequestWithToken, userToken, userId, `{"roles": {"*": "viewer"}}`, http.StatusOK, "successfully updated user roles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changePasswordEndpoint := test_utils.GetEndpoint("usersInstance", tt.userId)
			tt.validationFunction(t, "PATCH", changePasswordEndpoint, []byte(tt.requestBody), tt.authId, tt.statusCode, tt.message)
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
			e2e.ValidateRequestWithToken(t, "POST", changePasswordEndpoint, []byte(tt.requestBody), userToken, tt.statusCode, tt.message)
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
			e2e.ValidateRequestWithToken(t, "POST", resetPasswordEndpoint, []byte(tt.requestBody), tt.token, tt.statusCode, tt.message)
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
