package controllers

import (
	"fmt"
	"net/http"
	"p3/models"
	u "p3/utils"

	"github.com/gorilla/mux"
)

// swagger:operation GET /api/schemas/{id} Objects GetSchemaJSON
// Gets a JSON schema by its file name (id).
// Returns a HTTP response with the content of the schema in the body (json format).
// JSON schemas define the format of each entity of the API.
// ---
// security:
// - bearer: []
// produces:
// - application/json
// parameters:
// - name: id
//   in: path
//   description: Name of the JSON schema
//   required: true
//   type: string
// responses:
//     '200':
//         description: Found. A response body will be returned with the schema content.
//     '404':
//         description: Not Found. An error message will be returned.

func GetSchemaJSON(w http.ResponseWriter, r *http.Request) {
	fmt.Println("******************************************************")
	fmt.Println("FUNCTION CALL: 	 GetSchemaJSON ")
	fmt.Println("******************************************************")
	DispRequestMetaData(r)

	file, err := models.GetSchemaFile(mux.Vars(r)["id"])
	if err != nil {
		u.ErrLog("Error while getting schema",
			"GET GetSchemaJSON", err.Message, r)
		u.RespondWithError(w, err)
	} else if r.Method == "OPTIONS" {
		u.WriteOptionsHeader(w, "GET")
	} else {
		w.Header().Add("Content-Type", u.HttpResponseContentType)
		w.Write(file)
	}
}
