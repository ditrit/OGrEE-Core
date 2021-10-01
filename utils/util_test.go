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

func TestEntityStrToIntToReturnTrue(t *testing.T) {
	//Test Case 1
	testString := "tenant"
	ent := EntityStrToInt(testString)
	if ent != TENANT {
		t.Error("Test Case 1 failed")
	}

	//Test Case 2
	testString = "site"
	ent = EntityStrToInt(testString)
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

	//Test Case 7
	testString = "subdevice"
	ent = EntityStrToInt(testString)
	if ent != SUBDEV {
		t.Error("Test Case 7 failed")
	}

	//Test Case 8
	testString = "subdevice1"
	ent = EntityStrToInt(testString)
	if ent != SUBDEV1 {
		t.Error("Test Case 8 failed")
	}

	//Test Case 9
	testString = "bldg"
	ent = EntityStrToInt(testString)
	if ent != BLDG {
		t.Error("Test Case 9 failed")
	}
}

func TestEntityToStringToReturnTrue(t *testing.T) {
	//Test Case 1
	testString := TENANT
	ent := EntityToString(testString)
	if ent != "tenant" {
		t.Error("Test Case 1 failed")
	}

	//Test Case 2
	testString = SITE
	ent = EntityToString(testString)
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

	//Test Case 7
	testString = SUBDEV
	ent = EntityToString(testString)
	if ent != "subdevice" {
		t.Error("Test Case 7 failed")
	}

	//Test Case 8
	testString = SUBDEV1
	ent = EntityToString(testString)
	if ent != "subdevice1" {
		t.Error("Test Case 8 failed")
	}
}
