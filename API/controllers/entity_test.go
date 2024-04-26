package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"p3/models"
	"p3/test/e2e"
	"p3/test/integration"
	test_utils "p3/test/utils"
	"p3/utils"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func init() {
	integration.RequireCreateSite("site-no-temperature")
	integration.RequireCreateBuilding("site-no-temperature", "building-1")
	integration.RequireCreateBuilding("site-no-temperature", "building-2")
	integration.RequireCreateBuilding("site-no-temperature", "building-3")
	integration.RequireCreateRoom("site-no-temperature.building-1", "room-1")
	integration.RequireCreateRoom("site-no-temperature.building-1", "room-2")
	integration.RequireCreateRoom("site-no-temperature.building-2", "room-1")
	integration.RequireCreateRack("site-no-temperature.building-1.room-1", "rack-1")
	integration.RequireCreateRack("site-no-temperature.building-1.room-1", "rack-2")
	integration.RequireCreateDevice("site-no-temperature.building-1.room-1.rack-2", "device-1")
	integration.RequireCreateRack("site-no-temperature.building-1.room-2", "rack-1")
	integration.RequireCreateSite("site-with-temperature")
	integration.RequireCreateBuilding("site-with-temperature", "building-3")

	temperatureData := map[string]any{
		"attributes": map[string]any{
			"temperatureUnit": "30",
		},
	}
	models.UpdateObject("site", "site-with-temperature", temperatureData, true, integration.ManagerUserRoles, false)
	layer := map[string]any{
		"slug":          "racks-layer",
		"filter":        "category=rack",
		"applicability": "site-no-temperature.building-1.room-1",
	}
	models.CreateEntity(utils.LAYER, layer, integration.ManagerUserRoles)
	layer2 := map[string]any{
		"slug":          "racks-1-layer",
		"filter":        "category=rack & name=rack-1",
		"applicability": "site-no-temperature.building-1.room-*",
	}
	models.CreateEntity(utils.LAYER, layer2, integration.ManagerUserRoles)
}

func getTestEntity(entity string) map[string]any {
	// returns an entity to use in tests
	if entity == "room" {
		return map[string]any{
			"attributes": map[string]any{
				"floorUnit":       "t",
				"height":          "2.8",
				"heightUnit":      "m",
				"axisOrientation": "+x+y",
				"rotation":        "-90",
				"posXY":           "[0, 0]",
				"posXYUnit":       "m",
				"size":            "[-13, -2.9]",
				"sizeUnit":        "m",
				"template":        "",
			},
			"category":    "room",
			"description": "room",
			"domain":      "domain",
			"name":        "roomA",
			"parentId":    "site-no-temperature.building-1",
		}
	}
	return nil
}

// Tests with invalid body
func TestEntityRequestsWithInvalidBody(t *testing.T) {
	tests := []struct {
		name          string
		requestMethod string
		endpoint      string
	}{
		{"CreateEntity", "POST", test_utils.GetEndpoint("entity", "sites")},
		{"CreateBulkDomains", "POST", test_utils.GetEndpoint("domainsBulk")},
		{"ComplexFilterSearch", "POST", test_utils.GetEndpoint("complexFilterSearch")},
		{"validateEntity", "POST", test_utils.GetEndpoint("validateEntity", "rooms")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e2e.TestInvalidBody(t, tt.requestMethod, tt.endpoint, "Error while decoding request body")
		})
	}
}

// Tests domain bulk creation (/api/domains/bulk)
func TestCreateBulkDomains(t *testing.T) {
	// Test create two separate domains
	domain1 := "domain1"
	domain2 := "domain2"
	domain3 := "domain3"
	subDomain1 := "subDomain1"
	subDomain2 := "subDomain2"
	domain4 := "domain4"
	domains := []string{domain3 + "." + subDomain1, domain4 + "." + subDomain1, domain4 + "." + subDomain2, domain1, domain2, domain3, domain4}

	requestBody := []byte(`[
		{
			"name": "` + domain1 + `",
			"parentId": "",
			"color": "ffffff"
		},
		{
			"name": "` + domain2 + `",
			"parentId": "",
			"description": "Domain 2"
		},
		{
			"name": "` + domain3 + `",
			"description": "Domain 3",
			"color": "00ED00",
			"domains": [
				{
					"name": "` + subDomain1 + `",
					"description": "subDomain 1",
					"color": "ffffff"
				}
			]
		},
		{
			"name": "` + domain4 + `",
			"description": "Domain 4",
			"color": "00ED00",
			"domains": [
				{
					"name": "` + subDomain1 + `",
					"description": "subDomain 1",
					"color": "00ED00"
				},
				{
					"name": "` + subDomain2 + `",
					"description": "subDomain 2",
					"color": "ffffff"
				}
			]
		}
	]`)

	recorder := e2e.MakeRequest("POST", test_utils.GetEndpoint("domainsBulk"), requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.Len(t, response, 7)
	for _, domain := range domains {
		message, exists := response[domain].(string)
		assert.True(t, exists)
		assert.Equal(t, "successfully created domain", message)
	}

	// we delete the created domains
	for _, domain := range domains {
		models.DeleteObject(utils.EntityToString(utils.DOMAIN), domain, integration.ManagerUserRoles)
	}
}

func TestCreateBulkDomainWithDuplicateError(t *testing.T) {
	// Test try to create a domain that already exists
	test_utils.CreateTestDomain(t, "temporaryDomain", "", "")
	requestBody := []byte(`[
		{
			"name": "temporaryDomain",
			"description": "temporaryDomain",
			"color": "00ED00"
		}
	]`)

	recorder := e2e.MakeRequest("POST", test_utils.GetEndpoint("domainsBulk"), requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	message, exists := response["temporaryDomain"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Error while creating domain: Duplicates not allowed", message)
}

// Tests delete subdomains (/api/objects)
func TestDeleteSubdomains(t *testing.T) {
	// Test delete subdomain using a pattern
	test_utils.CreateTestDomain(t, "temporaryFatherDomain", "", "")
	test_utils.CreateTestDomain(t, "temporaryChildDomain", "temporaryFatherDomain", "")
	params, _ := url.ParseQuery("id=temporaryFatherDomain.*")
	endpoint := test_utils.GetEndpoint("getObject") + "?" + params.Encode()
	response := test_utils.ValidateManagedRequest(t, "DELETE", endpoint, nil, http.StatusOK, "successfully deleted objects")

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 1, len(data))
	deletedDomain := data[0].(map[string]interface{})
	id, exists := deletedDomain["id"].(string)
	assert.True(t, exists)
	assert.Equal(t, "temporaryFatherDomain.temporaryChildDomain", id)
}

// Tests handle complex filters (/api/objects/search)
func TestComplexFilterWithNoFilterInput(t *testing.T) {
	requestBody := []byte(`{}`)

	message := "Invalid body format: must contain a filter key with a not empty string as value"
	test_utils.ValidateManagedRequest(t, "POST", test_utils.GetEndpoint("complexFilterSearch"), requestBody, http.StatusBadRequest, message)
}

func TestComplexFilterSearch(t *testing.T) {
	// Test get subdomains with color 00ED00
	test_utils.CreateTestDomain(t, "temporaryFatherDomain", "", "")
	test_utils.CreateTestDomain(t, "temporaryChildDomain", "temporaryFatherDomain", "00ED00")
	test_utils.CreateTestDomain(t, "temporarySecondChildDomain", "temporaryFatherDomain", "ffffff")
	requestBody := []byte(`{
		"filter": "id=temporaryFatherDomain.* & color=00ED00"
	}`)

	message := "successfully processed request"
	response := test_utils.ValidateManagedRequest(t, "POST", test_utils.GetEndpoint("complexFilterSearch"), requestBody, http.StatusOK, message)

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 1, len(data))

	domain := data[0].(map[string]interface{})
	id, exists := domain["id"].(string)
	assert.True(t, exists)
	assert.Equal(t, "temporaryFatherDomain.temporaryChildDomain", id)
}

func TestComplexFilterSearchWithDateFilter(t *testing.T) {
	// Test get subdomains with color 00ED00 and different startDate
	test_utils.CreateTestDomain(t, "temporaryFatherDomain", "", "")
	test_utils.CreateTestDomain(t, "temporaryChildDomain", "temporaryFatherDomain", "00ED00")
	test_utils.CreateTestDomain(t, "temporarySecondChildDomain", "temporaryFatherDomain", "ffffff")
	requestBody := []byte(`{
		"filter": "id=temporaryFatherDomain.* & color=00ED00"
	}`)

	yesterday := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	message := "successfully processed request"
	baseEndpoint := test_utils.GetEndpoint("complexFilterSearch")

	tests := []struct {
		name         string
		queryParams  string
		resultLenght int
	}{
		{"StartDateYesterday", "?startDate=" + yesterday, 1},
		{"StartDateTomorrow", "?startDate=" + tomorrow, 0},
		{"EndDateYesterday", "?endDate=" + yesterday, 0},
		{"EndDateTomorrow", "?endDate=" + tomorrow, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := test_utils.ValidateManagedRequest(t, "POST", baseEndpoint+tt.queryParams, requestBody, http.StatusOK, message)
			data, exists := response["data"].([]interface{})
			assert.True(t, exists)
			assert.Equal(t, tt.resultLenght, len(data))
		})
	}
}

// Tests handle delete with complex filters (/api/objects/search)
func TestComplexFilterDelete(t *testing.T) {
	// Test delete subdomains with color 00ED00
	test_utils.CreateTestDomain(t, "temporaryFatherDomain", "", "")
	test_utils.CreateTestDomain(t, "temporaryChildDomain", "temporaryFatherDomain", "00ED00")
	test_utils.CreateTestDomain(t, "temporarySecondChildDomain", "temporaryFatherDomain", "ffffff")
	requestBody := []byte(`{
		"filter": "id=temporaryFatherDomain.* & color=00ED00"
	}`)

	message := "successfully deleted objects"
	response := test_utils.ValidateManagedRequest(t, "DELETE", test_utils.GetEndpoint("complexFilterSearch"), requestBody, http.StatusOK, message)

	data, exists := response["data"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 1, len(data))

	domain := data[0].(map[string]interface{})
	id, exists := domain["id"].(string)
	assert.True(t, exists)
	assert.Equal(t, "temporaryFatherDomain.temporaryChildDomain", id)
}

// Tests get different entities
func TestGetDomainEntity(t *testing.T) {
	test_utils.CreateTestDomain(t, "temporaryDomain", "", "")
	response := test_utils.ValidateManagedRequest(t, "GET", test_utils.GetEndpoint("domains"), nil, http.StatusOK, "successfully got domains")

	// we have multiple domains
	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)

	objects, exists := data["objects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, true, len(objects) > 0) // we have domains created in this file and others

	domainExists := slices.ContainsFunc(objects, func(value interface{}) bool {
		domain := value.(map[string]interface{})
		return domain["id"].(string) == "temporaryDomain"
	})
	assert.Equal(t, true, domainExists)
}

func TestGetBuildingsEntity(t *testing.T) {
	response := test_utils.ValidateManagedRequest(t, "GET", test_utils.GetEndpoint("entity", "buildings"), nil, http.StatusOK, "successfully got buildings")

	// we have multiple buildings
	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)

	objects, exists := data["objects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, true, len(objects) > 0)
}

func TestGetUnknownEntity(t *testing.T) {
	recorder := e2e.MakeRequest("GET", test_utils.GetEndpoint("entity", "unknown"), nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetDomainEntitiesFilteredByColor(t *testing.T) {
	test_utils.CreateTestDomain(t, "temporaryDomain1", "", "ffff01")
	test_utils.CreateTestDomain(t, "temporaryDomain2", "", "00ED00")
	test_utils.CreateTestDomain(t, "temporaryDomain3", "", "00ED00")

	endpoint := test_utils.GetEndpoint("domains") + "?color=00ED00"
	response := test_utils.ValidateManagedRequest(t, "GET", endpoint, nil, http.StatusOK, "successfully got query for domain")

	// we have multiple domains
	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)

	objects, exists := data["objects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 2, len(objects)) // temporaryDomain1 and temporaryDomain3
}

// Test get temperature unit
func TestGetTemperatureForDomain(t *testing.T) {
	test_utils.CreateTestDomain(t, "temporaryDomain", "", "")
	endpoint := test_utils.GetEndpoint("tempunits", "temporaryDomain")
	test_utils.ValidateManagedRequest(t, "GET", endpoint, nil, http.StatusNotFound, "Could not find parent site for given object")
}

func TestGetTemperatureForParentWithNoTemperature(t *testing.T) {
	endpoint := test_utils.GetEndpoint("tempunits", "site-no-temperature.building-1")
	test_utils.ValidateManagedRequest(t, "GET", endpoint, nil, http.StatusNotFound, "Parent site has no temperatureUnit in attributes")
}

func TestGetTemperature(t *testing.T) {
	endpoint := test_utils.GetEndpoint("tempunits", "site-with-temperature.building-3")
	response := test_utils.ValidateManagedRequest(t, "GET", endpoint, nil, http.StatusOK, "successfully got temperatureUnit from object's parent site")

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	temperatureUnit, exists := data["temperatureUnit"].(string)
	assert.True(t, exists)
	assert.Equal(t, "30", temperatureUnit)
}

// Tests get subentities
func TestErrorGetRoomsBuildingsInvalidHierarchy(t *testing.T) {
	endpoint := test_utils.GetEndpoint("entityAncestors", "rooms", "site-no-temperature.building-1.room-1", "buildings")
	message := "Invalid set of entities in URL: first entity should be parent of the second entity"
	test_utils.ValidateManagedRequest(t, "GET", endpoint, nil, http.StatusBadRequest, message)
}

func TestErrorGetSiteRoomsUnknownEntity(t *testing.T) {
	endpoint := test_utils.GetEndpoint("entityAncestors", "sites", "unknown", "rooms")
	message := "Nothing matches this request"
	test_utils.ValidateManagedRequest(t, "GET", endpoint, nil, http.StatusNotFound, message)
}

func TestGetSitesRooms(t *testing.T) {
	endpoint := test_utils.GetEndpoint("entityAncestors", "sites", "site-no-temperature", "rooms")
	message := "successfully got object"
	response := test_utils.ValidateManagedRequest(t, "GET", endpoint, nil, http.StatusOK, message)

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	objects, exists := data["objects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 3, len(objects))

	areRooms := true
	for _, element := range objects {
		if element.(map[string]interface{})["category"] != "room" {
			areRooms = false
			break
		}
	}
	assert.True(t, areRooms)
}

func TestGetHierarchyAttributes(t *testing.T) {
	test_utils.CreateTestDomain(t, "temporaryDomain", "", "ffff01")
	endpoint := test_utils.GetEndpoint("hierarchyAttributes")
	message := "successfully got attrs hierarchy"
	response := test_utils.ValidateManagedRequest(t, "GET", endpoint, nil, http.StatusOK, message)

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	keys := make([]int, len(data))
	assert.True(t, len(keys) > 0)

	// we test the color attribute is present for temporaryDomain
	domain, exists := data["temporaryDomain"].(map[string]interface{})
	assert.True(t, exists)
	color, exists := domain["color"].(string)
	assert.True(t, exists)
	assert.Equal(t, "ffff01", color)
}

// Tests link and unlink entity
func TestLinkUnlinkRoomss(t *testing.T) {
	// We create a temporary room (and its parents) to unlink it and link it to its parent and delete it at the end of the test
	strayName := "StrayRoom"
	parentId := "temporarySite.temporaryBuilding"
	roomName := "temporaryRoom"
	test_utils.CreateTestPhysicalEntity(t, utils.ROOM, roomName, parentId, true)

	unlinkEndpoint := test_utils.GetEndpoint("entityUnlink", "rooms", parentId+"."+roomName)
	linkEndpoint := test_utils.GetEndpoint("entityLink", "stray-objects", "StrayRoom")
	roomEndpoint := test_utils.GetEndpoint("entityInstance", "rooms", parentId+"."+roomName)
	strayObjectEndpoint := test_utils.GetEndpoint("entityInstance", "stray-objects", strayName)
	tests := []struct {
		name        string
		isUnlink    bool
		isSuccess   bool
		requestBody []byte
		statusCode  int
		message     string
	}{
		{"UnlinkWithNotAllowedAttributes", true, false, []byte(`{"name": "` + strayName + `","other": "other"}`), http.StatusBadRequest, "Body must be empty or only contain valid name"},
		{"UnlinkSuccess", true, true, []byte(`{"name": "` + strayName + `"}`), http.StatusOK, "successfully unlinked"},
		{"LinkWithoutParentId", false, false, []byte(`{"name": "` + roomName + `"}`), http.StatusBadRequest, "Error while decoding request body: must contain parentId"},
		{"LinkSuccess", false, true, []byte(`{"parentId": "` + parentId + `", "name": "` + roomName + `"}`), http.StatusOK, "successfully linked"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var patchEndpoint string
			var deletedInstanceEndpoint string
			var changedInstanceEndpoint string
			var changedInstanceMessage string
			var entitId string
			if tt.isUnlink {
				patchEndpoint = unlinkEndpoint
				deletedInstanceEndpoint = roomEndpoint
				changedInstanceEndpoint = strayObjectEndpoint
				changedInstanceMessage = "successfully got stray_object"
				entitId = strayName
			} else {
				patchEndpoint = linkEndpoint
				deletedInstanceEndpoint = strayObjectEndpoint
				changedInstanceEndpoint = roomEndpoint
				changedInstanceMessage = "successfully got room"
				entitId = parentId + "." + roomName
			}

			test_utils.ValidateManagedRequest(t, "PATCH", patchEndpoint, tt.requestBody, tt.statusCode, tt.message)

			if tt.isSuccess {
				// We verify the old entity does not exist
				test_utils.ValidateManagedRequest(t, "GET", deletedInstanceEndpoint, nil, http.StatusNotFound, "Nothing matches this request")

				// We verify the new entity exists
				response := test_utils.ValidateManagedRequest(t, "GET", changedInstanceEndpoint, nil, http.StatusOK, changedInstanceMessage)
				data, exists := response["data"].(map[string]interface{})
				assert.True(t, exists)
				id := data["id"].(string)
				assert.Equal(t, entitId, id)
			}
		})
	}
}

// Tests entity validation
func TestValidateNonExistentEntity(t *testing.T) {
	requestBody := []byte(`{}`)

	recorder := e2e.MakeRequest("POST", test_utils.GetEndpoint("validateEntity", "invalid"), requestBody)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestValidateEntityWithoutAttributes(t *testing.T) {
	requestBody := []byte(`{
		"category": "room",
		"description": "room",
		"domain": "domain1",
		"name": "roomA",
		"parentId": "site-no-temperature.building-1"
	}`)

	endpoint := test_utils.GetEndpoint("validateEntity", "rooms")
	expectedMessage := "JSON body doesn't validate with the expected JSON schema"
	test_utils.ValidateManagedRequest(t, "POST", endpoint, requestBody, http.StatusBadRequest, expectedMessage)
}

func TestValidateEntity(t *testing.T) {
	test_utils.CreateTestDomain(t, "temporaryDomain", "", "")
	room := getTestEntity("room")
	endpoint := test_utils.GetEndpoint("validateEntity", "rooms")
	tests := []struct {
		name       string
		domain     string
		statusCode int
		message    string
	}{
		{"NonExistentDomain", "invalid", http.StatusNotFound, "Domain not found: invalid"},
		{"InvalidDomain", "temporaryDomain", http.StatusBadRequest, "Object domain is not equal or child of parent's domain"},
		{"ValidRoomEntity", integration.TestDBName, http.StatusOK, "This object can be created"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room["domain"] = tt.domain
			requestBody, _ := json.Marshal(room)
			test_utils.ValidateManagedRequest(t, "POST", endpoint, requestBody, tt.statusCode, tt.message)
		})
	}
}

func TestErrorValidateValidRoomEntityNotEnoughPermissions(t *testing.T) {
	room := getTestEntity("room")
	room["domain"] = integration.TestDBName
	requestBody, _ := json.Marshal(room)

	endpoint := test_utils.GetEndpoint("validateEntity", "rooms")
	expectedMessage := "This user does not have sufficient permissions to create this object under this domain "
	test_utils.ValidateRequestWithUser(t, "POST", endpoint, requestBody, "viewer", http.StatusUnauthorized, expectedMessage)
}

func TestGetStats(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/stats", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	numberOfRacks, exists := response["Number of racks:"].(float64)
	assert.True(t, exists)
	assert.True(t, numberOfRacks > 0)
}

func TestGetApiVersion(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/version", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)
	status, exists := response["status"].(bool)
	assert.True(t, exists)
	assert.True(t, status)

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	customer, exists := data["Customer"].(string)
	assert.True(t, exists)
	assert.True(t, len(customer) > 0)
}

// Tests layers objects
func TestGetLayersObjectsRootRequired(t *testing.T) {
	endpoint := test_utils.GetEndpoint("layersObjects", "racks-layer")
	expectedMessage := "Query param root is mandatory"
	test_utils.ValidateManagedRequest(t, "GET", endpoint, nil, http.StatusBadRequest, expectedMessage)
}

func TestGetLayersObjectsLayerUnknown(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/layers/unknown/objects?root=site-no-temperature.building-1.room-1", nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func TestGetLayersObjectsWithSimpleFilter(t *testing.T) {
	endpoint := test_utils.GetEndpoint("layersObjects", "racks-layer")
	expectedMessage := "successfully processed request"
	response := test_utils.ValidateManagedRequest(t, "GET", endpoint+"?root=site-no-temperature.building-1.room-1", nil, http.StatusOK, expectedMessage)

	data, exists := response["data"].([]any)
	assert.True(t, exists)
	assert.Equal(t, 2, len(data))

	condition := true
	for _, rack := range data {
		condition = condition && rack.(map[string]any)["parentId"] == "site-no-temperature.building-1.room-1"
		condition = condition && rack.(map[string]any)["category"] == "rack"
	}

	assert.True(t, condition)
}

func TestGetLayersObjectsWithDoubleFilter(t *testing.T) {
	endpoint := test_utils.GetEndpoint("layersObjects", "racks-1-layer")
	expectedMessage := "successfully processed request"
	response := test_utils.ValidateManagedRequest(t, "GET", endpoint+"?root=site-no-temperature.building-1.room-*", nil, http.StatusOK, expectedMessage)

	data, exists := response["data"].([]any)
	assert.True(t, exists)
	assert.Equal(t, 2, len(data))

	condition := true
	for _, rack := range data {
		condition = condition && strings.HasPrefix(rack.(map[string]any)["parentId"].(string), "site-no-temperature.building-1.room-")
		condition = condition && rack.(map[string]any)["category"] == "rack"
		condition = condition && rack.(map[string]any)["name"] == "rack-1"
	}

	assert.True(t, condition)
}
