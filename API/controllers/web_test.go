package controllers_test

import (
	"encoding/json"
	"net/http"
	"p3/test/e2e"
	"p3/test/integration"
	test_utils "p3/test/utils"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	integration.RequireCreateSite("site-project")
}

var projectsEndpoint = test_utils.GetEndpoint("projects")
var alertsEndpoint = test_utils.GetEndpoint("alerts")

func TestCreateProjectInvalidBody(t *testing.T) {
	e2e.TestInvalidBody(t, "POST", projectsEndpoint, "Invalid request")
}

func TestCreateProject(t *testing.T) {
	requestBody, _ := json.Marshal(map[string]any{
		"attributes":       []string{"domain"},
		"authorLastUpdate": "admin@admin.com",
		"dateRange":        "01/01/2023-02/02/2023",
		"lastUpdate":       "02/02/2023",
		"name":             "project1",
		"namespace":        "physical",
		"objects":          []string{"site-project"},
		"showAvg":          false,
		"showSum":          false,
		"permissions":      []string{"admin@admin.com"},
	})

	e2e.ValidateManagedRequest(t, "POST", projectsEndpoint, requestBody, http.StatusOK, "successfully handled project request")
}

func TestGetProjectsWithNoUserRespondsWithError(t *testing.T) {
	e2e.ValidateManagedRequest(t, "GET", projectsEndpoint, nil, http.StatusBadRequest, "Error: user should be sent as query param")
}

func TestGetProjectsFromUserWithNoProjects(t *testing.T) {
	integration.CreateTestProject(t, "temporaryProject")
	response := e2e.ValidateManagedRequest(t, "GET", projectsEndpoint+"?user=someUser", nil, http.StatusOK, "successfully got projects")

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	projects, exists := data["projects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 0, len(projects))
}

func TestGetProjects(t *testing.T) {
	_, id := integration.CreateTestProject(t, "temporaryProject")
	response := e2e.ValidateManagedRequest(t, "GET", projectsEndpoint+"?user=admin@admin.com", nil, http.StatusOK, "successfully got projects")

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	projects, exists := data["projects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 2, len(projects)) // temporaryProject and project1

	exists = slices.ContainsFunc(projects, func(project interface{}) bool {
		return project.((map[string]interface{}))["Id"] == id
	})
	assert.True(t, exists)
}

func TestUpdateProject(t *testing.T) {
	temporaryProject, id := integration.CreateTestProject(t, "temporaryProject")
	temporaryProject.ShowAvg = true
	requestBody, _ := json.Marshal(temporaryProject)

	response := e2e.ValidateManagedRequest(t, "PUT", projectsEndpoint+"/"+id, requestBody, http.StatusOK, "successfully handled project request")

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	showAvg, exists := data["showAvg"].(bool)
	assert.True(t, exists)
	assert.True(t, showAvg)
}

func TestDeleteProject(t *testing.T) {
	_, id := integration.CreateTestProject(t, "temporaryProject")
	e2e.ValidateManagedRequest(t, "DELETE", projectsEndpoint+"/"+id, nil, http.StatusOK, "successfully removed project")

	// if we try to delete again we get an error
	e2e.ValidateManagedRequest(t, "DELETE", projectsEndpoint+"/"+id, nil, http.StatusNotFound, "Project not found")
}

func TestCreateAlertInvalidBody(t *testing.T) {
	e2e.TestInvalidBody(t, "POST", alertsEndpoint, "Invalid request")
}

func TestCreateAlert(t *testing.T) {
	requestBody, _ := json.Marshal(map[string]any{
		"id":       "OBJID.WITH.ALERT",
		"type":     "minor",
		"title":    "This is the title",
		"subtitle": "More information",
	})

	e2e.ValidateManagedRequest(t, "POST", alertsEndpoint, requestBody, http.StatusOK, "successfully created alert")
}

func TestGetAlerts(t *testing.T) {
	_, id := integration.CreateTestAlert(t, "temporaryAlert", false)
	response := e2e.ValidateManagedRequest(t, "GET", alertsEndpoint, nil, http.StatusOK, "successfully got alerts")

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	alerts, exists := data["alerts"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 2, len(alerts)) // temporaryAlert and OBJID.WITH.ALERT

	exists = slices.ContainsFunc(alerts, func(project interface{}) bool {
		return project.((map[string]interface{}))["id"] == id
	})
	assert.True(t, exists)
}

func TestGetAlert(t *testing.T) {
	_, id := integration.CreateTestAlert(t, "tempAlert", false)
	response := e2e.ValidateManagedRequest(t, "GET", alertsEndpoint+"/"+id, nil, http.StatusOK, "successfully got alert")

	alert, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	assert.Equal(t, id, alert["id"])
}

func TestDeleteAlert(t *testing.T) {
	_, id := integration.CreateTestAlert(t, "tempAlert2", true)
	println("one")
	e2e.ValidateManagedRequest(t, "DELETE", alertsEndpoint+"/"+id, nil, http.StatusOK, "successfully removed alert")
	println("two")
	// if we try to delete again we get an error
	// e2e.ValidateManagedRequest(t, "DELETE", alertsEndpoint+"/"+id, nil, http.StatusNotFound, "Alert not found")
}
