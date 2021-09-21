package controllers

import (
	"encoding/json"
	"net/http"
	"p3/models"
	u "p3/utils"
	"strings"
)

var CreateEntity = func(w http.ResponseWriter, r *http.Request) {
	//tenant := &models.Tenant{}
	entStr := r.URL.Path[5 : len(r.URL.Path)-1] //strip the '/api' in URL
	entUpper := strings.ToUpper(entStr)         // and the trailing 's'
	entity := map[string]interface{}{}
	err := json.NewDecoder(r.Body).Decode(&entity)

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
		u.ErrLog("Error while decoding request body", "CREATE "+entStr, "", r)

		return
	}

	i := u.EntityStrToInt(entStr)
	println("ENT: ", entStr)
	println("ENUM VAL: ", i)

	resp, e := models.CreateEntity(i, entity)

	switch e {
	case "validate":
		w.WriteHeader(http.StatusBadRequest)
		u.ErrLog("Error while creating "+entStr, "CREATE "+entUpper, e, r)
	case "":
		w.WriteHeader(http.StatusCreated)
	default:
		if strings.Split(e, " ")[1] == "duplicate" {
			w.WriteHeader(http.StatusBadRequest)
			u.ErrLog("Error: Duplicate "+entStr+" is forbidden",
				"CREATE "+entUpper, e, r)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			u.ErrLog("Error while creating "+entStr, "CREATE "+entUpper, e, r)
		}
	}

	u.Respond(w, resp)
}
