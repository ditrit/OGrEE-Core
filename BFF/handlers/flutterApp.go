package handlers

import (
	"ogree-bff/controllers"

	"github.com/gin-gonic/gin"
)

// swagger:operation GET /projects FlutterApp GetProjects
// Get a list of projects for the specified user.
// ---
// security:
//   - Bearer: []
// produces:
// 	 - application/json
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

func GetProjects(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation POST /projects FlutterApp CreateProjects
// Create a new project
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
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
//
// responses:
//
//	'200':
//		description: 'Project successfully created.'
//	'400':
//		description: 'Bad Request. Invalid project format.'
//	'500':
//		description: 'Internal server error.'
func CreateProjects(c *gin.Context) {
	controllers.Post(c, "objects")
}

// swagger:operation PUT /projects/{ProjectID} FlutterApp UpdateProjects
// Replace the data of an existing project.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
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
//
// responses:
//
//	'200':
//		description: Project successfully updated.
//	'400':
//		description: Bad Request. Invalid project format.
//	'500':
//		description: Internal server error
func UpdateProjects(c *gin.Context) {
	controllers.Put(c, "objects")
}

// swagger:operation DELETE /projects/{ProjectID} FlutterApp DeleteProjects
// Delete an existing project.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: ProjectID
//     in: path
//     description: 'ID of the project to delete.'
//     required: true
//     type: string
//     default: "1234"
//
// responses:
//
//	'200':
//	    description: Project successfully updated.
//	'404':
//	    description: Not Found. Invalid project ID.
//	'500':
//	    description: Internal server error
func DeleteProjects(c *gin.Context) {
	controllers.Delete(c, "objects")
}
