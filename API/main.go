package main

import (
	"fmt"
	"p3/app"
	"p3/controllers"

	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"
)

// Obtain by query
var dmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	//fmt.Println("The URL is: ", request.URL.String())
	//https://benhoyt.com/writings/go-routing/#regex-table
	//https://stackoverflow.com/questions/21664489/
	//golang-mux-routing-wildcard-custom-func-match
	println("Checking MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|acs|panels|cabinets|groups|corridors|racks|devices|sensors|stray-(devices|sensors)|(room|obj|bldg)-templates)\?.*)$`).
		MatchString(request.URL.String())
}

// Obtain object hierarchy
var hmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING H-MATCH")
	return regexp.MustCompile(`(^(\/api\/(tenant|site|building|room|room|rack|device|stray-device)\/[a-zA-Z0-9]{24}\/all)(\/(tenants|sites|buildings|rooms|rooms|racks|devices|stray-(devices|sensors)))*$)|(^(\/api\/(tenants|sites|buildings|rooms|rooms|racks|devices|stray-devices)\/[a-zA-Z0-9]{24}\/all)(\?limit=[0-9]+)*$)`).
		MatchString(request.URL.String())
}

// For Obtaining objects using parent
var pmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING P-MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|rooms|racks|devices|stray-devices)\/[a-zA-Z0-9]{24}(\/.*)+)$`).
		MatchString(request.URL.String())
}

// For Obtaining Tenant hierarchy
var tmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING T-MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|stray-devices)(\/[A-Za-z0-9_]+)(\/.*)+)$`).
		MatchString(request.URL.String())
}

// For Obtaining hierarchy with hierarchyName
var hnmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING HN-MATCH")
	return regexp.MustCompile(`^\/api\/(tenants|sites|buildings|rooms|racks|devices|stray-devices)+\/[A-Za-z0-9_.]+\/all(\?limit=[0-9]+)*$`).
		MatchString(request.URL.String())
}

func Router(jwt func(next http.Handler) http.Handler) *mux.Router {
	router := mux.NewRouter()

	router.HandleFunc("/api",
		controllers.CreateAccount).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/stats",
		controllers.GetStats).Methods("GET", "OPTIONS", "HEAD")

	router.HandleFunc("/api/login",
		controllers.Authenticate).Methods("POST", "OPTIONS")

	router.HandleFunc("/api/token/valid",
		controllers.Verify).Methods("GET", "OPTIONS", "HEAD")

	router.HandleFunc("/api/version",
		controllers.Version).Methods("GET", "OPTIONS", "HEAD")

	// For obtaining temperatureUnit from object's site
	router.HandleFunc("/api/tempunits/{id}",
		controllers.GetTempUnit).Methods("GET", "OPTIONS", "HEAD")

	// For obtaining the complete hierarchy (tree)
	router.HandleFunc("/api/hierarchy",
		controllers.GetCompleteHierarchy).Methods("GET", "OPTIONS", "HEAD")

	// ------ GET ------ //
	router.HandleFunc("/api/objects/{name}",
		controllers.GetGenericObject).Methods("GET", "HEAD", "OPTIONS")

	//GET ENTITY HIERARCHY
	//This matches ranged Tenant Hierarchy
	router.NewRoute().PathPrefix("/api/{entity}s/{id:[a-zA-Z0-9]{24}}/all").
		MatcherFunc(hmatch).HandlerFunc(controllers.GetEntityHierarchy).Methods("GET", "HEAD", "OPTIONS")

	router.NewRoute().PathPrefix("/api/{entity}s/{name}/all").
		MatcherFunc(hnmatch).HandlerFunc(controllers.GetHierarchyByName).Methods("GET", "HEAD", "OPTIONS")

	//GET EXCEPTIONS
	router.HandleFunc("/api/{ancestor:tenant}s/{tenant_name}/buildings",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/{ancestor:site}s/{id:[a-zA-Z0-9]{24}}/rooms",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/{ancestor:building}s/{id:[a-zA-Z0-9]{24}}/{sub:ac|corridor|cabinet|panel|sensor|group}s",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/{ancestor:building}s/{id:[a-zA-Z0-9]{24}}/racks",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/{ancestor:room}s/{id:[a-zA-Z0-9]{24}}/devices",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	// GET BY QUERY
	router.NewRoute().PathPrefix("/api/{entity:[a-z]+}").MatcherFunc(dmatch).
		HandlerFunc(controllers.GetEntityByQuery).Methods("HEAD", "GET")

	//GET ENTITY
	router.HandleFunc("/api/{entity}s/{id:[a-zA-Z0-9]{24}}",
		controllers.GetEntity).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/{entity}s/{name}",
		controllers.GetEntity).Methods("GET", "HEAD", "OPTIONS")

	//GET BY NAME OF PARENT
	router.NewRoute().PathPrefix("/api/{entity}s/{tenant_name}").
		MatcherFunc(tmatch).HandlerFunc(controllers.GetEntitiesUsingNamesOfParents).Methods("GET", "HEAD", "OPTIONS")

	router.NewRoute().PathPrefix("/api/{entity}s/{id:[a-zA-Z0-9]{24}}").
		MatcherFunc(pmatch).HandlerFunc(controllers.GetEntitiesUsingNamesOfParents).Methods("GET", "HEAD", "OPTIONS")

	// GET ALL ENTITY

	router.HandleFunc("/api/{entity}s",
		controllers.GetAllEntities).Methods("HEAD", "GET")

	// CREATE ENTITY
	router.HandleFunc("/api/{entity}s",
		controllers.CreateEntity).Methods("POST")

	//DELETE ENTITY
	router.HandleFunc("/api/{entity}s/{id:[a-zA-Z0-9]{24}}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/{entity}s/{name}",
		controllers.DeleteEntity).Methods("DELETE")

	// UPDATE ENTITY
	router.HandleFunc("/api/{entity}s/{id:[a-zA-Z0-9]{24}}",
		controllers.UpdateEntity).Methods("PUT", "PATCH")

	router.HandleFunc("/api/{entity}s/{name}",
		controllers.UpdateEntity).Methods("PUT", "PATCH")

	//OPTIONS BLOCK
	router.HandleFunc("/api/{entity}s",
		controllers.BaseOption).Methods("OPTIONS")

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
	port := os.Getenv("api_port")
	if port == "" {
		port = "3001" //localhost
	}

	fmt.Println(port)

	//Start app, localhost:8000/api
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		fmt.Print(err)
	}
}

//https://medium.com/@adigunhammedolalekan/build-and-deploy-a-secure-rest-api-with-go-postgresql-jwt-and-gorm-6fadf3da505b
