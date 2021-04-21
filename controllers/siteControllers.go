package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// swagger:operation POST /api/user/sites sites CreateSite
// Creates a Site in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: Name of site
//   required: true
//   type: string
//   default: "Site A"
// - name: Category
//   in: query
//   description: Category of Site (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "Research"
// - name: Domain
//   description: 'Domain of Site'
//   required: true
//   type: string
//   default: 999
// - name: ParentID
//   description: 'Parent of Site refers to Tenant ID'
//   required: true
//   type: int
//   default: 999
// - name: Description
//   in: query
//   description: Description of Site
//   required: false
//   type: string[]
//   default: ["Some abandoned site in Grenoble"]
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "NE", "NW", "SE", "SW" are acceptable'
//   required: true
//   type: string
//   default: "NE"
// - name: UsableColor
//   in: query
//   description: Usable Color of Site (useful for 3D rendering)
//   required: true
//   type: string
//   default: "Silver"
// - name: ReservedColor
//   in: query
//   description: Reserved Color of Site (useful for 3D rendering)
//   required: true
//   type: string
//   default: "Silver"
// - name: TechnicalColor
//   in: query
//   description: Color of Site (useful for 3D rendering)
//   required: true
//   type: string
//   default: "Silver"
// - name: Address
//   in: query
//   description: Address of Site
//   required: false
//   type: string
//   default: "Rue pour Nissan"
// - name: Zipcode
//   in: query
//   description: Zipcode of Site
//   required: false
//   type: string
//   default: "10000"
// - name: City
//   in: query
//   description: City of Site
//   required: false
//   type: string
//   default: "Paris"
// - name: Country
//   in: query
//   description: Country of Site
//   required: false
//   type: string
//   default: "France"
// - name: Gps
//   in: query
//   description: Gps of Site
//   required: false
//   type: string
//   default: "N'25 E'55"
// responses:
//     '201':
//         description: Created
//     '400':
//         description: Bad request

var CreateSite = func(w http.ResponseWriter, r *http.Request) {

	site := &models.Site{}
	err := json.NewDecoder(r.Body).Decode(site)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE SITE", "", r)
		return
	}

	resp, e := site.Create()

	switch e {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog(e+" Error", "CREATE SITE", "", r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog(e+" Error", "CREATE SITE", "", r)
	default:
		w.WriteHeader(http.StatusCreated)
	}

	u.Respond(w, resp)
}

// swagger:operation GET /api/user/sites sites GetSitesByUserID
// Gets a Site(s) from the system using User ID.
// The ID is automatically obtained from the Authorization header
// This is still in progress
// It is based on the idea of 1 site 1 user
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of user
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: Found
//     '400':
//         description: Not Found

//Retrieve sites using User ID
var GetSitesByUserID = func(w http.ResponseWriter, r *http.Request) {

	id := r.Context().Value("user").(uint)
	resp := u.Message(true, "success")

	data := models.GetSites(uint(id))
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/sites sites GetSitesByParentID
// Get all Sites of a Tenant using Site ID.
// The ID is provided in JSON and not in
// parameter. This is a new feature in progress
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of user
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: Found
//     '400':
//         description: Not Found

//Retrieve sites using Site ID
var GetSitesByParentID = func(w http.ResponseWriter, r *http.Request) {

	st := &models.Site{}
	err := json.NewDecoder(r.Body).Decode(st)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}
	resp := u.Message(true, "success")

	id, _ := strconv.Atoi(st.Domain)

	data := models.GetSites(uint(id))
	if data == nil {
		w.WriteHeader(http.StatusNoContent)
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/sites/{id} sites GetSite
// Gets a Site from the system using Site ID.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired site
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

//Retrieve site using Site ID
var GetSite = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET SITE", "", r)
	}

	data, e1 := models.GetSite(uint(id))

	if data == nil {
		resp = u.Message(false, "Error while getting Site: "+e1)
		u.ErrLog("Error while getting Site", "GET SITE", e1, r)

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

// swagger:operation GET /api/user/sites sites GetAllSites
// Gets all Sites from the system.
// ---
// produces:
// - application/json
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Bad request
var GetAllSites = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data, e := models.GetAllSites()

	if len(data) == 0 {
		resp = u.Message(false, "Error while getting all sites: "+e)
		u.ErrLog("Error while getting all sites", "GET ALL SITES", e, r)
		switch e {
		case "":
			w.WriteHeader(http.StatusNotFound)
			//
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

/*var DeleteSite = func(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("user").(uint)
	v := models.DeleteSite(id)
	u.Respond(w, v)
}
*/

// swagger:operation DELETE /api/user/sites/{id} sites DeleteSite
// Deletes a Site.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: query
//   description: ID of Site
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '404':
//        description: Not Found

var DeleteSiteByID = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	v := models.DeleteSite(uint(id))
	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
	u.Respond(w, v)
}

// swagger:operation DELETE /api/user/sites sites DeleteSitesOfTenant
// Deletes all Sites of a Tenant.
// You must provide the Parent ID of Site in the JSON Body
// ---
// produces:
// - application/json
// parameters:
// - name: ParentID
//   in: query
//   description: ParentID of Site refers to Tenant
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '400':
//        description: Not found

//Delete all sites of a tenant
var DeleteSites = func(w http.ResponseWriter, r *http.Request) {
	st := &models.Site{}
	err := json.NewDecoder(r.Body).Decode(st)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "DELETE SITE", "", r)
	}

	id, _ := strconv.Atoi(st.ParentID)
	v := models.DeleteSitesOfTenant(uint(id))
	u.Respond(w, v)
}

// swagger:operation PUT /api/user/sites/{id} sites UpdateSite
// Changes Site data in the system.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired site
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of site
//   required: false
//   type: string
//   default: "Site B"
// - name: Category
//   in: query
//   description: Category of Site (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Research"
// - name: Domain
//   description: 'Domain of Site'
//   required: false
//   type: string
//   default: "Some Domain"
// - name: Description
//   in: query
//   description: Description of Site
//   required: false
//   type: string
//   default: "Some abandoned site in Grenoble"
// - name: Orientation
//   in: query
//   description: 'Indicates the location. Only values of
//   "NE", "NW", "SE", "SW" are acceptable'
//   required: false
//   type: string
//   default: "NE"
// - name: UsableColor
//   in: query
//   description: Usable Color of Site (useful for 3D rendering)
//   required: false
//   type: string
//   default: "Black"
// - name: ReservedColor
//   in: query
//   description: Reserved Color of Site (useful for 3D rendering)
//   required: false
//   type: string
//   default: "Black"
// - name: TechnicalColor
//   in: query
//   description: Color of Site (useful for 3D rendering)
//   required: false
//   type: string
//   default: "Black"
// - name: Address
//   in: query
//   description: Address of Site
//   required: false
//   type: string
//   default: "New Rue"
// - name: Zipcode
//   in: query
//   description: Zipcode of Site
//   required: false
//   type: string
//   default: "99999"
// - name: City
//   in: query
//   description: City of Site
//   required: false
//   type: string
//   default: "Geneve"
// - name: Country
//   in: query
//   description: Country of Site
//   required: false
//   type: string
//   default: "Switzerland"
// - name: Gps
//   in: query
//   description: Gps of Site
//   required: false
//   type: string
//   default: "N'55 E'15"

// responses:
//     '200':
//         description: Updated
//     '404':
//         description: Not Found
//     '400':
//         description: Bad request

var UpdateSite = func(w http.ResponseWriter, r *http.Request) {

	site := &models.Site{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "UPDATE SITE", "", r)
	}

	err := json.NewDecoder(r.Body).Decode(site)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE SITE", "", r)
	}

	v, e1 := models.UpdateSite(uint(id), site)

	switch e1 {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating site", "UPDATE SITE", e1, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating site", "UPDATE SITE", e1, r)
	case "record not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating site", "UPDATE SITE", e1, r)
	default:
	}

	u.Respond(w, v)
}

var GetSiteByName = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}
	names := strings.Split(r.URL.String(), "=")

	if names[1] == "" {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "GET SITE BY NAME",
			"", r)
		return
	}

	data, e := models.GetSiteByName(names[1])

	if e != "" {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting site", "GET Site", e, r)

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

var GetSiteHierarchy = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("me & the irishman")
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET Site", "", r)
	}

	data, e1 := models.GetSiteHierarchy(id)

	if data == nil {
		resp = u.Message(false, "Error while getting Site: "+e1)
		u.ErrLog("Error while getting Site", "GET Site", e1, r)

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

var GetBuildingsOfSite = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET BUILDINGSOFSITE", "", r)
	}

	data, e1 := models.GetBuildingsOfSite(id)
	if data == nil {
		resp = u.Message(false, "Error while getting Buildings: "+e1)
		u.ErrLog("Error while getting Buildings Of Site",
			"GET BUILDINGSOFSITE", e1, r)

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

var GetNamedBuildingOfSite = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	name := mux.Vars(r)["building_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDBUILDINGSOFSITE", "", r)
	}

	data, e1 := models.GetNamedBuildingOfSite(id, name)
	if data == nil {
		resp = u.Message(false, "Error while getting Building: "+e1)
		u.ErrLog("Error while getting Named Building Of Site",
			"GET NAMEDBUILDINGSOFSITE", e1, r)

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

var GetRoomsUsingNamedBldgOfSite = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	bldg_name := mux.Vars(r)["building_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ROOMSUSINGNAMEDBLDGOFSITE", "", r)
	}

	data, e1 := models.GetRoomsUsingNamedBldgOfSite(id, bldg_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Rooms: "+e1)
		u.ErrLog("Error while getting Rooms of Site",
			"GET ROOMSUSINGNAMEDBLDGOFSITE", e1, r)

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

var GetNamedRoomOfSite = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	bldg_name := mux.Vars(r)["building_name"]
	room_name := mux.Vars(r)["room_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDROOMOFSITE", "", r)
	}

	data, e1 := models.GetNamedRoomOfSite(id, bldg_name, room_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Room: "+e1)
		u.ErrLog("Error while getting Named Room Of Site",
			"GET NAMEDROOMOFSITE", e1, r)

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

var GetRacksUsingNamedRoomOfSite = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	bldg_name := mux.Vars(r)["building_name"]
	room_name := mux.Vars(r)["room_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET RACKSUSINGNAMEDROOMOFSITE", "", r)
	}

	data, e1 := models.GetRacksUsingNamedRoomOfSite(id, bldg_name, room_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Racks: "+e1)
		u.ErrLog("Error while getting Racks using Named Room Of Site",
			"GET RACKSUSINGNAMEDROOMOFSITE", e1, r)

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

var GetNamedRackOfSite = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	bldg_name := mux.Vars(r)["building_name"]
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	resp := u.Message(true, "success")
	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDRACKOFSITE", "", r)
	}

	data, e1 := models.GetNamedRackOfSite(id, bldg_name, room_name, rack_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Rack: "+e1)
		u.ErrLog("Error while getting Named Rack Of Site",
			"GET NAMEDRACKOFSITE", e1, r)

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
