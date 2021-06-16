package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"p3/models"
	u "p3/utils"
	"reflect"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// swagger:operation POST /api/user/subdevices subdevices CreateSubdevice
// Creates a Subdevice in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: Name of subdevice
//   required: true
//   type: string
//   default: "Subdevice A"
// - name: Category
//   in: query
//   description: Category of Subdevice (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "internal"
// - name: Description
//   in: query
//   description: Description of Subdevice
//   required: false
//   type: string[]
//   default: ["Some abandoned subdevice in Grenoble"]
// - name: Domain
//   description: 'Domain of Subdevice'
//   required: true
//   type: string
//   default: "Some Domain"
// - name: ParentID
//   description: 'Parent of Subdevice refers to Rack ID'
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
//   description: 'Subdevice template'
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
//   description: 'The unit for Subdevice Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Slot
//   in: query
//   description: 'Subdevice Slot (if any)'
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
//   description: 'Height of Subdevice'
//   required: true
//   type: string
//   default: "5"
// - name: HeightU
//   in: query
//   description: 'The unit for Subdevice Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Vendor
//   in: query
//   description: 'Vendor of Subdevice'
//   required: false
//   type: string
//   default: "Some Vendor"
// - name: Model
//   in: query
//   description: 'Model of Subdevice'
//   required: false
//   type: string
//   default: "Some Model"
// - name: Type
//   in: query
//   description: 'Type of Subdevice'
//   required: false
//   type: string
//   default: "Some Type"
// - name: Serial
//   in: query
//   description: 'Serial of Subdevice'
//   required: false
//   type: string
//   default: "Some Serial"

// responses:
//     '201':
//         description: Created
//     '400':
//         description: Bad request
var CreateSubdevice = func(w http.ResponseWriter, r *http.Request) {

	subdevice := &models.Subdevice{}
	err := json.NewDecoder(r.Body).Decode(subdevice)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE SUBDEVICE", "", r)
		return
	}

	resp, e := subdevice.Create()

	switch e {
	case "":
		w.WriteHeader(http.StatusCreated)
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while creating Subdevice", "CREATE SUBDEVICE", e, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog(e+" Error", "CREATE SUBDEVICE", "", r)
	default:
		if strings.Split(e, " ")[1] == "duplicate" {
			w.WriteHeader(http.StatusBadRequest)
			u.ErrLog("Error: Duplicate Subdev is forbidden",
				"CREATE SUBDEVICE", e, r)
		} else {
			w.WriteHeader(http.StatusCreated)
		}
	}
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/subdevices/{id} subdevices GetSubdevice
// Gets Subdevice using Subdevice ID.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of Subdevice

// responses:
//     '200':
//         description: Found
//     '400':
//         description: Not Found
var GetSubdevice = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET SUBDEVICE", "", r)
		return
	}

	data, e1 := models.GetSubdevice(uint(id))
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevice: "+e1)
		u.ErrLog("Error while getting Subdevice", "GET SUBDEVICE", "", r)

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

// swagger:operation GET /api/user/subdevices devices GetAllSubdevices
// Gets all Subdevices from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired subdevice
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Nothing Found
var GetAllSubdevices = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data, e1 := models.GetAllSubdevices()
	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e1)
		u.ErrLog("Error while getting devices", "GET ALL SUBDEVICES", e1, r)

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

// swagger:operation DELETE /api/user/subdevices/{id} subdevices DeleteSubdevice
// Deletes a Subdevice in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired subdevice
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '404':
//        description: Not found
var DeleteSubdevice = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE SUBDEVICE", "", r)
		return
	}

	v := models.DeleteSubdevice(uint(id))

	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while deleting subdevice", "DELETE SUBDEVICE", "Not Found", r)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

	u.Respond(w, v)
}

// swagger:operation PUT /api/user/subdevices/{id} subdevices UpdateSubdevice
// Changes Subdevice data in the system.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired subdevice
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of subdevice
//   required: false
//   type: string
//   default: "Subdevice B"
// - name: Category
//   in: query
//   description: Category of Subdevice (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Subdevice
//   required: false
//   type: string[]
//   default: ["Some abandoned subdevice in Grenoble"]
// - name: Template
//   in: query
//   description: 'Subdevice template'
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
//   description: 'Size of Subdevice in an XY coordinate format'
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
//   description: 'The unit for Subdevice Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Slot
//   in: query
//   description: 'Subdevice Slot (if any)'
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
//   description: 'Height of Subdevice'
//   required: false
//   type: string
//   default: "999"
// - name: HeightU
//   in: query
//   description: 'The unit for Subdevice Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Vendor
//   in: query
//   description: 'Vendor of Subdevice'
//   required: false
//   type: string
//   default: "New Vendor"
// - name: Model
//   in: query
//   description: 'Model of Subdevice'
//   required: false
//   type: string
//   default: "New Model"
// - name: Type
//   in: query
//   description: 'Type of Subdevice'
//   required: false
//   type: string
//   default: "New Type"
// - name: Serial
//   in: query
//   description: 'Serial of Subdevice'
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
var UpdateSubdevice = func(w http.ResponseWriter, r *http.Request) {

	subdevice := &models.Subdevice{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "UPDATE SUBDEVICE", "", r)
		return
	}

	err := json.NewDecoder(r.Body).Decode(subdevice)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE SUBDEVICE", "", r)
	}

	v, e1 := models.UpdateSubdevice(uint(id), subdevice)

	switch e1 {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating subdevice", "UPDATE SUBDEVICE", e1, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating subdevice", "UPDATE SUBDEVICE", e1, r)
	case "record not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating subdevice", "UPDATE SUBDEVICE", e1, r)
	default:
	}

	u.Respond(w, v)
}

// swagger:operation GET /api/user/subdevices? subdevices GetSubdeviceByQuery
// Gets Subdevice By Query.
// Gets a Subdevice using any attribute (with the exception of description) via query
// The attributes are in the form {attr}=xyz&{attr1}=abc
// And any combination can be provided given that at least 1 is provided.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired subdevice
//   required: false
//   type: int
//   default: 999
// - name: Name
//   in: path
//   description: Name of subdevice
//   required: false
//   type: string
//   default: "Subdevice B"
// - name: Category
//   in: path
//   description: Category of Subdevice (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "internal"
// - name: Description
//   in: path
//   description: Description of Subdevice
//   required: false
//   type: string[]
//   default: ["Some abandoned subdevice in Grenoble"]
// - name: Template
//   in: path
//   description: 'Subdevice template'
//   required: false
//   type: string
//   default: "Some Template"
// - name: Orientation
//   in: path
//   description: 'Indicates the location. Only values of
//   "front", "rear", "frontflipped", "rearflipped" are acceptable'
//   required: false
//   type: string
//   default: "frontflipped"
// - name: PosXY
//   in: path
//   description: 'Indicates the position in a XY coordinate format'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: PosXYU
//   in: path
//   description: 'Indicates the unit of the PosXY position. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: PosZ
//   in: path
//   description: 'Indicates the position in the Z axis'
//   required: false
//   type: string
//   default: "999"
// - name: PosZU
//   in: path
//   description: 'Indicates the unit of the Z coordinate position. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Size
//   in: path
//   description: 'Size of Subdevice in an XY coordinate format'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeUnit
//   in: path
//   description: 'Extraneous Size Unit Attribute'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeU
//   in: path
//   description: 'The unit for Subdevice Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Slot
//   in: path
//   description: 'Subdevice Slot (if any)'
//   required: false
//   type: string
//   default: "01"
// - name: PosU
//   in: path
//   description: 'Extraneous Position Unit Attribute'
//   required: false
//   type: string
//   default: "???"
// - name: Height
//   in: path
//   description: 'Height of Subdevice'
//   required: false
//   type: string
//   default: "999"
// - name: HeightU
//   in: path
//   description: 'The unit for Subdevice Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Vendor
//   in: path
//   description: 'Vendor of Subdevice'
//   required: false
//   type: string
//   default: "New Vendor"
// - name: Model
//   in: path
//   description: 'Model of Subdevice'
//   required: false
//   type: string
//   default: "New Model"
// - name: Type
//   in: path
//   description: 'Type of Subdevice'
//   required: false
//   type: string
//   default: "New Type"
// - name: Serial
//   in: path
//   description: 'Serial of Subdevice'
//   required: false
//   type: string
//   default: "New Serial"

// responses:
//     '200':
//         description: Found
//     '404':
//         description: Nothing Found
//Updates work by passing ID in path parameter
var GetSubdeviceByQuery = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}

	query := u.ParamsParse(r.URL)

	mydata := &models.Subdevice{}
	json.Unmarshal(query, mydata)
	json.Unmarshal(query, &(mydata.Attributes))
	fmt.Println("This is the result: ", *mydata)

	if reflect.DeepEqual(&models.Subdevice{}, mydata) {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, `Error while extracting from path parameters. Please check your query parameters.`))
		u.ErrLog("Error while extracting from path parameters", "GET SUBDEVICE BY QUERY",
			"", r)
		return
	}

	data, e := models.GetSubdeviceByQuery(mydata)

	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting subdevice", "GET SUBDEVICEQUERY", e, r)

		switch e {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			resp = u.Message(false, "Error: No Records Found")
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/subdevices/{id}/all subdevices GetSubdevice
// Gets Subdevice Hierarchy.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired subdevice
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//        description: Successful
//     '404':
//        description: Not found
var GetSubdeviceHierarchy = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET SUBDEVICEHIERARCHY", "", r)
		return
	}

	data, e1 := models.GetSubdeviceHierarchy(id)

	if data == nil {
		resp = u.Message(false, "Error while getting Subdevice: "+e1)
		u.ErrLog("Error while getting Device", "GET SUBDEVICEHIERARCHY", e1, r)

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
