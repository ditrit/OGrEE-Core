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
