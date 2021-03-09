package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strconv"

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
//   type: int
//   default: 999
// - name: ParentID
//   description: 'Parent of Site refers to Tenant ID'
//   required: true
//   type: int
//   default: 999
// - name: Description
//   in: query
//   description: Description of Site
//   required: true
//   type: string
//   default: "Some abandoned site in Grenoble"
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
//   required: false
//   type: string
//   default: "Silver"
// - name: ReservedColor
//   in: query
//   description: Reserved Color of Site (useful for 3D rendering)
//   required: false
//   type: string
//   default: "Silver"
// - name: TechnicalColor
//   in: query
//   description: Color of Site (useful for 3D rendering)
//   required: false
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
//     '200':
//         description: Created
//     '400':
//         description: Bad request

var CreateSite = func(w http.ResponseWriter, r *http.Request) {

	site := &models.Site{}
	err := json.NewDecoder(r.Body).Decode(site)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		return
	}

	resp := site.Create()
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
// Get all Sites of a Site using Site ID.
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
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}
	resp := u.Message(true, "success")

	id, _ := strconv.Atoi(st.Domain)

	data := models.GetSites(uint(id))
	if data == nil {
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
//     '400':
//         description: Not Found

//Retrieve site using Site ID
var GetSite = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	data := models.GetSite(uint(id))
	if data == nil {
		resp = u.Message(false, "unsuccessful")
	}

	resp["data"] = data
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/sites/{id} sites GetTenant
// Gets a Site(s) from the system.
// The ID must be provided in the URL parameter
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
//     '400':
//         description: Bad request
var GetAllSites = func(w http.ResponseWriter, r *http.Request) {

	resp := u.Message(true, "success")

	data := models.GetAllSites()
	if data == nil {
		resp = u.Message(false, "unsuccessful")
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

var DeleteSiteByID = func(w http.ResponseWriter, r *http.Request) {
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	v := models.DeleteSite(uint(id))
	u.Respond(w, v)
}

// swagger:operation DELETE /api/user/sites/{id} sites DeleteSite
// Deletes a Site in the system.
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
//     '204':
//        description: Successful
//     '400':
//        description: Not found

//Delete all sites of a site
var DeleteSites = func(w http.ResponseWriter, r *http.Request) {
	st := &models.Site{}
	err := json.NewDecoder(r.Body).Decode(st)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	id, _ := strconv.Atoi(st.Domain)
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
//   type: int
//   default: 999
// - name: Description
//   in: query
//   description: Description of Site
//   required: false
//   type: string
//   default: "Some abandoned site in Grenoble"
// - name: Color
//   in: query
//   description: Color of Site (useful for 3D rendering)
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
//     '400':
//         description: Bad request
//Updates work by passing ID in path parameter

var UpdateSite = func(w http.ResponseWriter, r *http.Request) {

	site := &models.Site{}
	id, e := strconv.Atoi(mux.Vars(r)["id"])

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
	}

	err := json.NewDecoder(r.Body).Decode(site)
	if err != nil {
		u.Respond(w, u.Message(false, "Error while decoding request body"))
	}

	v := models.UpdateSite(uint(id), site)
	u.Respond(w, v)
}
