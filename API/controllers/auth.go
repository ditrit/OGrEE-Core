package controllers

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"p3/models"
	u "p3/utils"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// swagger:operation POST /api/users Organization CreateAccount
// Create a new user user.
// Create an account with email and password credentials, it returns
// a JWT key to use with the API.
// ---
// security:
// - bearer: []
// produces:
// - application/json
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
//     '201':
//         description: New account created
//     '400':
//         description: Bad request
//     '403':
//         description: User not authorised to create an account
//     '500':
//         description: Internal server error

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "POST")
	} else {
		account := &models.Account{}
		err := json.NewDecoder(r.Body).Decode(account)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Invalid request: wrong format body"))
			return
		}

		callerUser := getUserFromToken(w, r)
		if callerUser == nil {
			return
		}

		acc, e := account.Create(callerUser.Roles)
		if e != nil {
			u.RespondWithError(w, e)
		} else {
			w.WriteHeader(http.StatusCreated)
			resp := u.Message("Account has been created")
			resp["account"] = acc
			u.Respond(w, resp)
		}
	}
}

// swagger:operation POST /api/users/bulk Organization CreateBulk
// Create multiples users with one request.
// ---
// security:
// - bearer: []
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
//		'200':
//			description: Request processed, check response body for results
//		'400':
//			description: Bad request
//		'500':
//			description: Internal server error

func CreateBulkAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateBulkAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	var accounts []models.Account
	err := json.NewDecoder(r.Body).Decode(&accounts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Invalid request"))
		return
	}

	callerUser := getUserFromToken(w, r)
	if callerUser == nil {
		return
	}

	resp := map[string]interface{}{}
	for _, account := range accounts {
		password := ""
		if len(account.Password) <= 0 {
			password = randStringBytes(8)
			account.Password = password
		}
		resp[account.Email] = map[string]interface{}{}
		_, e := account.Create(callerUser.Roles)
		if e != nil {
			resp[account.Email].(map[string]interface{})["status"] = e.Message
		} else {
			resp[account.Email].(map[string]interface{})["status"] = "successfully created"
			if password != "" {
				resp[account.Email].(map[string]interface{})["password"] = password
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	u.Respond(w, resp)
}

const passChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = passChars[rand.Intn(len(passChars))]
	}
	return string(b)
}

// swagger:operation POST /api/login Authentication Authenticate
// Generates a new JWT Key for the client.
// Create a new JWT Key. This can also be used to verify credentials
// The authorize and 'Try it out' buttons don't work
// ---
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     description: 'Mandatory: email and password.'
//     required: true
//     format: object
//     example: '{"email": "user@test.com", "password": "secret123"}'
// responses:
//     '200':
//         description: Authenticated
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error

func Authenticate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 Authenticate ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "POST")
	} else {
		var account models.Account
		err := json.NewDecoder(r.Body).Decode(&account)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Invalid request"))
			return
		}

		acc, e := models.Login(account.Email, account.Password)
		if e != nil {
			u.RespondWithError(w, e)
		} else {
			resp := u.Message("Login succesful")
			resp["account"] = acc
			u.Respond(w, resp)
		}
	}
}

// swagger:operation GET /api/token/valid Authentication VerifyToken
// Verify if token sent in the header is valid.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// responses:
//     '200':
//         description: Token is valid.
//     '403':
//         description: Unauthorized
//     '500':
//         description: Internal server error

func VerifyToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 Verify ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "GET")
	} else {
		u.Respond(w, u.Message("working"))
	}
}

// swagger:operation GET /api/users Organization GetAllAccounts
// Get a list of users that the caller is allowed to see.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// responses:
//     '200':
//          description: Return all possible users
//     '500':
//          description: Internal server error

func GetAllAccounts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetAllAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "GET, HEAD")
	} else {
		var resp map[string]interface{}

		// Get caller user
		callerUser := getUserFromToken(w, r)
		if callerUser == nil {
			return
		}

		// Get users
		users, err := models.GetAllUsers(callerUser.Roles)
		if err != nil {
			u.RespondWithError(w, err)
		} else {
			resp = u.Message("successfully got users")
			resp["data"] = users
			u.Respond(w, resp)
		}
	}
}

// swagger:operation DELETE /api/users/{UserId} Organization RemoveAccount
// Remove the specified user account.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
// - name: UserId
//   in: path
//   description: 'The ID of the user to delete'
//   required: true
//   type: string
//   example: "someUserId"
// responses:
//		'200':
//			description: User removed
//		'400':
//			description: User ID not valid or not found
//		'403':
//			description: Caller not authorised to delete this user
//		'500':
//			description: Internal server error

func RemoveAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 RemoveAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "DELETE, HEAD")
	} else {
		// Get caller user
		callerUser := getUserFromToken(w, r)
		if callerUser == nil {
			return
		}

		// Get user to delete
		userId := mux.Vars(r)["id"]
		objID, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("User ID is not valid"))
			return
		}
		deleteUser := models.GetUser(objID)
		if deleteUser == nil {
			w.WriteHeader(http.StatusNotFound)
			u.Respond(w, u.Message("User not found"))
			return
		}

		// Check permissions
		if !models.CheckCanManageUser(callerUser.Roles, deleteUser.Roles) {
			w.WriteHeader(http.StatusUnauthorized)
			u.Respond(w, u.Message("Caller does not have permission to delete this user"))
			return
		}

		// Delete it
		e := models.DeleteUser(objID)
		if e != nil {
			u.RespondWithError(w, e)
		} else {
			u.Respond(w, u.Message("successfully removed user"))
		}
	}
}

// swagger:operation PATCH /api/users/{UserId} Organization ModifyUserRoles
// Modify user permissions: domain and role.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: UserId
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
//		'200':
//			description: User roles modified
//		'400':
//			description: Bad request
//		'403':
//			description: Caller not authorised to modify this user
//		'500':
//			description: Internal server error

func ModifyUserRoles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 ModifyUserRoles ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "PATCH, HEAD")
	} else {
		var resp map[string]interface{}
		userId := mux.Vars(r)["id"]

		// Check if POST body is valid
		rolesConverted, err := getUserRolesFromBody(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(err.Error()))
			return
		}

		// Get caller user
		callerUser := getUserFromToken(w, r)
		if callerUser == nil {
			return
		}

		// Get user to modify
		objID, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			resp = u.Message("User ID is not valid")
			u.Respond(w, resp)
			return
		}
		modifyUser := models.GetUser(objID)

		// Check permissions
		if !models.CheckCanManageUser(callerUser.Roles, modifyUser.Roles) {
			w.WriteHeader(http.StatusUnauthorized)
			resp = u.Message("Caller does not have permission to modify this user")
			u.Respond(w, resp)
			return
		}

		// Modify it
		e := models.ModifyUser(userId, rolesConverted)
		if e != nil {
			u.RespondWithError(w, e)
		} else {
			u.Respond(w, u.Message("successfully updated user roles"))
		}
	}
}

func getUserRolesFromBody(r *http.Request) (map[string]models.Role, error) {
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return nil, fmt.Errorf("invalid request")
	}

	roles, ok := data["roles"].(map[string]interface{})
	if len(data) > 1 || !ok {
		return nil, fmt.Errorf("only 'roles' should be provided to patch")
	}
	rolesConverted := map[string]models.Role{}
	for k := range roles {
		if v, ok := roles[k].(string); ok {
			rolesConverted[k] = models.Role(v)
		} else {
			return nil, fmt.Errorf("invalid roles format")
		}
	}
	return rolesConverted, nil
}

// swagger:operation POST /api/users/password/change Authentication ModifyUserPassword
// For logged in user to change its own password.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     description: 'Mandatory: currentPassword and newPassword.'
//     type: json
//     required: true
//     example: '{"currentPassword": "myOldPassword", "newPassword": "myNewPassword"}'
//
// responses:
//		'200':
//			description: Password changed
//		'400':
//			description: Bad request
//		'500':
//			description: Internal server error

// swagger:operation POST /api/users/password/reset Authentication ResetUserPassword
// Reset password after forgot.
// For user that first called forgot enpoint to change its password.
// A reset token generated by the forgot endpoint should be provided as the Authentication header.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     description: 'Mandatory: currentPassword and newPassword.'
//     type: json
//     required: true
//     example: '"newPassword": "myNewPassword"}'
//
// responses:
//		'200':
//			description: Password changed
//		'400':
//			description: Bad request
//		'500':
//			description: Internal server error

func ModifyUserPassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 ModifyUserPassword ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "POST, HEAD")
	} else {
		// Get user ID and email from token
		userData := r.Context().Value("user")
		if userData == nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Error while parsing path params"))
			u.ErrLog("Error while parsing path params", "GET GENERIC", "", r)
			return
		}
		userId := userData.(map[string]interface{})["userID"].(primitive.ObjectID)
		userEmail := userData.(map[string]interface{})["email"].(string)

		// Check if POST body is valid
		currentPassword, newPassword, isReset,
			err := getModifyPassDataFromBody(r, userEmail)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(err.Error()))
			return
		}

		// Check if user is valid
		var user *models.Account
		if isReset {
			user = models.GetUser(userId)
		} else {
			user = models.GetUserByEmail(userEmail)
		}
		if user == nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Invalid token: no valid user found"))
			u.ErrLog("Unable to find user associated to token", "GET GENERIC", "", r)
			return
		}

		// Change user password
		newToken, e := user.ChangePassword(currentPassword, newPassword, isReset)
		if e != nil {
			u.RespondWithError(w, e)
		} else {
			resp := u.Message("successfully updated user password")
			if !isReset {
				resp["token"] = newToken
			}
			u.Respond(w, resp)
		}
	}
}

func getModifyPassDataFromBody(r *http.Request, userEmail string) (string, string, bool, error) {
	isReset := false
	hasCurrent := true
	currentPassword := ""
	var data map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		return currentPassword, "", isReset, fmt.Errorf("invalid request")
	}
	if userEmail == u.RESET_TAG {
		// it's not change, it's reset (no need for current password)
		isReset = true
	} else {
		currentPassword, hasCurrent = data["currentPassword"].(string)
	}
	newPassword, hasNew := data["newPassword"].(string)
	if !hasCurrent || !hasNew {
		return currentPassword, "", isReset,
			fmt.Errorf("invalid request: wrong body format")
	}
	return currentPassword, newPassword, isReset, nil
}

// swagger:operation POST /api/users/password/forgot Authentication UserForgotPassword
// Forgot my password.
// Public endpoint to request a reset of a user's password (forgot my password).
// If the email is valid, an email with a reset token/link will be sent to the user.
// ---
// produces:
// - application/json
// parameters:
//   - name: body
//     in: body
//     description: 'Mandatory: email.'
//     type: string
//     required: true
//     example: '{"email": "user@test.com"}'
//
// responses:
//		'200':
//			description: request processed. If account exists, an email with a reset token is sent
//		'400':
//			description: Bad request
//		'500':
//			description: Internal server error

func UserForgotPassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 UserForgotPassword ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "POST, HEAD")
	} else {
		// Check if POST body is valid
		var data map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Invalid request"))
			return
		}
		userEmail, hasEmail := data["email"].(string)
		if !hasEmail {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message("Invalid request: email should be provided"))
			return
		}

		// Create token, if user exists, and send it by email
		user := models.GetUserByEmail(userEmail)
		if user != nil {
			token := models.GenerateToken(u.RESET_TAG, user.ID, time.Minute*15)
			if e := u.SendEmail(token, user.Email); e != "" {
				w.WriteHeader(http.StatusInternalServerError)
				u.Respond(w, u.Message("Unable to send email: "+e))
				return
			}
		}

		u.Respond(w, u.Message("request processed"))
	}
}
