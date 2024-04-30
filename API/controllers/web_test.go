package controllers_test

import (
	"encoding/json"
	"net/http"
	"p3/test/e2e"
	"p3/test/integration"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	integration.RequireCreateSite("site-project")
}

var project = map[string]any{
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
}
var projectId string

func TestCreateProjectInvalidBody(t *testing.T) {
	e2e.TestInvalidBody(t, "POST", "/api/projects", "Invalid request")
}

func TestCreateProject(t *testing.T) {
	json.Marshal(project)
	requestBody, _ := json.Marshal(project)

	recorder := e2e.MakeRequest("POST", "/api/projects", requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully handled project request", message)
}

// Tests domain bulk creation (/api/users/bulk)
func TestGetProjectsWithNoUserRespondsWithError(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/projects", nil)
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Error: user should be sent as query param", message)
}

func TestGetProjectsFromUserWithNoProjects(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/projects?user=someUser", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully got projects", message)

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	projects, exists := data["projects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 0, len(projects))
}

func TestGetProjects(t *testing.T) {
	recorder := e2e.MakeRequest("GET", "/api/projects?user=admin@admin.com", nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully got projects", message)

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	projects, exists := data["projects"].([]interface{})
	assert.True(t, exists)
	assert.Equal(t, 1, len(projects))

	projectName, exists := projects[0].(map[string]interface{})["name"].(string)
	assert.True(t, exists)
	assert.Equal(t, "project1", projectName)

	projectId = projects[0].(map[string]interface{})["Id"].(string)
}

func TestUpdateProject(t *testing.T) {
	project["showAvg"] = true
	json.Marshal(project)
	requestBody, _ := json.Marshal(project)

	recorder := e2e.MakeRequest("PUT", "/api/projects/"+projectId, requestBody)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully handled project request", message)

	data, exists := response["data"].(map[string]interface{})
	assert.True(t, exists)
	showAvg, exists := data["showAvg"].(bool)
	assert.True(t, exists)
	assert.True(t, showAvg)
}

func TestDeleteProject(t *testing.T) {
	recorder := e2e.MakeRequest("DELETE", "/api/projects/"+projectId, nil)
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "successfully removed project", message)
}

func TestDeleteProjectNonExistent(t *testing.T) {
	recorder := e2e.MakeRequest("DELETE", "/api/projects/"+projectId, nil)
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var response map[string]interface{}
	json.Unmarshal(recorder.Body.Bytes(), &response)

	message, exists := response["message"].(string)
	assert.True(t, exists)
	assert.Equal(t, "Project not found", message)
}
