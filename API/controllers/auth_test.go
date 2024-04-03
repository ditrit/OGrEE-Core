package controllers_test

import (
	"encoding/json"
	"net/http"
	"p3/models"
	"p3/test/e2e"
	u "p3/utils"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getUserToken(email string, password string) string {
	acc, e := models.Login(email, password)
	if e != nil {
		return ""
	}
	return acc.Token
}

// Tests domain bulk creation (/api/users/bulk)
func TestCreateBulkUsersInvalidBody(t *testing.T) {
	requestBody := []byte(`[
		{
			"name": "invalid json body"",
		},
	]`)

	recorder := e2e.MakeRequest("POST", "/api/users/bulk", requestBody)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Invalid request", message)
}

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

	recorder := e2e.MakeRequest("POST", "/api/users/bulk", requestBody)
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
}

func TestLoginWrongPassword(t *testing.T) {
	requestBody := []byte(`{
		"email": "user_with_password@test.com",
		"password": "wrong_password"
	}`)

	recorder := e2e.MakeRequest("POST", "/api/login", requestBody)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Invalid login credentials", message)
}

func TestLoginSuccess(t *testing.T) {
	requestBody := []byte(`{
		"email": "user_with_password@test.com",
		"password": "fake_password"
	}`)

	recorder := e2e.MakeRequest("POST", "/api/login", requestBody)
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
	assert.Equal(t, "user_with_password@test.com", email)
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
	recorder := e2e.MakeRequestWithHeaders("GET", "/api/users", nil, header)
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
	recorder := e2e.MakeRequestWithHeaders("GET", "/api/users", nil, header)
	assert.Equal(t, http.StatusForbidden, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Invalid/Malformed auth token", message)
}

func TestRequestWithInvalidToken(t *testing.T) {
	recorder := e2e.MakeRequestWithToken("GET", "/api/users", nil, "invalid")
	assert.Equal(t, http.StatusForbidden, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Malformed authentication token", message)
}

func TestGetAllUsers(t *testing.T) {
	// As admin, we get all users

	recorder := e2e.MakeRequest("GET", "/api/users", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully got users", message)

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 5, len(data))
}

func TestGetUsersWithNormalUser(t *testing.T) {
	userToken := getUserToken("user_with_password@test.com", "fake_password")
	assert.NotEmpty(t, userToken)

	recorder := e2e.MakeRequestWithToken("GET", "/api/users", nil, userToken)
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
	userId := models.GetUserByEmail("user_no_password@test.com").ID.Hex()
	assert.NotEmpty(t, userId)
	userToken := getUserToken("user_with_password@test.com", "fake_password")
	assert.NotEmpty(t, userToken)

	recorder := e2e.MakeRequestWithToken("DELETE", "/api/users/"+userId, nil, userToken)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Caller does not have permission to delete this user", message)
}

func TestDeleteUser(t *testing.T) {
	// we get the user ID
	userId := models.GetUserByEmail("user_no_password@test.com").ID.Hex()
	assert.NotEmpty(t, userId)

	recorder := e2e.MakeRequest("DELETE", "/api/users/"+userId, nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully removed user", message)

	// We get a Not Found if we try to delete again
	recorder = e2e.MakeRequest("DELETE", "/api/users/"+userId, nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists = response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "User not found", message)

}

func TestDeleteWithInvalidIdReturnsError(t *testing.T) {
	recorder := e2e.MakeRequest("DELETE", "/api/users/unknown", nil)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "User ID is not valid", message)
}

// Tests modify user role
func TestModifyRoleWithMoreDataReturnsError(t *testing.T) {
	// we get the user ID
	userId := models.GetUserByEmail("user_with_password@test.com").ID.Hex()
	assert.NotEmpty(t, userId)

	requestBody := []byte(`{
		"roles": {
			"*": "user"
		},
		"name": "other name"
	}`)

	recorder := e2e.MakeRequest("PATCH", "/api/users/"+userId, requestBody)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Only 'roles' should be provided to patch", message)
}

func TestModifyRoleWithInvalidRole(t *testing.T) {
	// we get the user ID
	userId := models.GetUserByEmail("user_with_password@test.com").ID.Hex()
	assert.NotEmpty(t, userId)

	requestBody := []byte(`{
		"roles": {
			"*": "invalid"
		}
	}`)

	recorder := e2e.MakeRequest("PATCH", "/api/users/"+userId, requestBody)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Role assigned is not valid: ", message)
}

func TestModifyRoleWithInvalidId(t *testing.T) {
	requestBody := []byte(`{
		"roles": {
			"*": "user"
		}
	}`)

	recorder := e2e.MakeRequest("PATCH", "/api/users/invalid", requestBody)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "User ID is not valid", message)
}

func TestModifyRoleWithNormalUser(t *testing.T) {
	userId := models.GetUserByEmail("user_with_password@test.com").ID.Hex()
	assert.NotEmpty(t, userId)
	userToken := getUserToken("user_with_password@test.com", "fake_password")
	assert.NotEmpty(t, userToken)

	requestBody := []byte(`{
		"roles": {
			"*": "manager"
		}
	}`)

	recorder := e2e.MakeRequestWithToken("PATCH", "/api/users/"+userId, requestBody, userToken)
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Caller does not have permission to modify this user", message)
}

func TestModifyRoleSuccess(t *testing.T) {
	// we get the user ID
	userId := models.GetUserByEmail("user_with_password@test.com").ID.Hex()
	assert.NotEmpty(t, userId)

	requestBody := []byte(`{
		"roles": {
			"*": "viewer"
		}
	}`)

	recorder := e2e.MakeRequest("PATCH", "/api/users/"+userId, requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully updated user roles", message)
}

// Tests modify and reset user password
func TestModifyPasswordNotEnoughArguments(t *testing.T) {
	userToken := getUserToken("user_with_password@test.com", "fake_password")
	requestBody := []byte(`{
		"newPassword": "fake_password"
	}`)

	recorder := e2e.MakeRequestWithToken("POST", "/api/users/password/change", requestBody, userToken)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Invalid request: wrong body format", message)
}

func TestModifyPasswordSuccess(t *testing.T) {
	userToken := getUserToken("user_with_password@test.com", "fake_password")
	requestBody := []byte(`{
		"currentPassword": "fake_password",
		"newPassword": "fake_password2"
	}`)

	recorder := e2e.MakeRequestWithToken("POST", "/api/users/password/change", requestBody, userToken)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully updated user password", message)

	token, exists := response["token"].(string)
	assert.True(t, exists)
	assert.NotEmpty(t, token)
}

func TestResetPasswordErrorWhenResetTokenIsNotValid(t *testing.T) {
	// User token is not a reset token
	userToken := getUserToken("user_with_password@test.com", "fake_password2")
	requestBody := []byte(`{
		"newPassword": "fake_password"
	}`)

	recorder := e2e.MakeRequestWithToken("POST", "/api/users/password/reset", requestBody, userToken)
	assert.Equal(t, http.StatusForbidden, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Token is not valid.", message)
}

func TestResetPasswordNotEnoughArguments(t *testing.T) {
	userId := models.GetUserByEmail("user_with_password@test.com").ID
	resetToken := models.GenerateToken(u.RESET_TAG, userId, time.Minute)
	requestBody := []byte(`{}`)

	recorder := e2e.MakeRequestWithToken("POST", "/api/users/password/reset", requestBody, resetToken)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Invalid request: wrong body format", message)
}

func TestResetPasswordSuccess(t *testing.T) {
	userId := models.GetUserByEmail("user_with_password@test.com").ID
	resetToken := models.GenerateToken(u.RESET_TAG, userId, time.Minute)
	//current password is not needed
	requestBody := []byte(`{
		"newPassword": "fake_password"
	}`)

	recorder := e2e.MakeRequestWithToken("POST", "/api/users/password/reset", requestBody, resetToken)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully updated user password", message)
}
