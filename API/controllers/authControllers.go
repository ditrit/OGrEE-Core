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

// swagger:operation POST /api auth Create
// Generate credentials for a user.
// Create an account with Email credentials, it returns
// a JWT key to use with the API. The
// authorize and 'Try it out' buttons don't work
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
//     '201':
//         description: Authenticated and new account created
//     '400':
//         description: Bad request
//     '403':
//         description: User not authorised to create an account
//     '500':
//         description: Internal server error

// swagger:operation OPTIONS /api auth CreateOptions
// Displays possible operations for the resource in response header.
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Returns header with possible operations

var CreateAccount = func(w http.ResponseWriter, r *http.Request) {
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

		userData := r.Context().Value("user")
		if userData == nil {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Error while parsing path params"))
			u.ErrLog("Error while parsing path params", "GET GENERIC", "", r)
			return
		}
		userId := userData.(map[string]interface{})["userID"].(primitive.ObjectID)
		user := models.GetUser(userId)
		fmt.Println(user)
		if user == nil || len(user.Roles) <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			u.Respond(w, u.Message(false, "Invalid token: no valid user found"))
			u.ErrLog("Unable to find user associated to token", "GET GENERIC", "", r)
			return
		}

		resp, e := account.Create(user.Roles)
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

var CreateBulkAccount = func(w http.ResponseWriter, r *http.Request) {
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

	userData := r.Context().Value("user")
	if userData == nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while parsing path params"))
		u.ErrLog("Error while parsing path params", "GET GENERIC", "", r)
		return
	}
	userId := userData.(map[string]interface{})["userID"].(primitive.ObjectID)
	user := models.GetUser(userId)
	fmt.Println(user)
	if user == nil || len(user.Roles) <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Invalid token: no valid user found"))
		u.ErrLog("Unable to find user associated to token", "GET GENERIC", "", r)
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
		res, _ := account.Create(user.Roles)
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
var Authenticate = func(w http.ResponseWriter, r *http.Request) {
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
var Verify = func(w http.ResponseWriter, r *http.Request) {
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

var GetAllAccounts = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetAllAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "GET, OPTIONS, HEAD")
	} else {
		var resp map[string]interface{}
		data, err := models.GetAllUsers()
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

var RemoveAccount = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 RemoveAccount ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "DELETE, OPTIONS, HEAD")
	} else {
		var resp map[string]interface{}
		userId := mux.Vars(r)["id"]
		err := models.DeleteUser(userId)
		if err != "" {
			w.WriteHeader(http.StatusInternalServerError)
			resp = u.Message(false, "Error: "+err)
		} else {
			resp = u.Message(true, "successfully removed user")
		}
		u.Respond(w, resp)
	}
}

var ModifyUserRoles = func(w http.ResponseWriter, r *http.Request) {
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

		rolesConverted := map[string]string{}
		for k := range roles {
			if v, ok := roles[k].(string); ok {
				rolesConverted[k] = v
			} else {
				w.WriteHeader(http.StatusBadRequest)
				u.Respond(w, u.Message(false, "Invalid roles format"))
				return
			}
		}

		e, eType := models.ModifyUser(userId, rolesConverted)
		if e != "" {
			switch eType {
			case "internal":
				w.WriteHeader(http.StatusInternalServerError)
			case "validate":
				w.WriteHeader(http.StatusBadRequest)
			case "unauthorised":
				w.WriteHeader(http.StatusForbidden)
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

var ModifyUserPassword = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 ModifyUserPassword ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	if r.Method == "OPTIONS" {
		w.Header().Add("Content-Type", "application/json")
		w.Header().Add("Allow", "POST, OPTIONS, HEAD")
	} else {
		var resp map[string]interface{}

		// Get userId from token, then get user
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
		if userEmail == "RESET" {
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

var UserForgotPassword = func(w http.ResponseWriter, r *http.Request) {
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

		// Create token, if user exists
		user := models.GetUserByEmail(userEmail)
		token := ""
		if user != nil {
			token = models.GenerateToken("RESET", user.ID, time.Minute*15)
			u.SendEmail(token, user.Email)
		}

		resp["resetToken"] = token
		u.Respond(w, resp)
	}
}
