package utils

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
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

	//Test Case 7
	testString = -1
	ent = EntityToString(testString)
	if ent != "INVALID" {
		t.Error("Test Case 7 failed")
	}
}

func TestErrTypeToStatusCodeToReturnTrue(t *testing.T) {
	//Test Case 1
	testString := ErrForbidden
	ent := ErrTypeToStatusCode(testString)
	if ent != http.StatusForbidden {
		t.Error("Test Case 1 failed")
	}

	//Test Case 2
	testString = ErrUnauthorized
	ent = ErrTypeToStatusCode(testString)
	if ent != http.StatusUnauthorized {
		t.Error("Test Case 2 failed")
	}

	//Test Case 3
	testString = ErrNotFound
	ent = ErrTypeToStatusCode(testString)
	if ent != http.StatusNotFound {
		t.Error("Test Case 3 failed")
	}

	//Test Case 4
	testString = ErrDuplicate
	ent = ErrTypeToStatusCode(testString)
	if ent != http.StatusBadRequest {
		t.Error("Test Case 4 failed")
	}

	//Test Case 5
	testString = ErrBadFormat
	ent = ErrTypeToStatusCode(testString)
	if ent != http.StatusBadRequest {
		t.Error("Test Case 5 failed")
	}

	//Test Case 6
	testString = ErrDBError
	ent = ErrTypeToStatusCode(testString)
	if ent != http.StatusInternalServerError {
		t.Error("Test Case 6 failed")
	}

	//Test Case 7
	testString = ErrInternal
	ent = ErrTypeToStatusCode(testString)
	if ent != http.StatusInternalServerError {
		t.Error("Test Case 7 failed")
	}

	//Test Case 8
	testString = -1
	ent = ErrTypeToStatusCode(testString)
	if ent != http.StatusInternalServerError {
		t.Error("Test Case 8 failed")
	}
}

func TestStrSliceContains(t *testing.T) {
	assert.True(t, StrSliceContains([]string{"hello", "world"}, "hello"))
	assert.False(t, StrSliceContains([]string{"hello", "world"}, "bye"))
	assert.False(t, StrSliceContains([]string{}, "bye"))
}

func TestFormatNotifyData(t *testing.T) {
	//Test Case 1
	message := FormatNotifyData("create", "room", nil)
	var messageJson map[string]any
	json.Unmarshal([]byte(message), &messageJson)
	if messageJson["type"] != "create" || messageJson["data"] != nil {
		t.Error("Test Case 1 failed")
	}

	//Test Case 2
	message = FormatNotifyData("create", "tag", nil)
	json.Unmarshal([]byte(message), &messageJson)
	if messageJson["type"].(string) != "create-tag" || messageJson["data"] != nil {
		t.Error("Test Case 2 failed")
	}

	//Test Case 3
	message = FormatNotifyData("create", "layer", nil)
	json.Unmarshal([]byte(message), &messageJson)
	if messageJson["type"].(string) != "create-layer" || messageJson["data"] != nil {
		t.Error("Test Case 3 failed")
	}

	//Test Case 4
	message = FormatNotifyData("create", "room", map[string]string{"extra": "data"})
	json.Unmarshal([]byte(message), &messageJson)
	data, exists := messageJson["data"].(map[string]interface{})
	if messageJson["type"].(string) != "create" || !exists || data["extra"].(string) != "data" {
		t.Error("Test Case 4 failed")
	}
}
