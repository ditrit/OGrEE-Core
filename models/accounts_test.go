package models

import (
	"testing"
)

func TestLoginToReturnFalse(t *testing.T) {
	//fmt.Println(res)
	//fmt.Println(reflect.TypeOf(res["status"]))

	//Test Case 1
	if res, _ := Login("throwaway", "password"); res["status"] != false {
		t.Error("Gave a false login request and did not receive error!")
	}

	//Test Case 2
	if res, _ := Login("", "password"); res["status"] != false {
		t.Error("Gave an empty email and did not receive error!")
	}

	//Test Case 3
	if res, _ := Login("", ""); res["status"] != false {
		t.Error("Gave an empty email and did not receive error!")
	}

	//Test Case 4
	if res, _ := Login("realcheat@gmail.com", ""); res["status"] != false {
		t.Error("Gave an empty email and did not receive error!")
	}

	//Test Case 5
	if res, _ := Login("realcheat@gmail.com", "password123"); res["status"] != false {
		t.Error("Test Case 5 failed")
	}
}

// Thoroughly test the Validate() function
func TestValidateToReturnFalse(t *testing.T) {
	var tst Account

	//Case 1
	if msi, _ := tst.Validate(); msi["message"] != "A valid email address is required" {
		t.Error("Gave empty email, but did not receive corresponding error")
	}

	//Case 2
	tst.Email = "realcheat"
	if msi, _ := tst.Validate(); msi["message"] != "A valid email address is required" {
		t.Error("Gave bad email, but did not receive corresponding error")
	}

	//Case 3
	tst.Email = "@"
	if msi, _ := tst.Validate(); msi["status"] != false {
		t.Error("Gave '@' as email, but did not receive corresponding error")
	}

	//Case 4
	tst.Email = "@"
	tst.Password = "secret123"
	if msi, _ := tst.Validate(); msi["status"] != false {
		t.Error(`Gave '@' as email and valid password, \
					but did not receive corresponding error`)
	}

	//Case 5
	tst.Email = "realcheat@"
	tst.Password = "secret123"
	if _, val := tst.Validate(); val != false {
		t.Error("Test Case 5 failed!")
	}

	//Case 6
	tst.Email = "realcheat@orness.com"
	tst.Password = ""
	if _, val := tst.Validate(); val != false {
		t.Error("Test Case 6 failed!")
	}
}

/*func TestValidateToReturnTrue(t *testing.T) {

}*/
