package utils

import (
	"testing"
)

func TestMessageToReturnTrueMSI(t *testing.T) {

	//Test Case 1
	testString := "Hello there"
	msi := Message(true, testString)
	if msi["status"] != true || msi["message"] != testString {
		t.Error("Test Case 1 failed")
	}

	//Test Case 2
	testString = ""
	msi = Message(true, testString)
	if msi["status"] != true || msi["message"] != testString {
		t.Error("Test Case 2 failed")
	}

}

func TestMessageToReturnFalseMSI(t *testing.T) {
	//Test Case 1
	testString := "Hello there"
	msi := Message(false, testString)
	if msi["status"] != false || msi["message"] != testString {
		t.Error("Test Case 1 failed")
	}

	//Test Case 2
	testString = ""
	msi = Message(false, testString)
	if msi["status"] != false || msi["message"] != testString {
		t.Error("Test Case 2 failed")
	}
}
