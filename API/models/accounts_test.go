package models_test

import (
	"p3/models"
	test_utils "p3/test/utils"
	"testing"
)

func TestLoginToReturnFalse(t *testing.T) {
	tests := []struct {
		name         string
		mail         string
		password     string
		errorMessage string
	}{
		{"FalseLogin", "throwaway", "password", "Gave a false login request and did not receive error!"},
		{"EmptyEmail", "", "password", "Gave an empty email and did not receive error!"},
		{"EmptyEmailAndPassword", "", "", "Gave an empty email and did not receive error!"},
		{"EmptyPassword", "realcheat@gmail.com", "", "Gave an empty email and did not receive error!"},
		{"UserDoesNotExist", "realcheat@gmail.com", "password123", "Test Case 5 failed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := models.Login(tt.mail, tt.password); err == nil {
				t.Error(tt.errorMessage)
			}
		})
	}
}

// Thoroughly test the Validate() function
func TestValidateToReturnFalse(t *testing.T) {
	existingEmail, existingPassword := test_utils.CreateTestUser(t, "manager")
	var tst models.Account

	//Case 1
	if err := tst.Validate(); err.Message != "A valid email address is required" {
		t.Error("Gave empty email, but did not receive corresponding error")
	}

	//Case 2
	tst.Email = "realcheat"
	if err := tst.Validate(); err.Message != "A valid email address is required" {
		t.Error("Gave bad email, but did not receive corresponding error")
	}

	//Case 3
	tst.Email = "@"
	if err := tst.Validate(); err == nil {
		t.Error("Gave '@' as email, but did not receive corresponding error")
	}

	//Case 4
	tst.Email = "@"
	tst.Password = "secret123"
	if err := tst.Validate(); err == nil {
		t.Error(`Gave '@' as email and valid password, \
					but did not receive corresponding error`)
	}

	//Case 5
	tst.Email = "realcheat@"
	tst.Password = "secret123"
	if err := tst.Validate(); err == nil {
		t.Error("Test Case 5 failed!")
	}

	//Case 6
	tst.Email = "realcheat@orness.com"
	tst.Password = ""
	if err := tst.Validate(); err == nil {
		t.Error("Test Case 6 failed!")
	}

	//Case 7
	tst.Email = existingEmail
	tst.Password = existingPassword
	if err := tst.Validate(); err.Message != "Error: User already exists" {
		t.Error("Duplicate user did not return error")
	}
}

func TestValidateToReturnTrue(t *testing.T) {
	newAccount := models.Account{
		Email:    "test@test.com",
		Password: "password123",
		Roles:    map[string]models.Role{"*": "user"},
	}
	if err := newAccount.Validate(); err != nil {
		t.Error("Validate should return nil")
	}
}
