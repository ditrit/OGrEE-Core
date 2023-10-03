package main

import (
	"fmt"
	"p3/app"
	"p3/controllers"

	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Obtain by query
var dmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("Checking MATCH")
	return regexp.MustCompile(`^(\/api\/(domains|sites|buildings|rooms|acs|panels|cabinets|groups|corridors|racks|devices|stray-objects|(room|obj|bldg)-templates|tags)\?.*)$`).
		MatchString(request.URL.String())
}

// For Obtaining hierarchy with hierarchyName
var hnmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING HN-MATCH")
	return regexp.MustCompile(`^\/api\/(sites|buildings|rooms|racks|devices|stray-objects|domains|objects)+\/[A-Za-z0-9_.-]+\/all(\?.*)*$`).
		MatchString(request.URL.String())
}

func Router(jwt func(next http.Handler) http.Handler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api/stats",
		controllers.GetStats).Methods("GET", "OPTIONS", "HEAD")

	router.HandleFunc("/api/version",
		controllers.GetVersion).Methods("GET", "OPTIONS", "HEAD")

	// User and Authentication
	router.HandleFunc("/api/login",
		controllers.Authenticate).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/token/valid",
		controllers.VerifyToken).Methods("GET", "OPTIONS", "HEAD")

	router.HandleFunc("/api/users",
		controllers.CreateAccount).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/users/bulk",
		controllers.CreateBulkAccount).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/users",
		controllers.GetAllAccounts).Methods("GET", "OPTIONS", "HEAD")

	router.HandleFunc("/api/users/{id}",
		controllers.RemoveAccount).Methods("DELETE", "OPTIONS")

	router.HandleFunc("/api/users/{id}",
		controllers.ModifyUserRoles).Methods("PATCH", "OPTIONS")

	router.HandleFunc("/api/users/password/change",
		controllers.ModifyUserPassword).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/users/password/reset",
		controllers.ModifyUserPassword).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/users/password/forgot",
		controllers.UserForgotPassword).Methods("POST", "OPTIONS")

	// For obtaining temperatureUnit from object's site
	router.HandleFunc("/api/tempunits/{id}",
		controllers.GetTempUnit).Methods("GET", "OPTIONS", "HEAD")

	// For obtaining the complete hierarchy (tree)
	router.HandleFunc("/api/hierarchy",
		controllers.GetCompleteHierarchy).Methods("GET", "OPTIONS", "HEAD")

	router.HandleFunc("/api/hierarchy/attributes",
		controllers.GetCompleteHierarchyAttributes).Methods("GET", "OPTIONS", "HEAD")

	// FLUTTER FRONT
	router.HandleFunc("/api/projects",
		controllers.GetProjects).Methods("HEAD", "GET", "OPTIONS")

	router.HandleFunc("/api/projects",
		controllers.CreateOrUpdateProject).Methods("POST")

	router.HandleFunc("/api/projects/{id:[a-zA-Z0-9]{24}}",
		controllers.CreateOrUpdateProject).Methods("PUT")

	router.HandleFunc("/api/projects/{id:[a-zA-Z0-9]{24}}",
		controllers.DeleteProject).Methods("DELETE", "OPTIONS")

	// GENERIC
	router.HandleFunc("/api/objects/{id}",
		controllers.HandleGenericObject).Methods("GET", "HEAD", "OPTIONS", "DELETE")

	router.HandleFunc("/api/objects-wildcard/{id}",
		controllers.HandleGenericObjectWildcard).Methods("GET", "HEAD", "OPTIONS", "DELETE")

	//GET ENTITY HIERARCHY
	router.NewRoute().PathPrefix("/api/{entity}s/{id}/all").
		MatcherFunc(hnmatch).HandlerFunc(controllers.GetHierarchyByName).Methods("GET", "HEAD", "OPTIONS")

	//GET SUBENT
	router.HandleFunc("/api/{ancestor:site|building|room|rack}s/{id}/{sub:building|room|ac|corridor|cabinet|panel|group|rack|device}s",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	// GET BY QUERY
	router.NewRoute().PathPrefix("/api/{entity:[a-z]+}").MatcherFunc(dmatch).
		HandlerFunc(controllers.GetEntityByQuery).Methods("HEAD", "GET")

	//GET ENTITY
	router.HandleFunc("/api/{entity}s/{id}",
		controllers.GetEntity).Methods("GET", "HEAD", "OPTIONS")

	// GET ALL ENTITY
	router.HandleFunc("/api/{entity}s",
		controllers.GetAllEntities).Methods("HEAD", "GET")

	// CREATE ENTITY
	router.HandleFunc("/api/{entity}s",
		controllers.CreateEntity).Methods("POST")

	router.HandleFunc("/api/domains/bulk",
		controllers.CreateBulkDomain).Methods("POST")

	//DELETE ENTITY
	router.HandleFunc("/api/{entity}s/{id}",
		controllers.DeleteEntity).Methods("DELETE")

	// UPDATE ENTITY
	router.HandleFunc("/api/{entity}s/{id}",
		controllers.UpdateEntity).Methods("PUT", "PATCH")

	//OPTIONS BLOCK
	router.HandleFunc("/api/{entity}s",
		controllers.BaseOption).Methods("OPTIONS")

	// LINK AND UNLINK
	router.HandleFunc("/api/{entity:building|room|ac|corridor|cabinet|panel|group|rack|device}s/{id}/unlink",
		controllers.LinkEntity).Methods("PATCH")

	router.HandleFunc("/api/stray-objects/{id}/link",
		controllers.LinkEntity).Methods("PATCH")

	//VALIDATION
	router.HandleFunc("/api/validate/{entity}s", controllers.ValidateEntity).Methods("POST", "OPTIONS")

	//Attach JWT auth middleware
	//router.Use(app.Log)
	router.Use(jwt)

	return router
}

func main() {
	//TODO:
	//Use the URL below to help make the router functions more
	//flexible and thus implement the http OPTIONS method
	//cleanly
	//https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81
	router := Router(app.JwtAuthentication)

	//Get port from .env file, no port was specified
	//So this should return an empty string when
	//tested locally
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001" //localhost
	}

	fmt.Println(port)

	//Start app, localhost:8000/api
	corsObj := handlers.AllowedOrigins([]string{"*"})
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Origin", "Accept"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "OPTIONS", "POST", "PUT", "DELETE", "PATCH"})
	err := http.ListenAndServe(":"+port, handlers.CORS(corsObj, headersOk, methodsOk)(router))
	if err != nil {
		fmt.Print(err)
	}
}

//https://medium.com/@adigunhammedolalekan/build-and-deploy-a-secure-rest-api-with-go-postgresql-jwt-and-gorm-6fadf3da505b
