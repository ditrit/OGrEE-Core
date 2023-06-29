package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"p3/models"
	u "p3/utils"

	"github.com/gorilla/mux"
)

// swagger:operation GET /api/projects FlutterApp GetProjects
// Get a list of projects for the specified user.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
// - name: user
//   in: query
//   description: 'Email of the user whose projects are being requested.
//   Example: /api/projects?user=user@test.com'
//   required: false
//   type: string
//   default: user@test.com
// responses:
//		'200':
//			description: 'Return all possible projects.'
//		'400':
//			description: 'Bad Request. Invalid user query param.'
//		'500':
//			description: 'Internal server error.'

func GetProjects(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetProjects ")
	fmt.Println("******************************************************")

	query, _ := url.ParseQuery(r.URL.RawQuery)

	if len(query["user"]) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		resp := u.Message("Error: user should be sent as query param")
		u.Respond(w, resp)
		return
	}

	projects, err := models.GetProjectsByUserEmail(query["user"][0])
	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			resp := map[string]interface{}{}
			resp["projects"] = projects
			u.Respond(w, u.RespDataWrapper("successfully got projects", resp))
		}
	}
}

// swagger:operation POST /api/projects FlutterApp CreateProjects
// Create a new project
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     description: 'Mandatory: name, dateRange, namespace, attributes,
//     objects, permissions, authorLastUpdate, lastUpdate.
//     Optional: showAvg, showSum, isPublic.'
//     required: true
//     format: object
//     example: '{"attributes":["domain"],"authorLastUpdate":"helder","dateRange":"01/01/2023-02/02/2023",
//     "lastUpdate":"02/02/2023","name":"test 1","namespace":"physical","objects":["siteB"],"showAvg":false,
//     "showSum":false,"permissions":["user@test.com","admin"]}'
// responses:
//		'200':
//			description: 'Project successfully created.'
//		'400':
//			description: 'Bad Request. Invalid project format.'
//		'500':
//			description: 'Internal server error.'

// swagger:operation PUT /api/projects/{ProjectID} FlutterApp UpdateProjects
// Replace the data of an existing project.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: ProjectID
//     in: path
//     description: 'ID of the project to update.'
//     required: true
//     type: string
//     default: "1234"
//   - name: body
//     in: body
//     description: 'Mandatory: name, dateRange, namespace, attributes,
//     objects, permissions, authorLastUpdate, lastUpdate.
//     Optional: showAvg, showSum, isPublic.'
//     required: true
//     format: object
//     example: '{"attributes":["domain"],"authorLastUpdate":"helder","dateRange":"01/01/2023-02/02/2023",
//     "lastUpdate":"02/02/2023","name":"test 1","namespace":"physical","objects":["siteB"],"showAvg":false,
//     "showSum":false,"permissions":["user@test.com","admin"]}'
// responses:
//		'200':
//			description: Project successfully updated.
//		'400':
//			description: Bad Request. Invalid project format.
//		'500':
//			description: Internal server error

func CreateOrUpdateProject(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateOrUpdateProject ")
	fmt.Println("******************************************************")

	project := &models.Project{}
	err := json.NewDecoder(r.Body).Decode(project)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Invalid request"))
		return
	}

	var modelErr *u.Error
	if r.Method == "POST" {
		// Create project
		modelErr = models.AddProject(*project)
	} else {
		// Update project
		modelErr = models.UpdateProject(*project, mux.Vars(r)["id"])
	}

	if modelErr != nil {
		u.RespondWithError(w, modelErr)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			resp := u.Message("successfully handled project request")
			resp["data"] = project
			u.Respond(w, resp)
		}
	}
}

// swagger:operation DELETE /api/projects/{ProjectID} FlutterApp DeleteProjects
// Delete an existing project.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
// - name: ProjectID
//   in: path
//   description: 'ID of the project to delete.'
//   required: true
//   type: string
//   default: "1234"
// responses:
//  '200':
//      description: Project successfully updated.
//  '404':
//      description: Not Found. Invalid project ID.
//  '500':
//      description: Internal server error

func DeleteProject(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 DeleteProject ")
	fmt.Println("******************************************************")

	err := models.DeleteProject(mux.Vars(r)["id"])

	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			w.Header().Add("Content-Type", "application/json")
			w.Header().Add("Allow", "GET, OPTIONS, HEAD")
		} else {
			u.Respond(w, u.Message("successfully removed project"))
		}
	}
}
