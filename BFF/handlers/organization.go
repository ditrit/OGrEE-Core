package handlers

import (
	"ogree-bff/controllers"

	"github.com/gin-gonic/gin"
)

// swagger:operation POST /users Organization CreateAccount
// Create a new user user.
// Create an account with email and password credentials, it returns
// a JWT key to use with the API.
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
//     description: 'Mandatory: email, password and roles. Optional: name.
//     Roles is an object with domains as keys and roles as values.
//     Possible roles: manager, user and viewer'
//     required: true
//     format: object
//     example: '{"name": "John Doe", "roles": {"*": "manager"}, "email": "user@test.com", "password": "secret123"}'
//
// responses:
//
//	'201':
//	    description: New account created
//	'400':
//	    description: Bad request
//	'403':
//	    description: User not authorised to create an account
//	'500':
//	    description: Internal server error
func CreateAccount(c *gin.Context) {
	controllers.Post(c, "objects")
}

// swagger:operation POST /users/bulk Organization CreateBulk
// Create multiples users with one request.
// ---
// security:
//   - Bearer: []
//
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     description: 'An array of users. Same mandatory and optional parameters as user apply,
//     except for password. If not provided, one will be automatically created by the API.'
//     required: true
//     format: object
//     example: '[{"name": "John Doe", "roles": {"*": "manager"}, "email": "user@test.com"}]'
//
// responses:
//
//	'200':
//		description: Request processed, check response body for results
//	'400':
//		description: Bad request
//	'500':
//		description: Internal server error
func CreateBulk(c *gin.Context) {
	controllers.Post(c, "objects")
}

// swagger:operation GET /users Organization GetAllAccounts
// Get a list of users that the caller is allowed to see.
// ---
// security:
//   - Bearer: []
//
// produces:
// - application/json
// responses:
//
//	'200':
//	     description: Return all possible users
//	'500':
//	     description: Internal server error
func GetAllAccounts(c *gin.Context) {
	controllers.Get(c, "objects")
}

// swagger:operation DELETE /users/{userid} Organization RemoveAccount
// Remove the specified user account.
// ---
// security:
//   - Bearer: []
//
// produces:
// - application/json
// parameters:
//   - name: userid
//     in: path
//     description: 'The ID of the user to delete'
//     required: true
//     type: string
//     example: "someUserId"
//
// responses:
//
//	'200':
//		description: User removed
//	'400':
//		description: User ID not valid or not found
//	'403':
//		description: Caller not authorised to delete this user
//	'500':
//		description: Internal server error
func RemoveAccount(c *gin.Context) {
	controllers.Delete(c, "objects")
}

// swagger:operation PATCH /users/{userid} Organization ModifyUserRoles
// Modify user permissions: domain and role.
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// parameters:
//   - name: userid
//     in: path
//     description: 'The ID of the user to modify roles'
//     required: true
//     type: string
//     example: "someUserId"
//   - name: roles
//     in: body
//     description: An object with domains as keys and roles as values
//     type: json
//     required: true
//     example: '{"roles": {"*": "manager"}}'
//
// responses:
//
//	'200':
//		description: User roles modified
//	'400':
//		description: Bad request
//	'403':
//		description: Caller not authorised to modify this user
//	'500':
//		description: Internal server error
func ModifyUserRoles(c *gin.Context) {
	controllers.Patch(c, "objects")
}

// swagger:operation POST /domains/bulk Organization CreateBulkDomain
// Create multiple domains in a single request.
// An array of domains should be provided in the body.
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
//     required: true
//     default: [{}]
//
// responses:
//
//	'200':
//	    description: 'Request processed. Check the response body
//	    for individual results for each of the sent domains'
//	'400':
//	    description: 'Bad format: body is not a valid list of domains.'
func CreateBulkDomain(c *gin.Context) {
	controllers.Post(c, "objects")
}

// swagger:operation GET /hierarchy/domains Organization GetCompleteDomainHierarchy
// Returns domain complete hierarchy.
// Return is arranged by relationship (father:[children]),
// starting with "Root":[root domains].
// ---
// security:
//   - Bearer: []
//
// produces:
//   - application/json
//
// responses:
//
//	'200':
//	     description: 'Request is valid.'
//	'500':
//	     description: Server error.
func GetCompleteDomainHierarchy(c *gin.Context) {
	controllers.Get(c, "objects")
}
