package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

	"github.com/gorilla/mux"
)

// swagger:operation POST /api/user/devices devices Create
// Creates a Device in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: Name of device
//   required: true
//   type: string
//   default: "Device A"
// - name: Category
//   in: query
//   description: Category of Device (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Device
//   required: true
//   type: string
//   default: "Some abandoned device in Grenoble"
// - name: Domain
//   description: 'This an attribute that refers to
//   an existing parent'
//   required: true
//   type: int
//   default: 999
// - name: Color
//   in: query
//   description: Color of Device (useful for 3D rendering)
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
var CreateDevice = func(w http.ResponseWriter, r *http.Request) {

	device := &models.Device{}
	err := json.NewDecoder(r.Body).Decode(device)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}

	resp := device.Create()
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/devices/{id} devices GetDevice
// Gets Device using Device ID.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of Device

//Retrieve device using Device ID
var GetDevice = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	data := models.GetDevice(uint(id))
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/devices devices GetDevice
// Gets All Devices in the system.
// ---
// produces:
// - application/json
// responses:
//     '204':
//        description: Successful
//     '400':
//        description: Not found
//Retrieve device using Device ID
var GetAllDevices = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data := models.GetAllDevices()
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation DELETE /api/user/devices/{id} devices DeleteDevice
// Deletes a Device in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired device
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '400':
//        description: Not found
var DeleteDevice = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	v := models.DeleteDevice(uint(id))
	u.Respond(w, v)
}

// swagger:operation PUT /api/user/devices/{id} devices UpdateDevice
// Changes Device data in the system.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired device
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of device
//   required: false
//   type: string
//   default: "Device B"
// - name: Category
//   in: query
//   description: Category of Device (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Device
//   required: false
//   type: string
//   default: "Some abandoned device in Grenoble"
// - name: Color
//   in: query
//   description: Color of Device (useful for 3D rendering)
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
var UpdateDevice = func(w http.ResponseWriter, r *http.Request) {

	device := &models.Device{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	err := json.NewDecoder(r.Body).Decode(device)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v := models.UpdateDevice(uint(id), device)
	u.Respond(w, v)
}
