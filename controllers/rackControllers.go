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

// swagger:operation POST /api/user/racks racks Create
// Creates a Rack in the system.
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
//   required: false
//   type: string[]
//   default: "Some abandoned rack in Grenoble"
// - name: Domain
//   description: 'Domain of Rack'
//   required: true
//   type: string
//   default: "Some Domain"
// - name: ParentID
//   description: 'ParentID of Rack refers to Room'
//   required: true
//   type: string
//   default: "999"
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "front", "rear", "left", "right" are acceptable'
//   required: true
//   type: string
//   default: "front"
// - name: Template
//   in: query
//   description: 'Room template'
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
//   required: false
//   type: string
//   default: "10"
// - name: PosZU
//   in: query
//   description: 'Indicates the unit of the Z coordinate position. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "m"
// - name: Size
//   in: query
//   description: 'Size of Rack in an XY coordinate format'
//   required: true
//   type: string
//   default: "{\"x\":25.0,\"y\":29.399999618530275}"
// - name: SizeU
//   in: query
//   description: 'The unit for Rack Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Height
//   in: query
//   description: 'Height of Rack'
//   required: true
//   type: string
//   default: "5"
// - name: HeightU
//   in: query
//   description: 'The unit for Rack Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Vendor
//   in: query
//   description: 'Vendor of Rack'
//   required: false
//   type: string
//   default: "Some Vendor"
// - name: Model
//   in: query
//   description: 'Model of Rack'
//   required: false
//   type: string
//   default: "Some Model"
// - name: Type
//   in: query
//   description: 'Type of Rack'
//   required: false
//   type: string
//   default: "Some Type"
// - name: Serial
//   in: query
//   description: 'Serial of Rack'
//   required: false
//   type: string
//   default: "Some Serial"

// responses:
//     '201':
//         description: Created
//     '400':
//         description: Bad request
var CreateRack = func(w http.ResponseWriter, r *http.Request) {

	rack := &models.Rack{}
	err := json.NewDecoder(r.Body).Decode(rack)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE RACK", "", r)
		return
	}

	resp, e := rack.Create()

	switch e {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while creating rack", "CREATE RACK", e, r)
	case "internal":
		//
	default:
		w.WriteHeader(http.StatusCreated)
	}
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
//     '200':
//        description: Successful
//     '404':
//        description: Not found
var GetRack = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET RACK", "", r)
	}

	data, e1 := models.GetRack(uint(id))
	if data == nil {
		resp = u.Message(false, "Error while getting Rack: "+e1)
		u.ErrLog("Error while getting Rack", "GET RACK", e1, r)

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

// swagger:operation GET /api/user/racks/ racks GetRack
// Gets All Racks in the system.
// ---
// produces:
// - application/json
// responses:
//     '200':
//        description: Successful
//     '404':
//        description: Not found
var GetAllRacks = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data, e1 := models.GetAllRacks()
	if len(data) == 0 {
		resp = u.Message(false, "Error while getting Rack: "+e1)
		u.ErrLog("Error while getting Rack", "GET ALL RACKS", e1, r)

		switch e1 {
		case "":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
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
//     '404':
//        description: Not found
var DeleteRack = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE RACK", "", r)
	}

	v := models.DeleteRack(uint(id))

	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while deleting rack", "DELETE RACK", "Not Found", r)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

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
// - name: Template
//   in: query
//   description: 'Room template'
//   required: false
//   type: string
//   default: "Some Template"
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "front", "rear", "left", "right" are acceptable'
//   required: false
//   type: string
//   default: "front"
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
//   description: 'Size of Rack in an XY coordinate format'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeU
//   in: query
//   description: 'The unit for Rack Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Height
//   in: query
//   description: 'Height of Rack'
//   required: false
//   type: string
//   default: "999"
// - name: HeightU
//   in: query
//   description: 'The unit for Rack Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Vendor
//   in: query
//   description: 'Vendor of Rack'
//   required: false
//   type: string
//   default: "New Vendor"
// - name: Model
//   in: query
//   description: 'Model of Rack'
//   required: false
//   type: string
//   default: "New Model"
// - name: Type
//   in: query
//   description: 'Type of Rack'
//   required: false
//   type: string
//   default: "New Type"
// - name: Serial
//   in: query
//   description: 'Serial of Rack'
//   required: false
//   type: string
//   default: "New Serial"

// responses:
//     '200':
//         description: Updated
//     '404':
//         description: Not Found
//     '400':
//         description: Bad request
//Updates work by passing ID in path parameter
var UpdateRack = func(w http.ResponseWriter, r *http.Request) {

	rack := &models.Rack{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "UPDATE RACK", "", r)
	}

	err := json.NewDecoder(r.Body).Decode(rack)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE RACK", "", r)
	}

	v, e1 := models.UpdateRack(uint(id), rack)

	switch e1 {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating rack", "UPDATE RACK", e1, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating rack", "UPDATE RACK", e1, r)
	case "record not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating rack", "UPDATE RACK", e1, r)
	default:
	}

	u.Respond(w, v)
}

var GetRackByName = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}
	names := strings.Split(r.URL.String(), "=")

	if names[1] == "" {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "GET RACK BY NAME",
			"", r)
		return
	}

	data, e := models.GetRackByName(names[1])

	if e != "" {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting rack", "GET Rack", e, r)

		switch e {
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

var GetRackHierarchy = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("me & the irishman")
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET RACK", "", r)
	}

	data, e1 := models.GetRackHierarchy(uint(id))

	if data == nil {
		resp = u.Message(false, "Error while getting Rack: "+e1)
		u.ErrLog("Error while getting Rack", "GET RACK", e1, r)

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

var GetRackDeviceByName = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	name := mux.Vars(r)["device_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET RACKDEVICEBYNAME", "", r)
	}

	data, e1 := models.GetDeviceByNameAndParentID(uint(id), name)

	/*data, e1 := models.GetRack(uint(id))*/
	if data == nil {
		resp = u.Message(false, "Error while getting Device: "+e1)
		u.ErrLog("Error while getting Device by name",
			"GET DEVICE USING PID &NAME", e1, r)

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

var GetRackByQuery = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}

	query := u.ParamsParse(r.URL)

	mydata := &models.Rack{}
	json.Unmarshal(query, mydata)
	json.Unmarshal(query, &(mydata.Attributes))
	fmt.Println("This is the result: ", *mydata)

	if reflect.DeepEqual(&models.Rack{}, mydata) {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from "+
			"path parameters. Please check your query parameters."))
		u.ErrLog("Error while extracting from path parameters",
			"GET RACK BY QUERY", "", r)
		return
	}

	data, e := models.GetRackByQuery(mydata)

	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting rack", "GET RACKQUERY", e, r)

		switch e {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	} else {
		resp = u.Message(true, "success")
	}

	if len(data) == 1 {
		resp["data"] = data[0]
	} else {
		resp["data"] = data
	}
	u.Respond(w, resp)
}
