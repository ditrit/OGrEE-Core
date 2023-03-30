package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"p3/app"
	"p3/models"
	u "p3/utils"
)

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
// - name: customer
//   in: json
//   description: Name of the the customer
//   required: true
//   format: string
//   default: "ORNESS"
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
	var role string
	var domain string

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

		//Extra block here because accounts can be created using
		//an adminPassword or via managers or super users
		if userData := app.ParseToken(w, r); userData != nil {
			role = userData["role"].(string)
			domain = userData["domain"].(string)
		} else {
			role = ""
			domain = ""
		}

		println("Domain:", domain)
		println("Role:", role)

		resp, e := account.Create(role, domain)
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
			if e == "invalid" {
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
