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

// swagger:operation POST /api/users auth Create
// Generate credentials for a user.
// Create an account with email credentials, it returns
// a JWT key to use with the API.
// ---
// produces:
// - application/json
// parameters:
//   - name: name
//     in: json
//     description: User name
//     type: string
//     required: false
//     default: "John Doe"
//   - name: email
//     in: json
//     description: User Email Address
//     type: string
//     required: true
//     default: "user@email.com"
//   - name: password
//     in: json
//     description: User password
//     required: true
//     format: password
//     default: "secret123"
//
// responses:
//     '201':
//         description: Authenticated and new account created
//     '400':
//         description: Bad request
//     '403':
//         description: User not authorised to create an account
//     '500':
//         description: Internal server error

// swagger:operation OPTIONS /api/users auth CreateOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Returns header with possible operations

func CreateAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "POST, OPTIONS")
	} else {
		account := &models.Account{}
		err := json.NewDecoder(r.Body).Decode(account)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Invalid request"))
			return
		}

		callerUser := getUserFromToken(w, r)
		if callerUser == nil {
			return
		}

		resp, e := account.Create(callerUser.Roles)
		switch e {
		case "internal":
			w.WriteHeader(http.StatusInternalServerError)
		case "clientError", "validate":
			w.WriteHeader(http.StatusBadRequest)
		case "unauthorised":
			w.WriteHeader(http.StatusForbidden)
		case "exists":
			w.WriteHeader(http.StatusConflict)
		default:
			w.WriteHeader(http.StatusCreated)
		}
		u.Respond(w, resp)
	}
}

// swagger:operation POST /api/users/bulk auth CreateBulk
// Create multiples users with one request.
// ---
// produces:
// - application/json
// parameters:
//   - name: name
//     in: json
//     description: User name
//     type: string
//     required: false
//     default: "John Doe"
//   - name: email
//     in: body
//     description: User Email Address
//     type: string
//     required: true
//     default: "user@email.com"
//   - name: password
//     in: json
//     description: User password
//     required: true
//     format: password
//     default: "secret123"
//
// responses:
//
//	'200':
//	    description: Request processed, check response body for results
//	'400':
//	    description: Bad request
//	'500':
//	    description: Internal server error
func CreateBulkAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 CreateBulkAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	var accounts []models.Account
	err := json.NewDecoder(r.Body).Decode(&accounts)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Invalid request"))
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
		res, _ := account.Create(callerUser.Roles)
		for key, value := range res {
			if key == "account" && password != "" {
				resp[account.Email].(map[string]interface{})["password"] = password
			} else {
				resp[account.Email].(map[string]interface{})[key] = value
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

// swagger:operation POST /api/login auth Authenticate
// Generates a new JWT Key for the client.
// Create a new JWT Key. This can also be used to verify credentials
// The authorize and 'Try it out' buttons don't work
// ---
// produces:
// - application/json
// parameters:
// - name: username
//   in: body
//   description: Your Email Address
//   type: string
//   required: true
//   default: "infiniti@nissan.com"
// - name: password
//   in: json
//   description: Your password
//   required: true
//   format: password
//   default: "secret"
// responses:
//     '200':
//         description: Authenticated
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error

// swagger:operation OPTIONS /api/login auth CreateOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Returns header with possible operations
func Authenticate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 Authenticate ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "POST, OPTIONS")
	} else {
		var account models.Account
		err := json.NewDecoder(r.Body).Decode(&account)
		if err != nil {
			u.Respond(w, u.Message(false, "Invalid request"))
			return
		}

		resp, e := models.Login(account.Email, account.Password)
		if resp["status"] == false {
			if e == "validate" {
				w.WriteHeader(http.StatusUnauthorized)
			} else if e == "internal" {
				w.WriteHeader(http.StatusInternalServerError)
			} else if e == "clientError" {
				w.WriteHeader(http.StatusBadRequest)
			}
		}
		u.Respond(w, resp)
	}
}

// swagger:operation GET /api/token/valid auth VerifyToken
// A custom client specified URL for verifying if their key is valid.
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Authenticated
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error

// swagger:operation OPTIONS /api/token/valid auth VerifyToken
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Returns header with possible operations
func VerifyToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 Verify ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS")
	} else {
		u.Respond(w, u.Message(true, "working"))
	}
}

// swagger:operation GET /api/users auth GetAllAccounts
// Get a list of users that the caller is allowed to see.
// ---
// produces:
// - application/json
// responses:
//
//	'200':
//	    description: Got all possible users
//	'500':
//	    description: Internal server error
func GetAllAccounts(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetAllAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS, HEAD")
	} else {
		var resp map[string]interface{}

		// Get caller user
		callerUser := getUserFromToken(w, r)
		if callerUser == nil {
			return
		}

		// Get users
		data, err := models.GetAllUsers(callerUser.Roles)
		if err != "" {
			w.WriteHeader(http.StatusInternalServerError)
			resp = u.Message(false, "Error: "+err)
		} else {
			resp = u.Message(true, "successfully got users")
			resp["data"] = data
		}
		u.Respond(w, resp)
	}
}

// swagger:operation DELETE /api/users/{id} auth RemoveAccount
// Remove the specified user account.
// ---
// produces:
// - application/json
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
func RemoveAccount(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 RemoveAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "DELETE, OPTIONS, HEAD")
	} else {
		var resp map[string]interface{}

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
			resp = u.Message(false, "User ID is not valid")
			u.Respond(w, resp)
			return
		}
		deleteUser := models.GetUser(objID)

		// Check permissions
		if !models.CheckCanManageUser(callerUser.Roles, deleteUser.Roles) {
			w.WriteHeader(http.StatusUnauthorized)
			resp = u.Message(false, "Caller does not have permission to delete this user")
			u.Respond(w, resp)
			return
		}

		// Delete it
		e := models.DeleteUser(objID)
		if e != "" {
			w.WriteHeader(http.StatusInternalServerError)
			resp = u.Message(false, "Error: "+e)
		} else {
			resp = u.Message(true, "successfully removed user")
		}
		u.Respond(w, resp)
	}
}

// swagger:operation PATCH /api/users/{id} auth ModifyUserRoles
// Modify user permissions: domain and role.
// ---
// produces:
// - application/json
// parameters:
//   - name: roles
//     in: body
//     description: An object with domains as keys and roles as values
//     type: json
//     required: true
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
func ModifyUserRoles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 ModifyUserRoles ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "PATCH, OPTIONS, HEAD")
	} else {
		var resp map[string]interface{}
		userId := mux.Vars(r)["id"]

		// Check if POST body is valid
		var data map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Invalid request"))
			return
		}
		roles, ok := data["roles"].(map[string]interface{})
		if len(data) > 1 || !ok {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Only 'roles' should be provided to patch"))
			return
		}
		rolesConverted := map[string]models.Role{}
		for k := range roles {
			if v, ok := roles[k].(string); ok {
				rolesConverted[k] = models.Role(v)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				u.Respond(w, u.Message(false, "Invalid roles format"))
				return
			}
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
			resp = u.Message(false, "User ID is not valid")
			u.Respond(w, resp)
			return
		}
		modifyUser := models.GetUser(objID)

		// Check permissions
		if !models.CheckCanManageUser(callerUser.Roles, modifyUser.Roles) {
			w.WriteHeader(http.StatusUnauthorized)
			resp = u.Message(false, "Caller does not have permission to modify this user")
			u.Respond(w, resp)
			return
		}

		// Modify it
		e, eType := models.ModifyUser(userId, rolesConverted)
		if e != "" {
			switch eType {
			case "internal":
				w.WriteHeader(http.StatusInternalServerError)
			case "validate":
				w.WriteHeader(http.StatusBadRequest)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			resp = u.Message(false, "Error: "+e)
		} else {
			resp = u.Message(true, "successfully updated user roles")
		}
		u.Respond(w, resp)
	}
}

// swagger:operation POST /api/users/password/change auth ModifyUserPassword
// For logged in user to change own password.
// ---
// produces:
// - application/json
// parameters:
//   - name: currentPassword
//     in: body
//     description: User current password
//     type: string
//     required: true
//   - name: newPassword
//     in: body
//     description: User new desired password
//     type: string
//     required: true
//
// responses:
//
//	'200':
//		description: Password changed
//	'400':
//		description: Bad request
//	'500':
//		description: Internal server error

// swagger:operation POST /api/users/password/reset auth ModifyUserPassword
// To change password of user that forgot password and received a reset token by email.
// ---
// produces:
// - application/json
// parameters:
//   - name: newPassword
//     in: body
//     description: User new desired password
//     type: string
//     required: true
//
// responses:
//
//	'200':
//		description: Password changed
//	'400':
//		description: Bad request
//	'500':
//		description: Internal server error
func ModifyUserPassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 ModifyUserPassword ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "POST, OPTIONS, HEAD")
	} else {
		var resp map[string]interface{}

		// Get user ID and email from token
		userData := r.Context().Value("user")
		if userData == nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Error while parsing path params"))
			u.ErrLog("Error while parsing path params", "GET GENERIC", "", r)
			return
		}
		userId := userData.(map[string]interface{})["userID"].(primitive.ObjectID)
		userEmail := userData.(map[string]interface{})["email"].(string)

		// Check if POST body is valid
		var data map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Invalid request"))
			return
		}
		isReset := false
		hasCurrent := true
		currentPassword := ""
		if userEmail == u.RESET_TAG {
			isReset = true
		} else {
			currentPassword, hasCurrent = data["currentPassword"].(string)
		}
		newPassword, hasNew := data["newPassword"].(string)
		if !hasCurrent || !hasNew {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Invalid request: wrong body format"))
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
			u.Respond(w, u.Message(false, "Invalid token: no valid user found"))
			u.ErrLog("Unable to find user associated to token", "GET GENERIC", "", r)
			return
		}

		// Change user password
		response, errType := user.ChangePassword(currentPassword, newPassword, isReset)
		if errType != "" {
			switch errType {
			case "internal":
				w.WriteHeader(http.StatusInternalServerError)
			case "validate":
				w.WriteHeader(http.StatusBadRequest)
			default:
				w.WriteHeader(http.StatusInternalServerError)
			}
			resp = u.Message(false, "Error: "+response)
		} else {
			resp = u.Message(true, "successfully updated user password")
			if !isReset {
				resp["token"] = response
			}
		}
		u.Respond(w, resp)

	}
}

// swagger:operation POST /api/users/password/forgot auth UserForgotPassword
// To request a reset of a user's password (forgot my password).
// ---
// produces:
// - application/json
// parameters:
//   - name: email
//     in: body
//     description: User email
//     type: string
//     required: true
//
// responses:
//
//	'200':
//		description: request processed. If account exists, an email with a reset token will be sent to it
//	'400':
//		description: Bad request
//	'500':
//		description: Internal server error
func UserForgotPassword(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 UserForgotPassword ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "POST, OPTIONS, HEAD")
	} else {
		resp := map[string]interface{}{}

		// Check if POST body is valid
		var data map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Invalid request"))
			return
		}
		userEmail, hasEmail := data["email"].(string)
		if !hasEmail {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Invalid request: email should be provided"))
			return
		}

		// Create token, if user exists, and send it by email
		user := models.GetUserByEmail(userEmail)
		if user != nil {
			token := models.GenerateToken(u.RESET_TAG, user.ID, time.Minute*15)
			if e := u.SendEmail(token, user.Email); e != "" {
				w.WriteHeader(http.StatusInternalServerError)
				u.Respond(w, u.Message(false, "Unable to send email: "+e))
				return
			}
		}

		resp = u.Message(true, "request processed")
		u.Respond(w, resp)
	}
}
