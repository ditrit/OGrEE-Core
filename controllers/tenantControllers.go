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

// swagger:operation POST /api/user/tenants tenants CreateTenant
// Creates a Tenant in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: Name of tenant
//   required: true
//   type: string
//   default: "Nissan"
// - name: Category
//   in: query
//   description: Category of Tenant (ex. Consumer Electronics, Medical)
//   required: true
//   type: string
//   default: "Auto"
// - name: Description
//   in: query
//   description: 'Description of the Tenant'
//   required: false
//   type: string[]
//   example: ["Worldwide automotive company"]
// - name: Domain
//   description: 'Domain of the Tenant'
//   required: true
//   type: string
//   default: "Auto"
// - name: Color
//   in: query
//   description: Color of Tenant (useful for 3D rendering)
//   required: true
//   type: string
//   default: "Silver"
// - name: MainContact
//   in: query
//   description: The main method of contact for Tenant
//   required: false
//   type: string
//   default: "Website"
// - name: MainPhone
//   in: query
//   description: Main Phone # of Tenant
//   required: false
//   type: string
//   default: "000"
// - name: MainEmail
//   in: query
//   description: Main Email Address of Tenant
//   required: false
//   type: string
//   default: "nissan@nissan.com"
// responses:
//     '201':
//         description: Created
//     '400':
//         description: Bad request

var CreateTenant = func(w http.ResponseWriter, r *http.Request) {
	tenant := &models.Tenant{}
	err := json.NewDecoder(r.Body).Decode(tenant)

	//Copy Request if you want to reuse the JSON
	//For Error logging
	//bt, _ := httputil.DumpRequest(r, true)
	//println(string(bt))
	//q := io.TeeReader(r.Body, bytes.Buffer)

	//q := r.Body
	//s, _ := ioutil.ReadAll(q)
	//println(string(s))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "CREATE TENANT", "", r)

		return
	}

	resp, e := tenant.Create()

	switch e {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while creating tenant", "CREATE TENANT", e, r)
	case "":
		w.WriteHeader(http.StatusCreated)
		u.ErrLog("Error while creating tenant", "CREATE TENANT", e, r)
	default:
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while creating tenant", "CREATE TENANT", e, r)
	}

	u.Respond(w, resp)
}

// swagger:operation GET /api/user/tenants/{id} tenants GetTenant
// Gets a Tenant(s) from the system.
// The ID must be provided in the URL parameter
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired tenant
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: Found
//     '400':
//         description: Bad request
//     '404':
//         description: Not Found
var GetTenantFor = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}
	id, err := strconv.Atoi(mux.Vars(r)["id"])

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "GET TENANT", "", r)
		return
	}

	data, e := models.GetTenant(uint(id))

	if e != "" {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting tenant", "GET TENANT", e, r)

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

// swagger:operation GET /api/user/tenants tenants GetAllTenants
// Obtain all tenants in the system.
// Returns JSON body with all tenants and their IDs
// ---
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Nothing Found
var GetAllTenants = func(w http.ResponseWriter, r *http.Request) {

	data, e := models.GetAllTenants()
	resp := u.Message(true, "success")

	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting tenants", "GET ALL TENANTS", e, r)

		switch e {
		case "validate":

		case "":
			resp = u.Message(false, "Error: No Records Found")
			w.WriteHeader(http.StatusNotFound)
			//
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}
	u.Respond(w, resp)
}

// swagger:operation PUT /api/user/tenants/{id} tenants UpdateTenant
// Changes Tenant data in the system.
// If no new or any information is provided
// an OK will still be returned
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired tenant
//   required: true
//   type: int
//   default: 999
// - name: Name
//   in: query
//   description: Name of tenant
//   required: false
//   type: string
//   default: "INFINITI"
// - name: Category
//   in: query
//   description: Category of Tenant (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Auto"
// - name: Description
//   in: query
//   description: Description of Tenant
//   required: false
//   type: string[]
//   default: "High End Worldwide automotive company"
// - name: Domain
//   description: 'Domain of the Tenant'
//   required: false
//   type: string
//   default: "High End Auto"
// - name: Color
//   in: query
//   description: Color of Tenant (useful for 3D rendering)
//   required: false
//   type: string
//   default: "Black"
// - name: MainContact
//   in: query
//   description: The main method of contact for Tenant
//   required: false
//   type: string
//   default: "Post"
// - name: MainPhone
//   in: query
//   description: Main Phone # of Tenant
//   required: false
//   type: string
//   default: "999"
// - name: MainEmail
//   in: query
//   description: Main Email Address of Tenant
//   required: false
//   type: string
//   default: "infiniti@nissan.com"
// responses:
//     '200':
//         description: Updated
//     '400':
//         description: Bad request
//     '404':
//         description: Not Found

var UpdateTenant = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "UPDATE TENANT", "", r)
		return
	}
	tenant := &models.Tenant{}

	err := json.NewDecoder(r.Body).Decode(tenant)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while decoding request body"))
		u.ErrLog("Error while decoding request body", "UPDATE TENANT", "", r)
	}

	v, e1 := models.UpdateTenant(uint(id), tenant)

	switch e1 {

	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while updating tenant", "UPDATE TENANT", e1, r)
	case "internal":
		w.WriteHeader(http.StatusInternalServerError)
		u.ErrLog("Error while updating tenant", "UPDATE TENANT", e1, r)
	case "record not found":
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Error while updating tenant", "UPDATE TENANT", e1, r)
	default:
	}

	u.Respond(w, v)
}

//This delete function is for 1 tenant 1 user
/*var DeleteTenant = func(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("user").(uint)
	v := models.DeleteTenant(id)
	u.Respond(w, v)
}*/

// swagger:operation DELETE /api/user/tenants/{id} tenants DeleteTenant
// Deletes a Tenant in the system.
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: path
//   description: ID of desired tenant
//   required: true
//   type: int
//   default: 999
// responses:
//     '204':
//        description: Successful
//     '404':
//        description: Not found

var DeleteTenant = func(w http.ResponseWriter, r *http.Request) {

	id, e := strconv.Atoi(mux.Vars(r)["id"])
	if e != nil {
		u.Respond(w, u.Message(false, "Error while extracting from path parameters"))
		u.ErrLog("Error while extracting from path parameters", "DELETE TENANT", "", r)
		return
	}

	v := models.DeleteTenant(uint(id))
	if v["status"] == false {
		w.WriteHeader(http.StatusNotFound)
		u.ErrLog("Not Found", "DELETE TENANT", "", r)

	} else {
		w.WriteHeader(http.StatusNoContent)
	}
	u.Respond(w, v)
}

// swagger:operation GET /api/user/tenants? tenants GetTenant
// Gets a Tenant using any attribute (with the exception of description) via query in the system
// The attributes are in the form {attr}=xyz&{attr1}=abc
// And any combination can be provided given that at least 1 is provided.
// ---
// produces:
// - application/json
// parameters:
// - name: Name
//   in: query
//   description: Name of tenant
//   required: false
//   type: string
//   default: "INFINITI"
// - name: Category
//   in: query
//   description: Category of Tenant (ex. Consumer Electronics, Medical)
//   required: false
//   type: string
//   default: "Auto"
// - name: Domain
//   description: 'Domain of the Tenant'
//   required: false
//   type: string
//   default: "High End Auto"
// - name: Color
//   in: query
//   description: Color of Tenant (useful for 3D rendering)
//   required: false
//   type: string
//   default: "Black"
// - name: MainContact
//   in: query
//   description: The main method of contact for Tenant
//   required: false
//   type: string
//   default: "Post"
// - name: MainPhone
//   in: query
//   description: Main Phone # of Tenant
//   required: false
//   type: string
//   default: "999"
// - name: MainEmail
//   in: query
//   description: Main Email Address of Tenant
//   required: false
//   type: string
//   default: "infiniti@nissan.com"
// responses:
//     '204':
//        description: Found
//     '404':
//        description: Not found

var GetTenantByQuery = func(w http.ResponseWriter, r *http.Request) {
	var resp map[string]interface{}

	query := u.ParamsParse(r.URL)

	mydata := &models.Tenant{}
	json.Unmarshal(query, mydata)
	json.Unmarshal(query, &(mydata.Attributes))
	fmt.Println("This is the result: ", *mydata)
	if reflect.DeepEqual(&models.Tenant{}, mydata) {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message(false, "Error while extracting from "+
			"path parameters. Please check your query parameters."))
		u.ErrLog("Error while extracting from path parameters",
			"GET TENANT BY QUERY", "", r)
		return
	}

	data, e := models.GetTenantByQuery(mydata)

	if len(data) == 0 {
		resp = u.Message(false, "Error: "+e)
		u.ErrLog("Error while getting tenant", "GET TENANTQUERY", e, r)

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

// swagger:operation GET /api/user/tenants/{id}/all tenants GetFromTenant
// Obtain all objects related to tenant in the system.
// Returns JSON body with all objects under the Tenant
// ---
// produces:
// - application/json
// parameters:
// - name: ID
//   in: query
//   description: ID of tenant
//   required: true
//   type: int
//   default: 999
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Nothing Found

var GetTenantHierarchy = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("me & the irishman")
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET Tenant", "", r)
		return
	}

	data, e1 := models.GetTenantHierarchy(id)

	if data == nil {
		resp = u.Message(false, "Error while getting Tenant: "+e1)
		u.ErrLog("Error while getting Tenant", "GET Tenant", e1, r)

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

var GetTenantHierarchyNonStandard = func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("me & the irishman")
	id, e := strconv.Atoi(mux.Vars(r)["id"])
	resp := u.Message(true, "success")

	if e != nil {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET Tenant", "", r)
		return
	}

	data, sites, bldgs, rooms,
		racks, devices, e1 := models.GetTenantHierarchyNonStandard(id)

	if data == nil {
		resp = u.Message(false, "Error while getting Tenant: "+e1)
		u.ErrLog("Error while getting Tenant", "GET Tenant", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = data
	resp["data"] = sites
	resp["buildings"] = bldgs
	resp["rooms"] = rooms
	resp["racks"] = racks
	resp["devices"] = devices
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/tenants/{tenant_name}/sites tenants GetFromTenant
// Gets all sites of a Tenant from the system.
// The Tenant name must be provided in the URL parameter
// ---
// produces:
// - application/json
// parameters:
// - name: name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetSitesOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	resp := u.Message(true, "success")
	if e != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET SITESOFTENANT", "", r)
		return
	}

	data, e1 := models.GetSitesOfTenant(name)
	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Sites: "+e1)
		u.ErrLog("Error while getting Sites of Tenant",
			"GET SITESOFTENANT", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			w.WriteHeader(http.StatusNotFound)
			resp["message"] = "Error: No Records Found"
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name} tenants GetFromTenant
// Gets a site of a Tenant by name from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetNamedSiteOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	site_name, e2 := mux.Vars(r)["site_name"]
	resp := u.Message(true, "success")
	if e != true || e2 != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDSITEOFTENANT", "", r)
		return
	}

	data, e1 := models.GetNamedSiteOfTenant(name, site_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Site: "+e1)
		u.ErrLog("Error while getting Named Site of Tenant",
			"GET NAMEDSITEOFTENANT", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings tenants GetFromTenant
// Gets Buildings of a named site of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetBuildingsUsingNamedSiteOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	site_name, e2 := mux.Vars(r)["site_name"]
	resp := u.Message(true, "success")
	if e != true || e2 != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET BUILDINGSUSINGNAMEDSITEOFTENANT", "", r)
		return
	}

	data, e1 := models.GetBuildingsUsingNamedSiteOfTenant(name, site_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Site: "+e1)
		u.ErrLog("Error while getting Named Site of Tenant",
			"GET BUILDINGSUSINGNAMEDSITEOFTENANT", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			w.WriteHeader(http.StatusNotFound)
			resp["message"] = "Error: No Records Found"
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name} tenants GetFromTenant
// Gets a Building by name of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetNamedBuildingOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	site_name, e2 := mux.Vars(r)["site_name"]
	bldg_name, e3 := mux.Vars(r)["building_name"]
	resp := u.Message(true, "success")
	if e != true || e2 != true || e3 != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDBLDGOFTENANT", "", r)
		return
	}

	data, e1 := models.GetNamedBuildingOfTenant(name, site_name, bldg_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Bldg: "+e1)
		u.ErrLog("Error while getting Named Bldg of Tenant",
			"GET NAMEDBLDGOFTENANT", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms tenants GetFromTenant
// Gets a Building by name of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetRoomsUsingNamedBuildingOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	site_name, e2 := mux.Vars(r)["site_name"]
	bldg_name, e3 := mux.Vars(r)["building_name"]
	resp := u.Message(true, "success")
	if e != true || e2 != true || e3 != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ROOMSUSINGNAMEDBLDGOFTENANT", "", r)
		return
	}

	data, e1 := models.GetRoomsUsingNamedBuildingOfTenant(name, site_name, bldg_name)
	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Rooms: "+e1)
		u.ErrLog("Error while getting Rooms using Named Bldg of Tenant",
			"GET ROOMSUSINGNAMEDBLDGOFTENANT", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			w.WriteHeader(http.StatusNotFound)
			resp["message"] = "Error: No Records Found"
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name} tenants GetFromTenant
// Gets a Room by name of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// - name: room_name
//   in: path
//   description: Name of desired room
//   required: true
//   type: string
//   default: "R1"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetNamedRoomOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	site_name, e2 := mux.Vars(r)["site_name"]
	bldg_name, e3 := mux.Vars(r)["building_name"]
	room_name, e4 := mux.Vars(r)["room_name"]
	resp := u.Message(true, "success")
	if e != true || e2 != true || e3 != true || e4 != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDROOMOFTENANT", "", r)
		return
	}

	data, e1 := models.GetNamedRoomOfTenant(name, site_name, bldg_name, room_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Room: "+e1)
		u.ErrLog("Error while getting Named Room of Tenant",
			"GET NAMEDROOMOFTENANT", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks tenants GetFromTenant
// Gets all Racks of a Room by name of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// - name: room_name
//   in: path
//   description: Name of desired room
//   required: true
//   type: string
//   default: "R1"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found

var GetRacksUsingNamedRoomOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	site_name, e2 := mux.Vars(r)["site_name"]
	bldg_name, e3 := mux.Vars(r)["building_name"]
	room_name, e4 := mux.Vars(r)["room_name"]
	resp := u.Message(true, "success")
	if e != true || e2 != true || e3 != true || e4 != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET RACKSUSINGNAMEDROOMOFTENANT", "", r)
		return
	}

	data, e1 := models.GetRacksUsingNamedRoomOfTenant(name, site_name, bldg_name, room_name)
	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Racks: "+e1)
		u.ErrLog("Error while getting Racks using Named Room of Tenant",
			"GET RACKSUSINGNAMEDROOMOFTENANT", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			w.WriteHeader(http.StatusNotFound)
			resp["message"] = "Error: No Records Found"
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name} tenants GetFromTenant
// Gets a Rack by name of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// - name: room_name
//   in: path
//   description: Name of desired room
//   required: true
//   type: string
//   default: "R1"
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

var GetNamedRackOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	site_name, e2 := mux.Vars(r)["site_name"]
	bldg_name, e3 := mux.Vars(r)["building_name"]
	room_name, e4 := mux.Vars(r)["room_name"]
	rack_name, e5 := mux.Vars(r)["rack_name"]
	resp := u.Message(true, "success")
	if e != true || e2 != true || e3 != true || e4 != true || e5 != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDRACKOFTENANT", "", r)
		return
	}

	data, e1 := models.GetNamedRackOfTenant(name, site_name, bldg_name, room_name, rack_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Rack: "+e1)
		u.ErrLog("Error while getting Named Rack of Tenant",
			"GET NAMEDRACKOFTENANT", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices tenants GetFromTenant
// Gets all Devices of a Rack by name of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// - name: room_name
//   in: path
//   description: Name of desired room
//   required: true
//   type: string
//   default: "R1"
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

var GetDevicesUsingNamedRackOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	site_name, e2 := mux.Vars(r)["site_name"]
	bldg_name, e3 := mux.Vars(r)["building_name"]
	room_name, e4 := mux.Vars(r)["room_name"]
	rack_name, e5 := mux.Vars(r)["rack_name"]
	resp := u.Message(true, "success")
	if e != true || e2 != true || e3 != true || e4 != true || e5 != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET DEVICESUSINGNAMEDRACKOFTENANT", "", r)
		return
	}

	data, e1 := models.GetDevicesUsingNamedRackOfTenant(name, site_name, bldg_name, room_name, rack_name)
	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Devices: "+e1)
		u.ErrLog("Error while getting Devices Using Named Rack of Tenant",
			"GET DEVICESUSINGNAMEDRACKOFTENANT", e1, r)

		switch e1 {
		case "record not found":
			w.WriteHeader(http.StatusNotFound)
		case "":
			w.WriteHeader(http.StatusNotFound)
			resp["message"] = "Error: No Records Found"
		default:
		}

	} else {
		resp = u.Message(true, "success")
	}

	resp["data"] = map[string]interface{}{"objects": data}
	u.Respond(w, resp)
}

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name} tenants GetFromTenant
// Gets all Devices of a Rack by name of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// - name: room_name
//   in: path
//   description: Name of desired room
//   required: true
//   type: string
//   default: "R1"
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

var GetNamedDeviceOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	site_name, e2 := mux.Vars(r)["site_name"]
	bldg_name, e3 := mux.Vars(r)["building_name"]
	room_name, e4 := mux.Vars(r)["room_name"]
	rack_name, e5 := mux.Vars(r)["rack_name"]
	dev_name, e6 := mux.Vars(r)["device_name"]
	resp := u.Message(true, "success")
	if e != true || e2 != true || e3 != true || e4 != true || e5 != true || e6 != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET NAMEDDEVICEOFTENANT", "", r)
		return
	}

	data, e1 := models.GetNamedDeviceOfTenant(name, site_name, bldg_name, room_name, rack_name, dev_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Device: "+e1)
		u.ErrLog("Error while getting Named Device of Tenant",
			"GET NAMEDDEVICEOFTENANT", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/all/sites/buildings tenants GetTenantHierarchy
// Gets hierarchy of Tenant until buildings.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetTenantHierarchyToBuilding = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	resp := u.Message(true, "success")
	if e != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET TENANTHIERARCHTOBLDG", "", r)
		return
	}

	data, e1 := models.GetTenantHierarchyToBuilding(name)
	if data == nil {
		resp = u.Message(false, "Error while getting Tenant Hierarchy: "+e1)
		u.ErrLog("Error while getting Tenant Hierarchy",
			"GET TENANTHIERARCHTOBLDG", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/all/sites/buildings/rooms tenants GetTenantHierarchy
// Gets hierarchy of Tenant until rooms.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetTenantHierarchyToRoom = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	resp := u.Message(true, "success")
	if e != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET TENANTHIERARCHTOROOM", "", r)
		return
	}

	data, e1 := models.GetTenantHierarchyToRoom(name)
	if data == nil {
		resp = u.Message(false, "Error while getting Tenant Hierarchy: "+e1)
		u.ErrLog("Error while getting Tenant Hierarchy",
			"GET TENANTHIERARCHTOROOM", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/all/sites/buildings/rooms/devices tenants GetTenantHierarchy
// Gets hierarchy of Tenant until devices.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetTenantHierarchyToDevice = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	resp := u.Message(true, "success")
	if e != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET TENANTHIERARCHTODEVICE", "", r)
		return
	}

	data, e1 := models.GetTenantHierarchyToDevice(name)
	if data == nil {
		resp = u.Message(false, "Error while getting Tenant Hierarchy: "+e1)
		u.ErrLog("Error while getting Tenant Hierarchy",
			"GET TENANTHIERARCHTODEVICE", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/all/sites/buildings/rooms/racks tenants GetTenantHierarchy
// Gets hierarchy of Tenant until racks.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetTenantHierarchyToRack = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	resp := u.Message(true, "success")
	if e != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET TENANTHIERARCHTORACK", "", r)
		return
	}

	data, e1 := models.GetTenantHierarchyToRack(name)
	if data == nil {
		resp = u.Message(false, "Error while getting Tenant Hierarchy: "+e1)
		u.ErrLog("Error while getting Tenant Hierarchy",
			"GET TENANTHIERARCHTORACK", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{subdevice_name}/subdevices tenants GetFromTenant
// Gets all Subdevices of a Named Device of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// - name: room_name
//   in: path
//   description: Name of desired room
//   required: true
//   type: string
//   default: "R1"
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
//   default: "Rack01"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetSubdevicesUsingNamedDeviceOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["tenant_name"]
	site_name := mux.Vars(r)["site_name"]
	bldg_name := mux.Vars(r)["building_name"]
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	device_name := mux.Vars(r)["device_name"]
	resp := u.Message(true, "success")

	data, e1 := models.GetSubdevicesUsingNamedDeviceOfTenant(name, site_name,
		bldg_name, room_name, rack_name, device_name)
	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Subdevices: "+e1)
		u.ErrLog("Error while getting Subdevices using Named Device of Tenant",
			"GETSUBDEVSUSINGNAMEDDEVICEOFTENANT", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name} tenants GetFromTenant
// Gets named Subdevices of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// - name: room_name
//   in: path
//   description: Name of desired room
//   required: true
//   type: string
//   default: "R1"
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
//   default: "SubDeviceA"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetNamedSubdeviceOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["tenant_name"]
	site_name := mux.Vars(r)["site_name"]
	bldg_name := mux.Vars(r)["building_name"]
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	dev_name := mux.Vars(r)["device_name"]
	subdev_name := mux.Vars(r)["subdevice_name"]
	resp := u.Message(true, "success")

	data, e1 := models.GetNamedSubdeviceOfTenant(name, site_name, bldg_name, room_name, rack_name, dev_name, subdev_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevice: "+e1)
		u.ErrLog("Error while getting Named Subdevice of Tenant",
			"GET NAMEDSUBDEVOFTENANT", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{subdevice_name}/subdevices/{subdevice_name}/subdevice1s/{subdevice1_name} tenants GetFromTenant
// Gets all Subdevice1s of a Named Subdevice of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// - name: room_name
//   in: path
//   description: Name of desired room
//   required: true
//   type: string
//   default: "R1"
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
//   default: "Rack01"
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
var GetSubdevice1sUsingNamedSubdeviceOfTenant = func(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["tenant_name"]
	site_name := mux.Vars(r)["site_name"]
	bldg_name := mux.Vars(r)["building_name"]
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	device_name := mux.Vars(r)["device_name"]
	subdevice_name := mux.Vars(r)["subdevice_name"]
	resp := u.Message(true, "success")

	data, e1 := models.GetSubdevice1sUsingNamedSubdeviceOfTenant(name, site_name,
		bldg_name, room_name, rack_name, device_name, subdevice_name)
	if data == nil || len(data) == 0 {
		resp = u.Message(false, "Error while getting Subdevice1s: "+e1)
		u.ErrLog("Error while getting Subdevice1s using Named Subdevice of Tenant",
			"GETSUBDEV1SUSINGNAMEDDEVICEOFTENANT", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/sites/{site_name}/buildings/{building_name}/rooms/{room_name}/racks/{rack_name}/devices/{device_name}/subdevices/{subdevice_name}/subdevice1s/{subdevice1_name} tenants GetFromTenant
// Gets Subdevice1 of a Tenant from the system.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// - name: site_name
//   in: path
//   description: Name of desired site
//   required: true
//   type: string
//   default: "SiteA"
// - name: building_name
//   in: path
//   description: Name of desired building
//   required: true
//   type: string
//   default: "BldgA"
// - name: room_name
//   in: path
//   description: Name of desired room
//   required: true
//   type: string
//   default: "R1"
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
//   default: "SubDeviceA"
// - name: subdevice1_name
//   in: path
//   description: Name of desired subdevice1
//   required: true
//   type: string
//   default: "SubDevice1A"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetNamedSubdevice1OfTenant = func(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["tenant_name"]
	site_name := mux.Vars(r)["site_name"]
	bldg_name := mux.Vars(r)["building_name"]
	room_name := mux.Vars(r)["room_name"]
	rack_name := mux.Vars(r)["rack_name"]
	dev_name := mux.Vars(r)["device_name"]
	subdev_name := mux.Vars(r)["subdevice_name"]
	subdev1_name := mux.Vars(r)["subdevice1_name"]
	resp := u.Message(true, "success")

	data, e1 := models.GetNamedSubdevice1OfTenant(name, site_name, bldg_name, room_name,
		rack_name, dev_name, subdev_name, subdev1_name)
	if data == nil {
		resp = u.Message(false, "Error while getting Subdevic1e: "+e1)
		u.ErrLog("Error while getting Named Subdevice1 of Tenant",
			"GET NAMEDSUBDEV1OFTENANT", e1, r)

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

// swagger:operation GET /api/user/tenants/{tenant_name}/all/sites/buildings/rooms/devices/subdevices tenants GetTenantHierarchy
// Gets hierarchy of Tenant until subdevices.
// ---
// produces:
// - application/json
// parameters:
// - name: tenant_name
//   in: path
//   description: Name of desired tenant
//   required: true
//   type: string
//   default: "INFINITI"
// responses:
//     '200':
//         description: Found
//     '404':
//         description: Not Found
var GetTenantHierarchyToSubdevice = func(w http.ResponseWriter, r *http.Request) {
	name, e := mux.Vars(r)["tenant_name"]
	resp := u.Message(true, "success")
	if e != true {
		u.Respond(w, u.Message(false, "Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET TENANTHIERARCHTOSUBDEVICE", "", r)
		return
	}

	data, e1 := models.GetTenantHierarchyToSubdevice(name)
	if data == nil {
		resp = u.Message(false, "Error while getting Tenant Hierarchy: "+e1)
		u.ErrLog("Error while getting Tenant Hierarchy",
			"GET TENANTHIERARCHTOSUBDEVICE", e1, r)

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
