package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

	"github.com/gorilla/mux"
)

// swagger:operation POST /api/user/rooms rooms Create
// Creates a Room in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: Name of room
//   required: true
//   type: string
//   default: "Room A"
// - name: Category
//   in: query
//   description: Category of Room (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "Room"
// - name: Description
//   in: query
//   description: Description of Room
//   required: false
//   type: string[]
//   default: ["Some abandoned room in Grenoble"]
// - name: Domain
//   description: 'Domain of Room'
//   required: true
//   type: string
//   default: "Some Domain"
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "NE", "NW", "SE", "SW" are acceptable'
//   required: true
//   type: string
//   default: "NE"
// - name: Template
//   in: query
//   description: 'Room template'
//   required: true
//   type: string
//   default: "Some Template"
// - name: PosXY
//   in: query
//   description: 'Indicates the position in a XY coordinate format'
//   required: true
//   type: string
//   default: "{\"x\":-30.0,\"y\":0.0}"
// - name: PosXYU
//   in: query
//   description: 'Indicates the unit of the PosXY position. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: PosZ
//   in: query
//   description: 'Indicates the position in the Z axis'
//   required: true
//   type: string
//   default: "10"
// - name: PosZU
//   in: query
//   description: 'Indicates the unit of the Z coordinate position. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Size
//   in: query
//   description: 'Size of Room in an XY coordinate format'
//   required: true
//   type: string
//   default: "{\"x\":25.0,\"y\":29.399999618530275}"
// - name: SizeU
//   in: query
//   description: 'The unit for Room Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Height
//   in: query
//   description: 'Height of Room'
//   required: true
//   type: string
//   default: "5"
// - name: HeightU
//   in: query
//   description: 'The unit for Room Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"

// responses:
//     '200':
//         description: Created
//     '400':
//         description: Bad request

var CreateRoom = func(w http.ResponseWriter, r *http.Request) {

	room := &models.Room{}
	err := json.NewDecoder(r.Body).Decode(room)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}

	resp, e := room.Create()
	switch e {
	case "validate":
		//
	case "internal":
		//
	default:
		w.WriteHeader(http.StatusCreated)
	}
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/rooms/{id} rooms GetRoom
// Gets Room using Room ID.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of Room
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: Found
//     '400':
//         description: Not Found

//Retrieve room using Room ID
var GetRoom = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	data, e1 := models.GetRoom(uint(id))
	if data == nil {
		resp = u.Message(false, "Error while getting Room: "+e1)

		switch e1 {
		case "validate":
			//
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/rooms rooms GetAllRooms
// Gets All Rooms in the system.
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Found
//     '400':
//         description: Not Found
var GetAllRooms = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data, e1 := models.GetAllRooms()
	if data == nil {
		resp = u.Message(false, "Error while getting Building: "+e1)

		switch e1 {
		case "validate":
			//
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation DELETE /api/user/rooms/{id} rooms DeleteRoom
// Deletes a Room in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired room
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '400':
//        description: Not found

var DeleteRoom = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	v := models.DeleteRoom(uint(id))

	if v["status"] == false {
		//
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	u.Respond(w, v)
}

// swagger:operation PUT /api/user/rooms/{id} rooms UpdateRoom
// Changes Room data in the system.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired room
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of room
//   required: false
//   type: string
//   default: "Room B"
// - name: Category
//   in: query
//   description: Category of Room (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Room
//   required: false
//   type: string[]
//   default: ["Some abandoned room in Grenoble"]
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "NE", "NW", "SE", "SW" are acceptable'
//   required: false
//   type: string
//   default: "NE"
// - name: Template
//   in: query
//   description: 'Room template'
//   required: false
//   type: string
//   default: "New Template"
// - name: PosXY
//   in: query
//   description: 'Indicates the position in a XY coordinate format'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: PosXYU
//   in: query
//   description: 'Indicates the unit of the PosXY position. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: PosZ
//   in: query
//   description: 'Indicates the position in the Z axis'
//   required: false
//   type: string
//   default: "999"
// - name: PosZU
//   in: query
//   description: 'Indicates the unit of the Z coordinate position. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Size
//   in: query
//   description: 'Size of Room in an XY coordinate format'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeU
//   in: query
//   description: 'The unit for Room Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Height
//   in: query
//   description: 'Height of Room'
//   required: false
//   type: string
//   default: "999"
// - name: HeightU
//   in: query
//   description: 'The unit for Room Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// responses:
//     '200':
//         description: Updated
//     '400':
//         description: Bad request

//Updates work by passing ID in path parameter
var UpdateRoom = func(w http.ResponseWriter, r *http.Request) {

	room := &models.Room{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	err := json.NewDecoder(r.Body).Decode(room)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v, e1 := models.UpdateRoom(uint(id), room)

	switch e1 {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
	default:
	}

	u.Respond(w, v)
}
