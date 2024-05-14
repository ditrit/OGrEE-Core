package main

import (
	"encoding/json"
	"net/http"
	"os"
	"p3/test/e2e"
	test_utils "p3/test/utils"
	u "p3/utils"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateLoginAccount(t *testing.T) {
	// Test create new account
	requestBody := []byte(`{
		"email": "test@test.com",
		"password": "pass123secret",
		"roles":{"*":"manager"}
	}`)
	usersEndpoint := test_utils.GetEndpoint("users")
	recorder := e2e.MakeRequest("POST", usersEndpoint, requestBody)
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	_, exists := response["account"].(map[string]interface{})["token"].(string)
	assert.True(t, exists)

	// Test recreate existing account
	recorder = e2e.MakeRequest("POST", usersEndpoint, requestBody)
	println(recorder.Body.String())
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	// Test login
	recorder = e2e.MakeRequest("POST", test_utils.GetEndpoint("login"), requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	json.Unmarshal(recorder.Body.Bytes(), &response)
	_, exists = response["account"].(map[string]interface{})["token"].(string)
	assert.True(t, exists)
}

func TestObjects(t *testing.T) {
	var response map[string]interface{}
	var parentId string

	// Create objects from schema examples
	for _, entInt := range []int{u.DOMAIN, u.SITE, u.BLDG, u.ROOM, u.CORRIDOR, u.RACK, u.DEVICE} {
		// Get object from schema
		entStr := u.EntityToString(entInt)
		data, _ := os.ReadFile("models/schemas/" + entStr + "_schema.json")
		var obj map[string]interface{}
		json.Unmarshal(data, &obj)
		obj = obj["examples"].([]interface{})[0].(map[string]interface{})
		if entInt != u.SITE && entInt != u.DOMAIN {
			// Add parentId
			obj["parentId"] = parentId
		}
		data, _ = json.Marshal(obj)

		endpoint := "/api/" + entStr + "s"
		// Create (POST)
		recorder := e2e.MakeRequest("POST", endpoint, data)

		assert.Equal(t, http.StatusCreated, recorder.Code)
		json.Unmarshal(recorder.Body.Bytes(), &response)
		id, exists := response["data"].(map[string]interface{})["id"].(string)
		assert.True(t, exists)

		instanceEndpoint := endpoint + "/" + id
		// Verify create with GET
		recorder = e2e.MakeRequest("GET", instanceEndpoint, nil)
		assert.Equal(t, http.StatusOK, recorder.Code)
		var responseGET map[string]interface{}
		json.Unmarshal(recorder.Body.Bytes(), &responseGET)
		delete(responseGET, "message")
		delete(response, "message")
		assert.True(t, reflect.DeepEqual(response, responseGET))

		// Update with PUT
		oldName := obj["name"].(string)
		obj["name"] = entStr + "Test"
		data, _ = json.Marshal(obj)
		recorder = e2e.MakeRequest("PUT", instanceEndpoint, data)
		assert.Equal(t, http.StatusOK, recorder.Code)

		// Verify it
		id = strings.Replace(id, oldName, obj["name"].(string), 1)
		instanceEndpoint = endpoint + "/" + id
		recorder = e2e.MakeRequest("GET", instanceEndpoint, nil)
		assert.Equal(t, http.StatusOK, recorder.Code)
		if entInt != u.CORRIDOR { // corridor has no child, but a rack brother
			parentId = id
		}
	}

	// Try to patch site name
	hierarchyName := "TESTPATCH"
	requestBody := []byte(`{
		"name": "` + hierarchyName + `"
	}`)
	siteEndpoint := test_utils.GetEndpoint("entityInstance", "sites", "siteTest")
	recorder := e2e.MakeRequest("PATCH", siteEndpoint, requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	// Verify the whole tree
	siteEndpoint = test_utils.GetEndpoint("entityInstance", "sites", hierarchyName)
	recorder = e2e.MakeRequest("GET", siteEndpoint+"/all", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)
	json.Unmarshal(recorder.Body.Bytes(), &response)
	response = response["data"].(map[string]interface{})
	for _, entInt := range []int{u.BLDG, u.ROOM, u.RACK, u.DEVICE} {
		entStr := u.EntityToString(entInt)
		println(entStr)
		child := response["children"].([]interface{})[0].(map[string]interface{})
		hierarchyName = hierarchyName + u.HN_DELIMETER + entStr + "Test"
		assert.Equal(t, hierarchyName, child["id"].(string))
		if entInt != u.DEVICE {
			response = child
		}
	}

	// Delete everything
	recorder = e2e.MakeRequest("DELETE", siteEndpoint, nil)
	assert.Equal(t, http.StatusNoContent, recorder.Code)
	recorder = e2e.MakeRequest("GET", test_utils.GetEndpoint("hierarchy"), nil)
	assert.Equal(t, http.StatusOK, recorder.Code)
	json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Nil(t, response["data"].(map[string]interface{})["tree"].(map[string]interface{})["physical"].(map[string]interface{})["TESTPATCH"])
}
