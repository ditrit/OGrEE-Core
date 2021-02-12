package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

	"github.com/gorilla/mux"
)

// swagger:operation POST /api/user/racks racks Create
// Creates a Rack in the system
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: Name of rack
//   required: true
//   type: string
//   default: "Rack A"
// - name: Category
//   in: query
//   description: Category of Rack (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Rack
//   required: true
//   type: string
//   default: "Some abandoned rack in Grenoble"
// - name: Domain
//   description: 'This an attribute that refers to
//   an existing parent'
//   required: true
//   type: int
//   default: 999
// - name: Color
//   in: query
//   description: Color of Rack (useful for 3D rendering)
//   required: true
//   type: string
//   default: "Silver"
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "NE", "NW", "SE", "SW" are acceptable'
//   required: true
//   type: string
//   default: "NE"
// responses:
//     '200':
//         description: Created
//     '400':
//         description: Bad request
var CreateRack = func(w http.ResponseWriter, r *http.Request) {

	rack := &models.Rack{}
	err := json.NewDecoder(r.Body).Decode(rack)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}

	resp := rack.Create()
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/racks/{id} racks GetRack
// Gets a Rack using Rack ID.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired rack
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '400':
//        description: Not found
var GetRack = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	data := models.GetRack(uint(id))
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation DELETE /api/user/racks/{id} racks DeleteRack
// Deletes a Rack in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired rack
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '400':
//        description: Not found
var DeleteRack = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	v := models.DeleteRack(uint(id))
	u.Respond(w, v)
}

// swagger:operation PUT /api/user/racks/{id} racks UpdateRack
// Changes Rack data in the system.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired rack
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of rack
//   required: false
//   type: string
//   default: "Rack B"
// - name: Category
//   in: query
//   description: Category of Rack (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Rack
//   required: false
//   type: string
//   default: "Some rack"
// - name: Color
//   in: query
//   description: Color of Rack (useful for 3D rendering)
//   required: false
//   type: string
//   default: "Blue"
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "NE", "NW", "SE", "SW" are acceptable'
//   required: false
//   type: string
//   default: "NE"

// responses:
//     '200':
//         description: Updated
//     '400':
//         description: Bad request
//Updates work by passing ID in path parameter
var UpdateRack = func(w http.ResponseWriter, r *http.Request) {

	rack := &models.Rack{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	err := json.NewDecoder(r.Body).Decode(rack)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v := models.UpdateRack(uint(id), rack)
	u.Respond(w, v)
}
