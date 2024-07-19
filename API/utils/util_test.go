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

func getEntityTestsList() []struct {
	name         string
	stringEntity string
	intEntity    int
} {
	return []struct {
		name         string
		stringEntity string
		intEntity    int
	}{
		{"SiteEntity", "site", SITE},
		{"BuildingEntity", "building", BLDG},
		{"RoomEntity", "room", ROOM},
		{"RackEntity", "rack", RACK},
		{"DeviceEntity", "device", DEVICE},
		{"AcEntity", "ac", AC},
		{"PanelEntity", "panel", PWRPNL},
		{"DomainEntity", "domain", DOMAIN},
		{"RoomTemplateEntity", "room_template", ROOMTMPL},
		{"ObjectTemplateEntity", "obj_template", OBJTMPL},
		{"BuildingTemplateEntity", "bldg_template", BLDGTMPL},
		{"CabinetEntity", "cabinet", CABINET},
		{"GroupEntity", "group", GROUP},
		{"CorridorEntity", "corridor", CORRIDOR},
		{"GenericEntity", "generic", GENERIC},
		{"TagEntity", "tag", TAG},
		{"LayerEntity", "layer", LAYER},
		{"InvalidEntity", "INVALID", -1},
	}

}

func TestEntityStrToIntToReturnTrue(t *testing.T) {
	tests := getEntityTestsList()
	// we add the short building name case
	tests = append(tests, struct {
		name         string
		stringEntity string
		intEntity    int
	}{"ShortBuildingEntity", "bldg", BLDG})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ent := EntityStrToInt(tt.stringEntity)
			assert.Equal(t, tt.intEntity, ent, "The entity \"%s\" should have the code %d", tt.stringEntity, tt.intEntity)
		})
	}
}

func TestEntityToStringToReturnTrue(t *testing.T) {
	tests := getEntityTestsList()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ent := EntityToString(tt.intEntity)
			assert.Equal(t, tt.stringEntity, ent, "The entity with code \"%d\" should have the string \"%s\"", tt.intEntity, tt.stringEntity)
		})
	}
}

func TestErrTypeToStatusCodeToReturnTrue(t *testing.T) {
	tests := []struct {
		name       string
		errorType  ErrType
		httpStatus int
	}{
		{"ErrForbidden", ErrForbidden, http.StatusForbidden},
		{"ErrUnauthorized", ErrUnauthorized, http.StatusUnauthorized},
		{"ErrNotFound", ErrNotFound, http.StatusNotFound},
		{"ErrDuplicate", ErrDuplicate, http.StatusBadRequest},
		{"ErrBadFormat", ErrBadFormat, http.StatusBadRequest},
		{"ErrDBError", ErrDBError, http.StatusInternalServerError},
		{"ErrInternal", ErrInternal, http.StatusInternalServerError},
		{"ErrInternal", -1, http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := ErrTypeToStatusCode(tt.errorType)
			assert.Equal(t, tt.httpStatus, code, "The errorType \"%d\" should have the status code \"%d\"", tt.errorType, tt.httpStatus)
		})
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

func TestConvertString(t *testing.T) {
	tests := []struct {
		name      string
		toconvert string
		expected  any
	}{
		{"StringToFloat", "1.2", 1.2},
		{"StringToFloatSlice", "[1.2, 1.3]", []float64{1.2, 1.3}},
		{"StringToStringSlice", "[hi, there]", []string{"hi", "there"}},
		{"StringToBoolean", "true", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			converted := ConvertString(tt.toconvert)
			assert.Equal(t, tt.expected, converted)
		})
	}
}
