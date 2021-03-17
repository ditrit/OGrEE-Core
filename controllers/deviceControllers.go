package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

	"github.com/gorilla/mux"
)

// swagger:operation POST /api/user/devices devices CreateDevice
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
//   required: false
//   type: string[]
//   default: ["Some abandoned device in Grenoble"]
// - name: Domain
//   description: 'Domain of Device'
//   required: true
//   type: string
//   default: "Some Domain"
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "front", "rear", "frontflipped", "rearflipped" are acceptable'
//   required: true
//   type: string
//   default: "front"
// - name: Template
//   in: query
//   description: 'Device template'
//   required: false
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
//   description: 'Size of Device in an XY coordinate format'
//   required: true
//   type: string
//   default: "{\"x\":25.0,\"y\":29.399999618530275}"
// - name: SizeU
//   in: query
//   description: 'The unit for Device Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Height
//   in: query
//   description: 'Height of Device'
//   required: true
//   type: string
//   default: "5"
// - name: HeightU
//   in: query
//   description: 'The unit for Device Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Vendor
//   in: query
//   description: 'Vendor of Device'
//   required: false
//   type: string
//   default: "Some Vendor"
// - name: Model
//   in: query
//   description: 'Model of Device'
//   required: false
//   type: string
//   default: "Some Model"
// - name: Type
//   in: query
//   description: 'Type of Device'
//   required: false
//   type: string
//   default: "Some Type"
// - name: Serial
//   in: query
//   description: 'Serial of Device'
//   required: false
//   type: string
//   default: "Some Serial"

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

	resp, e := device.Create()

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

// swagger:operation GET /api/user/devices/{id} devices GetDevice
// Gets Device using Device ID.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of Device

var GetDevice = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	data, e1 := models.GetDevice(uint(id))
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

// swagger:operation GET /api/user/devices/{id} devices GetDevice
// Gets a Device from the system.
// The ID must be provided in the URL parameter.
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
//     '200':
//         description: Found
//     '400':
//         description: Bad request
var GetAllDevices = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data, e1 := models.GetAllDevices()
	if data == nil {
		resp = u.Message(false, "Error while getting Rack: "+e1)

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

	if v["status"] == false {
		//
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

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
//   type: string[]
//   default: ["Some abandoned device in Grenoble"]
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "front", "rear", "frontflipped", "rearflipped" are acceptable'
//   required: false
//   type: string
//   default: "frontflipped"
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
//   description: 'Size of Device in an XY coordinate format'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeU
//   in: query
//   description: 'The unit for Device Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Height
//   in: query
//   description: 'Height of Device'
//   required: false
//   type: string
//   default: "999"
// - name: HeightU
//   in: query
//   description: 'The unit for Device Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Vendor
//   in: query
//   description: 'Vendor of Device'
//   required: false
//   type: string
//   default: "New Vendor"
// - name: Model
//   in: query
//   description: 'Model of Device'
//   required: false
//   type: string
//   default: "New Model"
// - name: Type
//   in: query
//   description: 'Type of Device'
//   required: false
//   type: string
//   default: "New Type"
// - name: Serial
//   in: query
//   description: 'Serial of Device'
//   required: false
//   type: string
//   default: "New Serial"

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

	v, e1 := models.UpdateDevice(uint(id), device)

	switch e1 {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
	default:
	}

	u.Respond(w, v)
}
