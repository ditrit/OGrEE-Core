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
// - name: ParentID
//   description: 'ParentID of Building refers to Site'
//   required: true
//   type: string
//   default: "999"
// - name: Category
//   in: query
//   description: Category of Building (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "Research"
// - name: Description
//   in: query
//   description: Description of Building
//   required: false
//   type: string[]
//   default: ["Some abandoned building in Grenoble"]
// - name: Domain
//   description: 'Domain Of Building'
//   required: true
//   type: string
//   default: "Some Domain"
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
//   description: 'Size of Building in an XY coordinate format'
//   required: true
//   type: string
//   default: "{\"x\":25.0,\"y\":29.399999618530275}"
// - name: SizeU
//   in: query
//   description: 'The unit for Building Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Height
//   in: query
//   description: 'Height of Building'
//   required: true
//   type: string
//   default: "5"
// - name: HeightU
//   in: query
//   description: 'The unit for Building Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: true
//   type: string
//   default: "m"
// - name: Floors
//   in: query
//   description: 'Number of floors'
//   required: false
//   type: string
//   default: "3"

// responses:
//     '201':
//         description: Created
//     '400':
//         description: Bad request

var CreateBuilding = func(w http.ResponseWriter, r *http.Request) {

	bldg := &models.Building{}
	err := json.NewDecoder(r.Body).Decode(bldg)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE BUILDING", "", r)
		return
	}

	resp, e := bldg.Create()
	switch e {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while creating building", "CREATE BUILDING", e, r)
	case "internal":
		u.ErrLog("Error while creating building", "CREATE BUILDING", e, r)
	default:
		w.WriteHeader(http.StatusCreated)
	}
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
//     '404':
//         description: Not Found

//Retrieve bldg using Bldg ID
var GetBuilding = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET BUILDING", "", r)
		return
	}

	data, e1 := models.GetBuilding(uint(id))
	if data == nil {
		resp = u.Message(false, "Error while getting Building: "+e1)
		u.ErrLog("Error while getting building", "GET BUILDING", e1, r)

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

// swagger:operation GET /api/user/buildings buildings GetAllBuildings
// Gets All Buildings in the system.
// ---
// produces:
// - application/json
// parameters:
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetAllBuildings = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data, e := models.GetAllBuildings()
	if len(data) == 0 {
		resp = u.Message(false, "Error while getting Building: "+e)
		u.ErrLog("Error while getting building", "GET ALL BUILDINGS", e, r)

		switch e {
		case "":
			resp = u.Message(false,
				"Error while getting Building: No Records Found")
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}

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
//     '404':
//        description: Not found
var DeleteBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "DELETE BUILDING", "", r)
		return
	}

	v := models.DeleteBuilding(uint(id))

	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while deleting building", "DELETE BUILDING", "Not Found", r)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}

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
//   default: "New Building"
// - name: Description
//   in: query
//   description: Description of Building
//   required: false
//   type: string[]
//   default: ["Derelict", "Building"]
// - name: Domain
//   description: 'Domain Of Building'
//   required: false
//   type: string
//   default: "Derelict Domain"
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
//   description: 'Size of Building in an XY coordinate format'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeU
//   in: query
//   description: 'The unit for Building Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Height
//   in: query
//   description: 'Height of Building'
//   required: false
//   type: string
//   default: "999"
// - name: HeightU
//   in: query
//   description: 'The unit for Building Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Floors
//   in: query
//   description: 'Number of floors'
//   required: false
//   type: string
//   default: "999"

// responses:
//     '200':
//         description: Updated
//     '404':
//         description: Not Found
//     '400':
//         description: Bad request
//Updates work by passing ID in path parameter
var UpdateBuilding = func(w http.ResponseWriter, r *http.Request) {

	bldg := &models.Building{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "UPDATE BUILDING", "", r)
		return
	}

	err := json.NewDecoder(r.Body).Decode(bldg)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE BUILDING", "", r)
	}

	v, e1 := models.UpdateBuilding(uint(id), bldg)

	switch e1 {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating building", "UPDATE BUILDING", e1, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating building", "UPDATE BUILDING", e1, r)
	case "record not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating building", "UPDATE BUILDING", e1, r)
	default:

	}

	u.Respond(w, v)
}

// swagger:operation GET /api/user/buildings? buildings QueryBuilding
// Gets a Building using any attribute (with the exception of description) via query
// The attributes are in the form {attr}=xyz&{attr1}=abc
// And any combination can be provided given that at least 1 is provided.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired building
//   required: false
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
//   default: "New Building"
// - name: Description
//   in: query
// - name: Domain
//   description: 'Domain Of Building'
//   required: false
//   type: string
//   default: "Derelict Domain"
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
//   description: 'Size of Building in an XY coordinate format'
//   required: false
//   type: string
//   default: "{\"x\":999,\"y\":999}"
// - name: SizeU
//   in: query
//   description: 'The unit for Building Size. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Height
//   in: query
//   description: 'Height of Building'
//   required: false
//   type: string
//   default: "999"
// - name: HeightU
//   in: query
//   description: 'The unit for Building Height. Only values of
//   "mm", "cm", "m", "U", "OU", "tile" are acceptable'
//   required: false
//   type: string
//   default: "cm"
// - name: Floors
//   in: query
//   description: 'Number of floors'
//   required: false
//   type: string
//   default: "999"

// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetBuildingByQuery = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}

	query := u.ParamsParse(r.URL)

	mydata := &models.Building{}
	json.Unmarshal(query, mydata)
	json.Unmarshal(query, &(mydata.Attributes))
	fmt.Println("This is the result: ", *mydata)
	if reflect.DeepEqual(&models.Building{}, mydata) {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from "+
			"path parameters. Please check your query parameters."))
		u.ErrLog("Error while extracting from path parameters",
			"GET BLDG BY QUERY", "", r)
		return
	}

	data, e := models.GetBuildingByQuery(mydata)

	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting building", "GET BLDGQUERY", e, r)

		switch e {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			resp = u.Message(false, "Error: No Records Found")
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}

	} else if len(data) == 1 {
		resp = u.Message(true, "success")
		resp["data"] = map[string]interface{}{"objects": data[0]}

	} else {
		resp = u.Message(true, "success")
		resp["data"] = map[string]interface{}{"objects": data}
	}
	u.Respond(w, resp)
}

/*
var GetBuildingByName = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}
	names := strings.Split(r.URL.String(), "=")

	if names[1] == "" {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "GET BUILDING BY NAME",
			"", r)
		return
	}

	data, e := models.GetBuildingByName(names[1])

	if e != "" {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting building", "GET Building", e, r)

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

// swagger:operation GET /api/user/buildings/{id}/all buildings GetBuildingHierarchy
// Gets Building hierarchy using Building ID.
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
//     '404':
//         description: Not Found

var GetBuildingHierarchy = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("me & the irishman")
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET Building", "", r)
		return
	}

	data, e1 := models.GetBuildingHierarchy(uint(id))

	if data == nil {
		resp = u.Message(false, "Error while getting Building: "+e1)
		u.ErrLog("Error while getting Building", "GET Building", e1, r)

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

var GetBuildingHierarchyNonStandard = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("me & the irishman")
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET Building", "", r)
		return
	}

	data, rooms, racks, devices, e1 :=
		models.GetBuildingHierarchyNonStandard(uint(id))

	if data == nil {
		resp = u.Message(false, "Error while getting Building: "+e1)
		u.ErrLog("Error while getting Building", "GET Building", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	resp["rooms"] = rooms
	resp["racks"] = racks
	resp["devices"] = devices
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/buildings/{id}/rooms buildings GetRoomsOfBuilding
// Gets Rooms of Building.
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
//     '404':
//         description: Not Found

var GetRoomsOfBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ROOMSOFBUILDING", "", r)
		return
	}

	data, e1 := models.GetRoomsOfBuilding(id)
	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Rooms: "+e1)
		u.ErrLog("Error while getting Rooms of Building",
			"GET RACKSOFPARENT", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			w.WriteHeader(http.StatusNotFound)
			resp = u.Message(false, "Error: No Records Found")
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}

// swagger:operation GET /api/user/buildings/{id}/rooms/{room_name} buildings GetRoomsOfBuilding
// Gets a Room of Building.
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
// - name: room_name
//   in: path
//   description: name of room
//   required: true
//   type: string
//   default: "R1"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetNamedRoomOfBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	name := mux.Vars(r)["room_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDROOMOFBUILDING", "", r)
		return
	}

	data, e1 := models.GetNamedRoomOfBuilding(id, name)

	/*data, e1 := models.GetRack(uint(id))*/
	if data == nil {
		resp = u.Message(false, "Error while getting Room: "+e1)
		u.ErrLog("Error while getting NamedRoom Of Building",
			"GET NAMEDROOMOFBUILDING", e1, r)

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

// swagger:operation GET /api/user/buildings/{id}/rooms/{room_name}/racks buildings GetRacksOfBuilding
// Gets Racks of named Room of Building.
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
// - name: room_name
//   in: path
//   description: name of room
//   required: true
//   type: string
//   default: "R1"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetRacksUsingNamedRoomOfBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	name := mux.Vars(r)["room_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET RACKSOFBLDG", "", r)
		return
	}

	data, e1 := models.GetRacksUsingNamedRoomOfBuilding(id, name)
	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Racks: "+e1)
		u.ErrLog("Error while getting Racks of Building",
			"GET RACKSUSINGNAMEDROOMOFBLDG", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			resp["message"] = "Error: No Records Found"
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}

// swagger:operation GET /api/user/buildings/{id}/rooms/{room_name}/racks/{rack_name} buildings GetRacksOfBuilding
// Gets a Rack of Building.
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
// - name: room_name
//   in: path
//   description: name of room
//   required: true
//   type: string
//   default: "R1"
// - name: rack_name
//   in: path
//   description: name of rack
//   required: true
//   type: string
//   default: "Rack01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetNamedRackOfBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDRACKOFBLDG", "", r)
		return
	}

	data, e1 := models.GetNamedRackOfBuilding(id, room_name, rack_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Rack: "+e1)
		u.ErrLog("Error while getting Named Rack of Building",
			"GET NAMEDRACKOFBLDG", e1, r)

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

// swagger:operation GET /api/user/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices buildings GetDevicesOfBuilding
// Gets a Devices of Building.
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
// - name: room_name
//   in: path
//   description: name of room
//   required: true
//   type: string
//   default: "R1"
// - name: rack_name
//   in: path
//   description: name of rack
//   required: true
//   type: string
//   default: "Rack01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetDevicesUsingNamedRackOfBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET DEVICESOFBLDG", "", r)
		return
	}

	data, e1 := models.GetDevicesUsingNamedRackOfBuilding(id, room_name, rack_name)
	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Devices: "+e1)
		u.ErrLog("Error while getting Devices of Building",
			"GET DEVICESOFBLDG", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}

// swagger:operation GET /api/user/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name} buildings GetDevicesOfBuilding
// Gets a Devices of Building.
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
// - name: room_name
//   in: path
//   description: name of room
//   required: true
//   type: string
//   default: "R1"
// - name: rack_name
//   in: path
//   description: name of rack
//   required: true
//   type: string
//   default: "Rack01"
// - name: device_name
//   in: path
//   description: name of device
//   required: true
//   type: int
//   default: "Device01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetNamedDeviceOfBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	device_name := mux.Vars(r)["device_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDDEVICEOFBLDG", "", r)
		return
	}

	data, e1 := models.GetNamedDeviceOfBuilding(id, room_name, rack_name, device_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Device: "+e1)
		u.ErrLog("Error while getting Named Device of Building",
			"GET NAMEDDEVICEOFBLDG", e1, r)

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

// swagger:operation GET /api/user/buildings/{id}/all/rooms/racks buildings GetBuildingHierarchy
// Gets hierarchy of Building until racks.
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
//     '404':
//         description: Not Found
var GetBuildingHierarchyToRack = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET BLDGHIERARCHTORACK", "", r)
		return
	}

	data, e1 := models.GetBuildingHierarchyToRack(id)
	if data == nil {
		resp = u.Message(false, "Error while getting Building Hierarchy: "+e1)
		u.ErrLog("Error while getting Building Hierarchy",
			"GET BLDGHIERARCHTORACK", e1, r)

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

// swagger:operation GET /api/user/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices buildings GetDevicesOfBuilding
// Gets Subdevices using named Device of Building.
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
// - name: room_name
//   in: path
//   description: name of room
//   required: true
//   type: string
//   default: "R1"
// - name: rack_name
//   in: path
//   description: name of rack
//   required: true
//   type: string
//   default: "Rack01"
// - name: device_name
//   in: path
//   description: name of device
//   required: true
//   type: int
//   default: "Device01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetSubdevicesUsingNamedDeviceOfBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	device_name := mux.Vars(r)["device_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters",
			"GET SUBDEVUSINGNAMEDDEVICEOFBLDG", "", r)
		return
	}

	data, e1 := models.GetSubdevicesUsingNamedDeviceOfBuilding(id,
		room_name, rack_name, device_name)

	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Subdevices: "+e1)
		u.ErrLog("Error while getting Subdevices of Building",
			"GET SUBDEVUSINGNAMEDDEVICEOFBLDG", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}

	u.Respond(w, resp)
}

// swagger:operation GET /api/user/buildings/{id}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name} buildings GetSubdevicesOfBuilding
// Gets Subdevices of Building.
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
// - name: room_name
//   in: path
//   description: name of room
//   required: true
//   type: string
//   default: "R1"
// - name: rack_name
//   in: path
//   description: name of rack
//   required: true
//   type: string
//   default: "Rack01"
// - name: device_name
//   in: path
//   description: name of device
//   required: true
//   type: int
//   default: "Device01"
// - name: subdevice_name
//   in: path
//   description: name of subdevice
//   required: true
//   type: int
//   default: "Subdevice01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetNamedSubdeviceOfBuilding = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	device_name := mux.Vars(r)["device_name"]
	subdev_name := mux.Vars(r)["subdevice_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDSUBDEVOFBLDG", "", r)
		return
	}

	data, e1 := models.GetNamedSubdeviceOfBuilding(id, room_name,
		rack_name, device_name, subdev_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevice: "+e1)
		u.ErrLog("Error while getting Named Subdevice of Building",
			"GET NAMEDSUBDEVOFBLDG", e1, r)

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
