package controllers

import (
	"fmt"
	"net/http"
	"p3/models"
	u "p3/utils"

	"github.com/gorilla/mux"
)

func getImpactFiltersFromQueryParams(r *http.Request) models.ImpactFilters {
	var filters models.ImpactFilters
	fmt.Println(r.URL.Query())
	decoder.Decode(&filters, r.URL.Query())
	fmt.Println(filters)
	return filters
}

// swagger:operation GET /api/impact/{id} Objects GetImpact
// Returns all objects that could directly or indirectly impacted
// by the object sent in the request.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
//   - name: id
//     in: path
//     description: 'ID of target object.'
//     required: true
//     type: string
//     default: "siteA"
//   - name: categories
//     in: query
//     description: 'Categories to include on indirect impact search.
//     Can be repeated to create a list.'
//   - name: ptypes
//     in: query
//     description: 'Physical types to include on indirect impact search.
//     Can be repeated to create a list.'
//   - name: vtypes
//     in: query
//     description: 'Virtual types to include on indirect impact search.
//     Can be repeated to create a list.'
// responses:
//		'200':
//			description: 'Request is valid.'
//		'500':
//			description: Server error.

func GetImpact(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetImpact ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	// Get user roles for permissions
	user := getUserFromToken(w, r)
	if user == nil {
		return
	}

	filters := getImpactFiltersFromQueryParams(r)

	// Get id of impact target
	id, canParse := mux.Vars(r)["id"]
	if !canParse {
		w.WriteHeader(http.StatusBadRequest)
		u.Respond(w, u.Message("Error while parsing path parameters"))
		u.ErrLog("Error while parsing path parameters", "GET ENTITY", "", r)
		return
	}

	data, err := models.GetImpact(id, user.Roles, filters)
	if err != nil {
		u.RespondWithError(w, err)
	} else {
		if r.Method == "OPTIONS" {
			u.WriteOptionsHeader(w, "GET, HEAD")
		} else {
			u.Respond(w, u.RespDataWrapper("successfully got hierarchy", data))
		}
	}
}
