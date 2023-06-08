package models

import (
	"testing"
)

func TestLoginToReturnFalse(t *testing.T) {
	//fmt.Println(res)
	//fmt.Println(reflect.TypeOf(res["status"]))

	//Test Case 1
	if _, err := Login("throwaway", "password"); err == nil {
		t.Error("Gave a false login request and did not receive error!")
	}

	//Test Case 2
	if _, err := Login("", "password"); err == nil {
		t.Error("Gave an empty email and did not receive error!")
	}

	//Test Case 3
	if _, err := Login("", ""); err == nil {
		t.Error("Gave an empty email and did not receive error!")
	}

	//Test Case 4
	if _, err := Login("realcheat@gmail.com", ""); err == nil {
		t.Error("Gave an empty email and did not receive error!")
	}

	//Test Case 5
	if _, err := Login("realcheat@gmail.com", "password123"); err == nil {
		t.Error("Test Case 5 failed")
	}
}

// Thoroughly test the Validate() function
func TestValidateToReturnFalse(t *testing.T) {
	var tst Account

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
}

/*func TestValidateToReturnTrue(t *testing.T) {

}*/
