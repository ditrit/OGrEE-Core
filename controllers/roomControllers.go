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
// - name: ParentID
//   description: 'ParentID of Room refers to Building'
//   required: true
//   type: string
//   default: "999"
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   (-|+)E(-|+)N, (-|+)N(-|+)W,
//   (-|+)W(-|+)S, (-|+)S(-|+)E
//   are acceptable'
//   required: true
//   type: string
//   default: "NE"
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
//     '201':
//         description: Created
//     '400':
//         description: Bad request
//     '404':
//         description: Not Found

var CreateRoom = func(w http.ResponseWriter, r *http.Request) {

	room := &models.Room{}
	err := json.NewDecoder(r.Body).Decode(room)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE ROOM", "", r)
		return
	}

	resp, e := room.Create()
	switch e {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while creating room", "CREATE ROOM", e, r)
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
//     '404':
//         description: Not Found

//Retrieve room using Room ID
var GetRoom = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ROOM", "", r)
		return
	}

	data, e1 := models.GetRoom(uint(id))
	if data == nil {
		resp = u.Message(false, "Error while getting Room: "+e1)
		u.ErrLog("Error while getting Room: ", "GET ROOM", e1, r)

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

// swagger:operation GET /api/user/rooms rooms GetAllRooms
// Gets All Rooms in the system.
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetAllRooms = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data, e1 := models.GetAllRooms()
	if len(data) == 0 {
		resp = u.Message(false, "Error while getting Room: "+e1)
		u.ErrLog("Error while getting Room: ", "GET ALL ROOMS", e1, r)

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
//     '404':
//        description: Not found

var DeleteRoom = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE ROOM", "", r)
		return
	}

	v := models.DeleteRoom(uint(id))

	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while deleting room", "DELETE ROOM", "Not Found", r)
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
//   (-|+)E(-|+)N, (-|+)N(-|+)W,
//   (-|+)W(-|+)S, (-|+)S(-|+)E
//   are acceptable'
//   required: false
//   type: string
//   default: "+N+E"
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
// - name: Technical
//   in: query
//   description: 'Unity Unique Attribute'
//   required: false
//   type: string
//   default: "???"
// - name: Reserved
//   in: query
//   description: 'Unity Unique Attribute'
//   required: false
//   type: string
//   default: "???"
// responses:
//     '200':
//         description: Updated
//     '404':
//         description: Not Found
//     '400':
//         description: Bad request

//Updates work by passing ID in path parameter
var UpdateRoom = func(w http.ResponseWriter, r *http.Request) {

	room := &models.Room{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "UPDATE ROOM", "", r)
		return
	}

	err := json.NewDecoder(r.Body).Decode(room)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE ROOM", "", r)
	}

	v, e1 := models.UpdateRoom(uint(id), room)

	switch e1 {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating room", "UPDATE ROOM", e1, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating room", "UPDATE ROOM", e1, r)
	case "record not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating room", "UPDATE ROOM", e1, r)
	default:
	}

	u.Respond(w, v)
}

// swagger:operation GET /api/user/rooms? rooms GetRoomByQuery
// Gets a Room using any attribute (with the exception of description) via query
// The attributes are in the form {attr}=xyz&{attr1}=abc
// And any combination can be provided given that at least 1 is provided.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired room
//   required: false
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
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   (-|+)E(-|+)N, (-|+)N(-|+)W,
//   (-|+)W(-|+)S, (-|+)S(-|+)E
//   are acceptable'
//   required: false
//   type: string
//   default: "+N+E"
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
// - name: Technical
//   in: query
//   description: 'Unity Unique Attribute'
//   required: false
//   type: string
//   default: "???"
// - name: Reserved
//   in: query
//   description: 'Unity Unique Attribute'
//   required: false
//   type: string
//   default: "???"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetRoomByQuery = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}

	query := u.ParamsParse(r.URL)

	mydata := &models.Room{}
	json.Unmarshal(query, mydata)
	json.Unmarshal(query, &(mydata.Attributes))
	fmt.Println("This is the result: ", *mydata)

	if reflect.DeepEqual(&models.Room{}, mydata) {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from "+
			"path parameters. Please check your query parameters."))
		u.ErrLog("Error while extracting from path parameters",
			"GET ROOM BY QUERY", "", r)
		return
	}

	data, e := models.GetRoomByQuery(mydata)

	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting room", "GET ROOMQUERY", e, r)

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

/*
var GetRoomByName = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}
	names := strings.Split(r.URL.String(), "=")

	if names[1] == "" {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "GET ROOM BY NAME",
			"", r)
		return
	}

	data, e := models.GetRoomByName(names[1])

	if e != "" {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting room", "GET Room", e, r)

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
*/

// swagger:operation GET /api/user/rooms/{id}/all rooms GetAllRooms
// Gets Room Hierarchy.
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
//     '404':
//         description: Not Found

var GetRoomHierarchy = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("me & the irishman2")
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ROOM", "", r)
		return
	}

	data, e1 := models.GetRoomHierarchy(uint(id))

	if data == nil {
		resp = u.Message(false, "Error while getting Room: "+e1)
		u.ErrLog("Error while getting Room", "GET ROOM", e1, r)

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

// swagger:operation GET /api/user/rooms/{id}/all/racks/devices rooms GetAllRooms
// Gets Room Hierarchy to Devices.
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
//     '404':
//         description: Not Found
var GetRoomHierarchyToDevices = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ROOMHIERARCHYTODEVICES", "", r)
		return
	}

	data, e1 := models.GetRoomHierarchyToDevices(id)

	if data == nil {
		resp = u.Message(false, "Error while getting Room Hierarchy: "+e1)
		u.ErrLog("Error while getting Room", "GET ROOMHIERARCHYTODEVICES", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

var GetRoomHierarchyNonStandard = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("me & the irishman2")
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ROOM", "", r)
		return
	}

	data, racks, devices, e1 := models.GetRoomHierarchyNonStandard(uint(id))

	if data == nil {
		resp = u.Message(false, "Error while getting Room: "+e1)
		u.ErrLog("Error while getting Room", "GET ROOM", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	resp["racks"] = racks
	resp["devices"] = devices
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/rooms/{id}/racks/{rack_name} rooms GetAllRooms
// Gets Rack Of Room.
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
// - name: rack_name
//   in: path
//   description: Name of desired rack
//   required: true
//   type: string
//   default: "Rack01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetRackOfRoomByName = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	name := mux.Vars(r)["rack_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ROOMRACKBYNAME", "", r)
		return
	}

	data, e1 := models.GetRackByNameAndParentID(id, name)

	/*data, e1 := models.GetRack(uint(id))*/
	if data == nil {
		resp = u.Message(false, "Error while getting Rack: "+e1)
		u.ErrLog("Error while getting Rack by name",
			"GET ROOMRACK USING PID &NAME", e1, r)

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

// swagger:operation GET /api/user/rooms/{id}/racks rooms GetAllRooms
// Gets All Racks of Room.
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
//     '404':
//         description: Not Found

var GetRacksOfParent = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET RACKSOFPARENT", "", r)
		return
	}

	data, e1 := models.GetRacksOfParent(uint(id))
	if data == nil {
		resp = u.Message(false, "Error while getting Rack: "+e1)
		u.ErrLog("Error while getting Racks of Room",
			"GET RACKSOFPARENT", e1, r)

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

// swagger:operation GET /api/user/rooms/{id}/racks/{rack_name}/devices rooms GetDevicesOfRoom
// Gets Devices of Room.
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
// - name: rack_name
//   in: path
//   description: Name of desired rack
//   required: true
//   type: string
//   default: "Rack01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetDevicesUsingNamedRackOfRoom = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	name := mux.Vars(r)["rack_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET RACKSOFPARENT", "", r)
		return
	}

	data, e1 := models.GetDevicesUsingNamedRackOfRoom(id, name)
	if data == nil {
		resp = u.Message(false, "Error while getting Devices: "+e1)
		u.ErrLog("Error while getting Devices of Room",
			"GET DEVICESOFROOM", e1, r)

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

// swagger:operation GET /api/user/rooms/{id}/racks/{rack_name}/devices/{device_name} rooms GetDevicesOfRoom
// Get Named Subdevice of Room.
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
// - name: rack_name
//   in: path
//   description: Name of desired rack
//   required: true
//   type: string
//   default: "Rack01"
// - name: device_name
//   in: path
//   description: Name of desired device
//   required: true
//   type: string
//   default: "Device01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetNamedDeviceOfRoom = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	rack_name := mux.Vars(r)["rack_name"]
	device_name := mux.Vars(r)["device_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDDEVICEOFROOM", "", r)
		return
	}

	data, e1 := models.GetNamedDeviceOfRoom(id, rack_name, device_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Device: "+e1)
		u.ErrLog("Error while getting Named Device of Room",
			"GET NAMEDDEVICEOFROOM", e1, r)

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

// swagger:operation GET /api/user/rooms/{id}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name} rooms GetSubdevicesOfRoom
// Get Named Subdevice of Room.
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
// - name: rack_name
//   in: path
//   description: Name of desired rack
//   required: true
//   type: string
//   default: "Rack01"
// - name: device_name
//   in: path
//   description: Name of desired device
//   required: true
//   type: string
//   default: "Device01"
// - name: subdevice_name
//   in: path
//   description: Name of desired subdevice
//   required: true
//   type: string
//   default: "SubdeviceA"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetNamedSubdeviceOfRoom = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	rack_name := mux.Vars(r)["rack_name"]
	device_name := mux.Vars(r)["device_name"]
	subdev_name := mux.Vars(r)["subdevice_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDSUBDEVOFROOM", "", r)
		return
	}

	data, e1 := models.GetNamedSubdeviceOfRoom(id, rack_name, device_name, subdev_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevice: "+e1)
		u.ErrLog("Error while getting Named Subdevice of Room",
			"GET NAMEDSUBDEVOFROOM", e1, r)

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

// swagger:operation GET /api/user/rooms/{id}/racks/{rack_name}/devices/{device_name}/subdevices rooms GetDevicesOfRoom
// Gets Subdevices of a Named Device in Room.
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
// - name: rack_name
//   in: path
//   description: Name of desired rack
//   required: true
//   type: string
//   default: "Rack01"
// - name: device_name
//   in: path
//   description: Name of desired device
//   required: true
//   type: string
//   default: "Device01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetSubdevicesUsingNamedDeviceOfRoom = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	name := mux.Vars(r)["rack_name"]
	dname := mux.Vars(r)["device_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET SUBDEVICESUSINGNAMEDDEVOFROOM", "", r)
		return
	}

	data, e1 := models.GetSubdevicesUsingNamedDeviceOfRoom(id, name, dname)
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevices: "+e1)
		u.ErrLog("Error while getting Subdevices of Room",
			"GET SUBDEVICESUSINGNAMEDDEVOFROOM", e1, r)

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

// swagger:operation GET /api/user/rooms/{id}/racks/{rack_name}/devices/{device_name}/subdevices rooms GetDevicesOfRoom
// Gets Subdevices of a Named Device in Room.
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
// - name: rack_name
//   in: path
//   description: Name of desired rack
//   required: true
//   type: string
//   default: "Rack01"
// - name: device_name
//   in: path
//   description: Name of desired device
//   required: true
//   type: string
//   default: "Device01"
// - name: subdevice_name
//   in: path
//   description: Name of desired subdevice
//   required: true
//   type: string
//   default: "SubdeviceA"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetSubdevice1sUsingUsingNamedSubdeviceOfRoom = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	name := mux.Vars(r)["rack_name"]
	dname := mux.Vars(r)["device_name"]
	sname := mux.Vars(r)["subdevice_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET SUBDEVICESUSINGNAMEDDEVOFROOM", "", r)
		return
	}

	data, e1 := models.GetSubdevice1sUsingUsingNamedSubdeviceOfRoom(id, name, dname, sname)
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevices: "+e1)
		u.ErrLog("Error while getting Subdevices of Room",
			"GET SUBDEVICESUSINGNAMEDDEVOFROOM", e1, r)

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

// swagger:operation GET /api/user/rooms/{id}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s/{subdev1_name} rooms GetDevicesOfRoom
// Get Named Subdevice1 of Room.
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
// - name: rack_name
//   in: path
//   description: Name of desired rack
//   required: true
//   type: string
//   default: "Rack01"
// - name: device_name
//   in: path
//   description: Name of desired device
//   required: true
//   type: string
//   default: "Device01"
// - name: subdevice_name
//   in: path
//   description: Name of desired subdevice
//   required: true
//   type: string
//   default: "Subdevice01"
// - name: subdev1_name
//   in: path
//   description: Name of desired subdevice1
//   required: true
//   type: string
//   default: "SubDev1"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetNamedSubdevice1OfRoom = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	rack_name := mux.Vars(r)["rack_name"]
	device_name := mux.Vars(r)["device_name"]
	subdev_name := mux.Vars(r)["subdevice_name"]
	sd1_name := mux.Vars(r)["subdevice1_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDSUBDEV1OFROOM", "", r)
		return
	}

	data, e1 := models.GetNamedSubdevice1OfRoom(id, rack_name,
		device_name, subdev_name, sd1_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevice1: "+e1)
		u.ErrLog("Error while getting Named Subdevice1 of Room",
			"GET NAMEDSUBDEV1OFROOM", e1, r)

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
