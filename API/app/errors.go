package app

import (
	"net/http"
	u "p3/utils"
)

//A more sophisticated error system might be
//better later on
/*
var ErrorCatalogue map[int]string{
	0: "",
	1: "",
	2: "",
	3: "",
	4: "",
	5: "",
}
*/

var NotFoundHandler = func(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		u.Respond(w, u.Message("This resources was not found"))
		next.ServeHTTP(w, r)
	})
}
