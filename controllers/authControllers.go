package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
)

// swagger:operation POST /api/user auth Create
// Generate credentials for the API.
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
//     '200':
//         description: Authenticated
//     '400':
//         description: Bad request
//     '500':
//         description: Internal server error

var CreateAccount = func(w http.ResponseWriter, r *http.Request) {

	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}
	resp := account.Create()
	w.WriteHeader(http.StatusCreated)
	u.Respond(w, resp)
}

var Authenticate = func(w http.ResponseWriter, r *http.Request) {
	account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}

	resp, e := models.Login(account.Email, account.Password, account.Database)
	if resp["status"] == false {
		if e == "invalid" {
			w.WriteHeader(http.StatusUnauthorized)
		} else if e == "internal" {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	u.Respond(w, resp)
}

var Verify = func(w http.ResponseWriter, r *http.Request) {
	/*account := &models.Account{}
	err := json.NewDecoder(r.Body).Decode(account)
	if err != nil {
		u.Respond(w, u.Message(false, "Invalid request"))
		return
	}*/

	/*var resp map[string]interface{}
	resp["message"] = "Token OK"*/
	u.Respond(w, u.Message(true, "working"))
}
