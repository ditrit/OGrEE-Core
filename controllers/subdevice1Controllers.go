package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

	"github.com/gorilla/mux"
)

// swagger:operation POST /api/user/subdevices1 subdevices1 CreateSubdevice1
// Creates a Subdevice1 in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: Name of subdevice1
//   required: true
//   type: string
//   default: "Subdevice1A"
// - name: Category
//   in: query
//   description: Category of Subdevice1 (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "internal"
// - name: Description
//   in: query
//   description: Description of Subdevice1
//   required: false
//   type: string[]
//   default: ["Some abandoned subdevice1 in Grenoble"]
// - name: Domain
//   description: 'Domain of Subdevice1'
//   required: true
//   type: string
//   default: "Some Domain"
// - name: ParentID
//   description: 'Parent of Subdevice1 refers to Rack ID'
//   required: true
//   type: int
//   default: 999
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "front", "rear", "frontflipped", "rearflipped" are acceptable'
//   required: true
//   type: string
//   default: "front"
// - name: Template
//   in: query
//   description: 'Subdevice1 template'
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
//   description: 'Size of Subdevice in an XY coordinate format'
//   required: true
//   type: string
//   default: "{\"x\":25.0,\"y\":29.399999618530275}"
// - name: SizeUnit
//   in: query
//   description: 'Extraneous Size Unit Attribute'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeU
//   in: query
//   description: 'The unit for Subdevice1 Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Slot
//   in: query
//   description: 'Subdevice1 Slot (if any)'
//   required: false
//   type: string
//   default: "01"
// - name: PosU
//   in: query
//   description: 'Extraneous Position Unit Attribute'
//   required: false
//   type: string
//   default: "???"
// - name: Height
//   in: query
//   description: 'Height of Subdevice1'
//   required: true
//   type: string
//   default: "5"
// - name: HeightU
//   in: query
//   description: 'The unit for Subdevice1 Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Vendor
//   in: query
//   description: 'Vendor of Subdevice1'
//   required: false
//   type: string
//   default: "Some Vendor"
// - name: Model
//   in: query
//   description: 'Model of Subdevice1'
//   required: false
//   type: string
//   default: "Some Model"
// - name: Type
//   in: query
//   description: 'Type of Subdevice1'
//   required: false
//   type: string
//   default: "Some Type"
// - name: Serial
//   in: query
//   description: 'Serial of Subdevice1'
//   required: false
//   type: string
//   default: "Some Serial"

// responses:
//     '201':
//         description: Created
//     '400':
//         description: Bad request
var CreateSubdevice1 = func(w http.ResponseWriter, r *http.Request) {

	subdevice1 := &models.Subdevice1{}
	err := json.NewDecoder(r.Body).Decode(subdevice1)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE SUBDEVICE1", "", r)
		return
	}

	resp, e := subdevice1.Create()

	switch e {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while creating Subdevice1", "CREATE SUBDEVICE1", e, r)
	case "internal":
		//
	default:
		w.WriteHeader(http.StatusCreated)
	}
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/subdevices1/{id} subdevices1 GetSubdevice1
// Gets Subdevice1 using Subdevice1 ID.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of Subdevice1

// responses:
//     '200':
//         description: Found
//     '400':
//         description: Not Found
var GetSubdevice1 = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET SUBDEVICE1", "", r)
		return
	}

	data, e1 := models.GetSubdevice1(id)
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevice1: "+e1)
		u.ErrLog("Error while getting Subdevice", "GET SUBDEVICE1", "", r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/subdevices1 subdevices1 GetAllSubdevices1
// Gets all Subdevices1 from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired subdevice1
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Nothing Found
var GetAllSubdevices1 = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data, e1 := models.GetAllSubdevices1()
	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e1)
		u.ErrLog("Error while getting devices", "GET ALL SUBDEVICES1", e1, r)

		switch e1 {
		case "":
			resp = u.Message(false, "Error: No Records Found")
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}
	u.Respond(w, resp)
}

// swagger:operation DELETE /api/user/subdevices1/{id} subdevices1 DeleteSubdevice1
// Deletes a Subdevice1 in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired subdevice1
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '404':
//        description: Not found
var DeleteSubdevice1 = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE SUBDEVICE1", "", r)
		return
	}

	v := models.DeleteSubdevice1(id)

	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while deleting subdevice1", "DELETE SUBDEVICE1", "Not Found", r)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	u.Respond(w, v)
}

// swagger:operation PUT /api/user/subdevices1/{id} subdevices1 UpdateSubdevice1
// Changes Subdevice1 data in the system.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired subdevice1
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of subdevice1
//   required: false
//   type: string
//   default: "Subdevice1 B"
// - name: Category
//   in: query
//   description: Category of Subdevice1 (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Subdevice1
//   required: false
//   type: string[]
//   default: ["Some abandoned subdevice1 in Grenoble"]
// - name: Template
//   in: query
//   description: 'Subdevice1 template'
//   required: false
//   type: string
//   default: "Some Template"
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
//   description: 'Size of Subdevice1 in an XY coordinate format'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeUnit
//   in: query
//   description: 'Extraneous Size Unit Attribute'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeU
//   in: query
//   description: 'The unit for Subdevice1 Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Slot
//   in: query
//   description: 'Subdevice1 Slot (if any)'
//   required: false
//   type: string
//   default: "01"
// - name: PosU
//   in: query
//   description: 'Extraneous Position Unit Attribute'
//   required: false
//   type: string
//   default: "???"
// - name: Height
//   in: query
//   description: 'Height of Subdevice1'
//   required: false
//   type: string
//   default: "999"
// - name: HeightU
//   in: query
//   description: 'The unit for Subdevice1 Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Vendor
//   in: query
//   description: 'Vendor of Subdevice1'
//   required: false
//   type: string
//   default: "New Vendor"
// - name: Model
//   in: query
//   description: 'Model of Subdevice1'
//   required: false
//   type: string
//   default: "New Model"
// - name: Type
//   in: query
//   description: 'Type of Subdevice1'
//   required: false
//   type: string
//   default: "New Type"
// - name: Serial
//   in: query
//   description: 'Serial of Subdevice1'
//   required: false
//   type: string
//   default: "New Serial"

// responses:
//     '200':
//         description: Updated
//     '400':
//         description: Bad request
//     '404':
//         description: Not Found
//Updates work by passing ID in path parameter
var UpdateSubdevice1 = func(w http.ResponseWriter, r *http.Request) {

	subdevice1 := &models.Subdevice1{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "UPDATE SUBDEVICE1", "", r)
		return
	}

	err := json.NewDecoder(r.Body).Decode(subdevice1)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE SUBDEVICE1", "", r)
	}

	v, e1 := models.UpdateSubdevice1(id, subdevice1)

	switch e1 {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating subdevice", "UPDATE SUBDEVICE1", e1, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating subdevice", "UPDATE SUBDEVICE1", e1, r)
	case "record not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating subdevice", "UPDATE SUBDEVICE1", e1, r)
	default:
	}

	u.Respond(w, v)
}
