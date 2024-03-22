package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"p3/test/e2e"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBulkDomains(t *testing.T) {
	// Test create two separate domains
	requestBody := []byte(`[
		{
			"name": "domain1",
			"parentId": "",
			"description": "Domain 1",
			"attributes": {
				"color": "ffffff"
			}
		},
		{
			"name": "domain2",
			"parentId": "",
			"description": "Domain 2",
			"attributes": {
				"color": "ffffff"
			}
		}
	]`)

	recorder := e2e.MakeRequest("POST", "/api/domains/bulk", requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["domain1"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created domain", message)

	message, exists = response["domain2"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created domain", message)
}

func TestCreateBulkDomainWithSubdomains(t *testing.T) {
	// Test create one domaain with a sub domain
	requestBody := []byte(`[
		{
			"name": "domain3",
			"description": "Domain 3",
			"attributes": {
				"color": "00ED00"
			},
			"domains": [
				{
					"name": "subDomain1",
					"description": "subDomain 1",
					"attributes": {
						"color": "ffffff"
					}
				}
			]
		}
	]`)

	recorder := e2e.MakeRequest("POST", "/api/domains/bulk", requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["domain3"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created domain", message)

	message, exists = response["domain3.subDomain1"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created domain", message)
}

func TestCreateBulkDomainWithDuplicateError(t *testing.T) {
	// Test try to create a domain that already exists
	requestBody := []byte(`[
		{
			"name": "domain3",
			"description": "Domain 3",
			"attributes": {
				"color": "00ED00"
			}
		}
	]`)

	recorder := e2e.MakeRequest("POST", "/api/domains/bulk", requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["domain3"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Error while creating domain: Duplicates not allowed", message)
}

func TestDeleteSubdomains(t *testing.T) {
	// Test delete subdomain using a pattern
	params, _ := url.ParseQuery("id=domain3.*")

	recorder := e2e.MakeRequest("DELETE", "/api/objects?"+params.Encode(), nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully deleted objects", message)

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 1, len(data))
	deletedDomain := data[0].(map[string]interface{})
	id, exists := deletedDomain["id"].(string)
	assert.True(t, exists)
	assert.Equal(t, "domain3.subDomain1", id)
}
