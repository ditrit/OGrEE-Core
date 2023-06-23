package utils

import (
	"testing"
)

func TestMessageToReturnMSI(t *testing.T) {

	//Test Case 1
	testString := "Hello there"
	msi := Message(testString)
	if msi["message"] != testString {
		t.Error("Test Case 1 failed")
	}

	//Test Case 2
	testString = ""
	msi = Message(testString)
	if msi["message"] != testString {
		t.Error("Test Case 2 failed")
	}

}

func TestEntityStrToIntToReturnTrue(t *testing.T) {
	//Test Case 1
	testString := "site"
	ent := EntityStrToInt(testString)
	if ent != SITE {
		t.Error("Test Case 2 failed")
	}

	//Test Case 3
	testString = "building"
	ent = EntityStrToInt(testString)
	if ent != BLDG {
		t.Error("Test Case 3 failed")
	}

	//Test Case 4
	testString = "room"
	ent = EntityStrToInt(testString)
	if ent != ROOM {
		t.Error("Test Case 4 failed")
	}

	//Test Case 5
	testString = "rack"
	ent = EntityStrToInt(testString)
	if ent != RACK {
		t.Error("Test Case 5 failed")
	}

	//Test Case 6
	testString = "device"
	ent = EntityStrToInt(testString)
	if ent != DEVICE {
		t.Error("Test Case 6 failed")
	}

	//Test Case 9
	testString = "bldg"
	ent = EntityStrToInt(testString)
	if ent != BLDG {
		t.Error("Test Case 7 failed")
	}
}

func TestEntityToStringToReturnTrue(t *testing.T) {
	//Test Case 1
	testString := SITE
	ent := EntityToString(testString)
	if ent != "site" {
		t.Error("Test Case 2 failed")
	}

	//Test Case 3
	testString = BLDG
	ent = EntityToString(testString)
	if ent != "building" {
		t.Error("Test Case 3 failed")
	}

	//Test Case 4
	testString = ROOM
	ent = EntityToString(testString)
	if ent != "room" {
		t.Error("Test Case 4 failed")
	}

	//Test Case 5
	testString = RACK
	ent = EntityToString(testString)
	if ent != "rack" {
		t.Error("Test Case 5 failed")
	}

	//Test Case 6
	testString = DEVICE
	ent = EntityToString(testString)
	if ent != "device" {
		t.Error("Test Case 6 failed")
	}
}
