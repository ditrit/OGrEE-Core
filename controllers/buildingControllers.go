package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

	"github.com/gorilla/mux"
)

// swagger:operation POST /api/user/buildings buildings Create
// Creates a Building in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: Name of building
//   required: true
//   type: string
//   default: "Building A"
// - name: Category
//   in: query
//   description: Category of Building (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Building
//   required: true
//   type: string
//   default: "Some abandoned building in Grenoble"
// - name: Domain
//   description: 'This an attribute that refers to
//   an existing parent'
//   required: true
//   type: int
//   default: 999
// - name: Color
//   in: query
//   description: Color of Building (useful for 3D rendering)
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

var CreateBuilding = func(w http.ResponseWriter, r *http.Request) {

	bldg := &models.Building{}
	err := json.NewDecoder(r.Body).Decode(bldg)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}

	resp := bldg.Create()
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/buildings/{id} buildings GetBuilding
// Gets Building using Building ID.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of Building
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: Found
//     '400':
//         description: Not Found

//Retrieve bldg using Bldg ID
var GetBuilding = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	data := models.GetBuilding(uint(id))
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/buildings/ buildings GetBuilding
// Gets All Buildings in the system.
// ---
// produces:
// - application/json
// parameters:
// responses:
//     '200':
//         description: Found
//     '400':
//         description: Not Found

//Retrieve bldg using Bldg ID
var GetAllBuildings = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data := models.GetAllBuildings()
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation DELETE /api/user/buildings/{id} buildings DeleteBuilding
// Deletes a Building.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired building
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '400':
//        description: Not found
var DeleteBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	v := models.DeleteBuilding(uint(id))
	u.Respond(w, v)
}

// swagger:operation PUT /api/user/buildings/{id} buildings UpdateBuilding
// Changes Building data in the system.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired building
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of building
//   required: false
//   type: string
//   default: "Building B"
// - name: Category
//   in: query
//   description: Category of Building (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Building
//   required: false
//   type: string
//   default: "Some abandoned building in Grenoble"
// - name: Color
//   in: query
//   description: Color of Building (useful for 3D rendering)
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
var UpdateBuilding = func(w http.ResponseWriter, r *http.Request) {

	bldg := &models.Building{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	err := json.NewDecoder(r.Body).Decode(bldg)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v := models.UpdateBuilding(uint(id), bldg)
	u.Respond(w, v)
}
