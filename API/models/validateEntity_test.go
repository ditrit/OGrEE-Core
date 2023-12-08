package models

import (
	"encoding/json"
	"os"
	u "p3/utils"
	"regexp"
	"strings"
	"testing"
)

func TestValidateJsonSchemaExamples(t *testing.T) {
	// Test schemas examples
	testingEntities := []int{u.SITE, u.BLDG, u.ROOM, u.RACK, u.DEVICE, u.GROUP, u.BLDGTMPL, u.OBJTMPL, u.ROOMTMPL, u.TAG, u.LAYER}
	for _, entInt := range testingEntities {
		entStr := u.EntityToString(entInt)
		println("*** Testing " + entStr)
		var obj map[string]interface{}
		data, e := os.ReadFile("schemas/" + entStr + "_schema.json")
		if e != nil {
			t.Error(e.Error())
		}
		json.Unmarshal(data, &obj) // only one example per schema
		ok, err := validateJsonSchema(entInt, obj["examples"].([]interface{})[0].(map[string]interface{}))
		if !ok {
			t.Errorf("Error validating json schema: %s", err.Message)
		}
	}
}

func TestValidateJsonSchema(t *testing.T) {
	// Test test_data/OK json files
	testDataDir := "schemas/test_data/OK/"
	entries, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Error(err.Error())
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			t.Error("Not a JSON file")
		}
		testObjName := e.Name()[:len(e.Name())-5] // remove .json
		entStr := regexp.MustCompile(`[0-9]+`).ReplaceAllString(testObjName, "")
		entInt := u.EntityStrToInt(entStr)
		if entInt < 0 {
			t.Error("Unable to get entity from file name")
		}
		testObj := getMapFromJsonFile(testDataDir + e.Name())
		if testObj == nil {
			t.Error("Unable to convert json test file")
		}

		println("*** Testing " + testObjName)
		ok, err := validateJsonSchema(entInt, testObj)
		if !ok {
			t.Errorf("Error validating json schema: %s", err.Message)
		}
	}
}

func TestErrorValidateJsonSchema(t *testing.T) {
	// Test test_data/KO json files
	expectedErrors := map[string][]string{
		"site1":     {"missing properties: 'domain'", "/attributes/reservedColor does not match pattern"},
		"building1": {"missing properties: 'posXYUnit'", "/attributes/height expected string, but got number"},
		"room1":     {"additionalProperties 'banana' not allowed", "/attributes/axisOrientation value must be one of"},
		"rack1":     {"/attributes/posXYZ does not match pattern", "/attributes/heightUnit value must be one of"},
		"device1":   {"missing properties: 'template'", "/description expected array, but got string"},
		"group1":    {"/attributes missing properties: 'content'", "/name does not match pattern"},
		"obj_template5": {
			"/slug does not match pattern",
			"/attributes/vendor expected string, but got number",
			"/slots/0/elemOrient value must be one of ",
			"/slots/1/elemPos maximum 3 items required, but found 4 items",
			"/slots/1/elemSize minimum 3 items required, but found 2 items",
			"/slots/1/elemOrient value must be one of",
			"/slots/2 missing properties: 'elemOrient'",
			"/slots/2/labelPos value must be one of ",
			"/slots/3/color does not match pattern",
		},
		"obj_template4": {
			"allOf failed",
			"if-then failed",
			"/components/0/elemPos minimum 3 items required, but found 0 items",
			`/components/1/elemOrient value must be one of "horizontal", "vertical", ""`,
			"/components/1/color does not match pattern",
			"/components/3/labelPos value must be one of",
			"/slots/0/elemOrient value must be one of",
		},
		"bldg_template2": {
			"/sizeWDHm minimum 3 items required, but found 2 items",
			"/vertices/2 minimum 2 items required, but found 1 items",
			"/center minimum 2 items required, but found 0 items",
		},
		"room_template2": {
			"/tiles/0 missing properties: 'location'",
			"/axisOrientation value must be one of",
			"/separators/sep/type value must be one of",
			"/vertices/4 minimum 2 items required, but found 1 items",
			"/floorUnit value must be one of",
			"property 'tileAngle' is required, if 'vertices' property exists",
			"property 'center' is required, if 'vertices' property exists",
		},
	}

	testDataDir := "schemas/test_data/KO/"
	entries, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Error(err.Error())
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".json") {
			t.Error("Not a JSON file")
		}
		testObjName := e.Name()[:len(e.Name())-5] // remove .json
		entStr := regexp.MustCompile(`[0-9]+`).ReplaceAllString(testObjName, "")

		t.Run("TestErrorValidateJsonSchema_"+testObjName, func(t *testing.T) {
			entInt := u.EntityStrToInt(entStr)
			if entInt < 0 {
				t.Error("Unable to get entity from file name")
			}
			testObj := getMapFromJsonFile(testDataDir + e.Name())
			if testObj == nil {
				t.Error("Unable to convert json test file")
			}

			ok, err := validateJsonSchema(entInt, testObj)
			if ok {
				t.Errorf("Validated json schema that should have these errors: %v", expectedErrors[testObjName])
			} else {
				if len(err.Details) != len(expectedErrors[testObjName]) {
					t.Errorf("Validation errors do not correspond expected errors:\n%v\nGot:\n%v", expectedErrors[testObjName], err.Details)
				} else {
					for _, expected := range expectedErrors[testObjName] {
						if !contains(err.Details, expected) {
							t.Errorf("Expected error %v not contained in %v", expected, err.Details)
						}
					}
				}
			}
		})
	}
}

// helper functions
func contains(slice []string, elem string) bool {
	for _, e := range slice {
		if strings.Contains(e, elem) {
			return true
		}
	}

	return false
}

func getMapFromJsonFile(file string) map[string]interface{} {
	var obj map[string]interface{}
	data, e := os.ReadFile(file)
	if e != nil {
		return nil
	}
	json.Unmarshal(data, &obj)
	return obj
}
