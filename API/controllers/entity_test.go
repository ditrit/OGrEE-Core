package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"p3/test/e2e"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests domain bulk creation (/api/domains/bulk)
func TestCreateBulkDomains(t *testing.T) {
	// Test create two separate domains
	requestBody := []byte(`[
		{
			"name": "domain1",
			"parentId": "",
			"description": "Domain 1",
			"color": "ffffff"
		},
		{
			"name": "domain2",
			"parentId": "",
			"description": "Domain 2",
			"color": "ffffff"
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
			"color": "00ED00",
			"domains": [
				{
					"name": "subDomain1",
					"description": "subDomain 1",
					"color": "ffffff"
				}
			]
		},
		{
			"name": "domain4",
			"description": "Domain 4",
			"color": "00ED00",
			"domains": [
				{
					"name": "subDomain1",
					"description": "subDomain 1",
					"color": "00ED00"
				},
				{
					"name": "subDomain2",
					"description": "subDomain 2",
					"color": "ffffff"
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

	message, exists = response["domain4"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created domain", message)

	message, exists = response["domain4.subDomain1"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created domain", message)

	message, exists = response["domain4.subDomain2"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully created domain", message)
}

func TestCreateBulkDomainWithDuplicateError(t *testing.T) {
	// Test try to create a domain that already exists
	requestBody := []byte(`[
		{
			"name": "domain3",
			"description": "Domain 3",
			"color": "00ED00"
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

// Tests delete subdomains (/api/objects)
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

// Tests handle complex filters (/api/objects/search)
func TestComplexFilterSearch(t *testing.T) {
	// Test get subdomains of domain4 with color 00ED00
	requestBody := []byte(`{
		"$and": [
			{
				"id": {
					"$regex": "^domain4[.].*"
				}
			},
			{
				"attributes.color": "00ED00"
			}
		]
	}`)

	recorder := e2e.MakeRequest("POST", "/api/objects/search", requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully processed request", message)

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 1, len(data))

	domain := data[0].(map[string]interface{})
	id, exists := domain["id"].(string)
	assert.True(t, exists)
	assert.Equal(t, "domain4.subDomain1", id)
}

// Tests handle delete with complex filters (/api/objects/search)
func TestComplexFilterDelete(t *testing.T) {
	// Test delete subdomains of domain4 with color 00ED00
	requestBody := []byte(`{
		"$and": [
			{
				"id": {
					"$regex": "^domain4[.].*"
				}
			},
			{
				"attributes.color": "00ED00"
			}
		]
	}`)

	recorder := e2e.MakeRequest("DELETE", "/api/objects/search", requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully deleted objects", message)

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 1, len(data))

	domain := data[0].(map[string]interface{})
	id, exists := domain["id"].(string)
	assert.True(t, exists)
	assert.Equal(t, "domain4.subDomain1", id)
}

// Tests get different entities
func TestGetDomainEntity(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/domains", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully got domains", message)

	// we have multiple domains
	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)

	objects, exists := data["objects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, true, len(objects) > 0) // we have domains created in this file and others

	// domain3 exists but domain3.subDomain1 does not
	hasDomain3 := slices.ContainsFunc(objects, func(value interface{}) bool {
		domain := value.(map[string]interface{})
		return domain["id"].(string) == "domain3"
	})
	assert.Equal(t, true, hasDomain3)

	hasDomain3Subdomain1 := slices.ContainsFunc(objects, func(value interface{}) bool {
		domain := value.(map[string]interface{})
		return domain["id"].(string) == "domain3.subDomain1"
	})
	assert.Equal(t, false, hasDomain3Subdomain1)
}

func TestGetBuildingsEntity(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/buildings", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully got buildings", message)

	// we have multiple buildings
	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)

	objects, exists := data["objects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, true, len(objects) > 0)
}

func TestGetUnknownEntity(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/unknown", nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetDomainEntitiesFilteredByColor(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/domains?color=00ED00", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully got query for domain", message)

	// we have multiple domains
	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)

	objects, exists := data["objects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 2, len(objects)) // domain3 and domain4
}
