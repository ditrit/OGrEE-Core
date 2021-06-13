package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"p3/models"
	u "p3/utils"
	"reflect"
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
// - name: ParentID
//   description: 'Parent of Device refers to Rack ID'
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
//   description: 'Device template'
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
//   description: 'Size of Device in an XY coordinate format'
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
//   description: 'The unit for Device Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Slot
//   in: query
//   description: 'Device Slot (if any)'
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
//     '201':
//         description: Created
//     '400':
//         description: Bad request
var CreateDevice = func(w http.ResponseWriter, r *http.Request) {

	device := &models.Device{}
	err := json.NewDecoder(r.Body).Decode(device)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE DEVICE", "", r)
		return
	}

	resp, e := device.Create()

	switch e {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while creating Device", "CREATE DEVICE", e, r)
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

// responses:
//     '200':
//         description: Found
//     '400':
//         description: Not Found
var GetDevice = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET DEVICE", "", r)
		return
	}

	data, e1 := models.GetDevice(uint(id))
	if data == nil {
		resp = u.Message(false, "Error while getting Device: "+e1)
		u.ErrLog("Error while getting Device", "GET DEVICE", "", r)

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

// swagger:operation GET /api/user/devices devices GetAllDevices
// Gets all Devices from the system.
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
//     '404':
//         description: Nothing Found
var GetAllDevices = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data, e1 := models.GetAllDevices()
	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e1)
		u.ErrLog("Error while getting devices", "GET ALL DEVICES", e1, r)

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
//     '404':
//        description: Not found
var DeleteDevice = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE DEVICE", "", r)
		return
	}

	v := models.DeleteDevice(uint(id))

	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while deleting device", "DELETE DEVICE", "Not Found", r)
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
// - name: Template
//   in: query
//   description: 'Device template'
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
//   description: 'Size of Device in an XY coordinate format'
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
//   description: 'The unit for Device Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Slot
//   in: query
//   description: 'Device Slot (if any)'
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
//     '404':
//         description: Not Found
//Updates work by passing ID in path parameter
var UpdateDevice = func(w http.ResponseWriter, r *http.Request) {

	device := &models.Device{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "UPDATE DEVICE", "", r)
		return
	}

	err := json.NewDecoder(r.Body).Decode(device)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE DEVICE", "", r)
	}

	v, e1 := models.UpdateDevice(uint(id), device)

	switch e1 {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating device", "UPDATE DEVICE", e1, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating device", "UPDATE DEVICE", e1, r)
	case "record not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating device", "UPDATE DEVICE", e1, r)
	default:
	}

	u.Respond(w, v)
}

// swagger:operation GET /api/user/devices? devices GetDeviceByQuery
// Gets Device By Query.
// Gets a Device using any attribute (with the exception of description) via query
// The attributes are in the form {attr}=xyz&{attr1}=abc
// And any combination can be provided given that at least 1 is provided.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired device
//   required: false
//   type: int
//   default: 999
// - name: Name
//   in: path
//   description: Name of device
//   required: false
//   type: string
//   default: "Device B"
// - name: Category
//   in: path
//   description: Category of Device (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Research"
// - name: Description
//   in: path
//   description: Description of Device
//   required: false
//   type: string[]
//   default: ["Some abandoned device in Grenoble"]
// - name: Template
//   in: path
//   description: 'Device template'
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
//   description: 'Size of Device in an XY coordinate format'
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
//   description: 'The unit for Device Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Slot
//   in: path
//   description: 'Device Slot (if any)'
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
//   description: 'Height of Device'
//   required: false
//   type: string
//   default: "999"
// - name: HeightU
//   in: path
//   description: 'The unit for Device Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Vendor
//   in: path
//   description: 'Vendor of Device'
//   required: false
//   type: string
//   default: "New Vendor"
// - name: Model
//   in: path
//   description: 'Model of Device'
//   required: false
//   type: string
//   default: "New Model"
// - name: Type
//   in: path
//   description: 'Type of Device'
//   required: false
//   type: string
//   default: "New Type"
// - name: Serial
//   in: path
//   description: 'Serial of Device'
//   required: false
//   type: string
//   default: "New Serial"

// responses:
//     '200':
//         description: Found
//     '404':
//         description: Nothing Found
//Updates work by passing ID in path parameter
var GetDeviceByQuery = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}

	query := u.ParamsParse(r.URL)

	mydata := &models.Device{}
	json.Unmarshal(query, mydata)
	json.Unmarshal(query, &(mydata.Attributes))
	fmt.Println("This is the result: ", *mydata)

	if reflect.DeepEqual(&models.Device{}, mydata) {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, `Error while extracting from path parameters. Please check your query parameters.`))
		u.ErrLog("Error while extracting from path parameters", "GET DEVICE BY QUERY",
			"", r)
		return
	}

	data, e := models.GetDeviceByQuery(mydata)

	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting device", "GET DEVICEQUERY", e, r)

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

// swagger:operation GET /api/user/devices/{id}/all devices GetDevice
// Gets Device Hierarchy.
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
//        description: Successful
//     '404':
//        description: Not found
var GetDeviceHierarchy = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET DEVICE", "", r)
		return
	}

	data, e1 := models.GetDeviceHierarchy(id)

	if data == nil {
		resp = u.Message(false, "Error while getting Device: "+e1)
		u.ErrLog("Error while getting Device", "GET DEVICE", e1, r)

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

// swagger:operation GET /api/user/devices/{id}/subdevices/{subdevice_name} devices GetDevice
// Gets a Subdevice of Device.
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
// - name: subdevice_name
//   in: path
//   description: name of desired subdevice
//   required: true
//   type: string
//   default: "Subdevice01"
// responses:
//     '200':
//        description: Successful
//     '404':
//        description: Not found
var GetDeviceSubdeviceByName = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	name := mux.Vars(r)["subdevice_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET DEVICESUBDEVICEBYNAME", "", r)
		return
	}

	data, e1 := models.GetSubdeviceByNameAndParentID(id, name)

	/*data, e1 := models.GetRack(uint(id))*/
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevice: "+e1)
		u.ErrLog("Error while getting Subdevice by name",
			"GET SUBDEVICE USING PID &NAME", e1, r)

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

// swagger:operation GET /api/user/devices/{id}/subdevices/{subdevice_name} devices GetDevice
// Gets a Subdevice of Device.
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
// - name: subdevice_name
//   in: path
//   description: name of desired subdevice
//   required: true
//   type: string
//   default: "Subdevice01"
// responses:
//     '200':
//        description: Successful
//     '404':
//        description: Not found
var GetNamedSubdevice1OfDevice = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	subdev_name := mux.Vars(r)["subdevice_name"]
	subdev1_name := mux.Vars(r)["subdevone_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDSUBDEV1OFDEV", "", r)
		return
	}

	data, e1 := models.GetNamedSubdevice1OfDevice(id, subdev_name, subdev1_name)

	if data == nil {
		resp = u.Message(false, "Error while getting Subdevice1: "+e1)
		u.ErrLog("Error while getting Subdevice1 by name",
			"GET NAMEDSUBDEV1OFDEV", e1, r)

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

// swagger:operation GET /api/user/devices/{id}/subdevices devices GetDevice
// Gets Subdevices of Device.
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
//        description: Successful
//     '404':
//        description: Not found
var GetSubdevicesOfDevice = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET SUBDEVSOFDEVICE", "", r)
		return
	}

	data, e1 := models.GetSubdevicesOfDevice(id)
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevices: "+e1)
		u.ErrLog("Error while getting Subdevices of Device",
			"GET SUBDEVSOFDEVICE", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}
