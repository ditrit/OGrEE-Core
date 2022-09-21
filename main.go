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

//Obtain by query
var dmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {

	//fmt.Println("The URL is: ", request.URL.String())
	//https://benhoyt.com/writings/go-routing/#regex-table
	//https://stackoverflow.com/questions/21664489/
	//golang-mux-routing-wildcard-custom-func-match
	println("Checking MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|acs|panels|cabinets|groups|corridors|racks|devices|sensors|stray-(devices|sensors)|(room|obj)-templates)\?.*)$`).
		MatchString(request.URL.String())
}

//Obtain object hierarchy
var hmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING H-MATCH")
	return regexp.MustCompile(`(^(\/api\/(tenants|sites|buildings|rooms|rooms|racks|devices|stray-devices)\/[a-zA-Z0-9]{24}\/all)(\/(tenants|sites|buildings|rooms|rooms|racks|devices|stray-(devices|sensors)))*$)|(^(\/api\/(tenants|sites|buildings|rooms|rooms|racks|devices|stray-devices)\/[a-zA-Z0-9]{24}\/all)(\?limit=[0-9]+)*$)`).
		MatchString(request.URL.String())
}

//For Obtaining objects using parent
var pmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING P-MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|sites|buildings|rooms|rooms|racks|devices|stray-devices)\/[a-zA-Z0-9]{24}(\/.*)+)$`).
		MatchString(request.URL.String())
}

//For Obtaining Tenant hierarchy
var tmatch mux.MatcherFunc = func(request *http.Request, match *mux.RouteMatch) bool {
	println("CHECKING T-MATCH")
	return regexp.MustCompile(`^(\/api\/(tenants|stray-devices)(\/[A-Za-z0-9_]+)(\/.*)+)$`).
		MatchString(request.URL.String())
}

func main() {

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

	// ------ GET ------ //
	//GET ENTITY HIERARCHY
	//This matches ranged Tenant Hierarchy
	router.NewRoute().PathPrefix("/api/{entity}/{id:[a-zA-Z0-9]{24}}/all").
		MatcherFunc(hmatch).HandlerFunc(controllers.GetEntityHierarchy).Methods("GET", "HEAD", "OPTIONS")

	router.NewRoute().PathPrefix("/api/{entity}/{name}/all").
		MatcherFunc(tmatch).HandlerFunc(controllers.GetHierarchyByName).Methods("GET", "HEAD", "OPTIONS")

	//GET EXCEPTIONS
	router.HandleFunc("/api/tenants/{tenant_name}/buildings",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/sites/{id:[a-zA-Z0-9]{24}}/rooms",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/buildings/{id:[a-zA-Z0-9]{24}}/{sub:acs|corridors|cabinets|panels|sensors|groups}",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/buildings/{id:[a-zA-Z0-9]{24}}/racks",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/rooms/{id:[a-zA-Z0-9]{24}}/devices",
		controllers.GetEntitiesOfAncestor).Methods("GET", "HEAD", "OPTIONS")

	/*router.HandleFunc("/api/rooms/{id:[a-zA-Z0-9]{24}}/sensors",
		controllers.GetEntitiesOfAncestor).Methods("GET")

	router.HandleFunc("/api/racks/{id:[a-zA-Z0-9]{24}}/sensors",
		controllers.GetEntitiesOfAncestor).Methods("GET")*/

	// GET BY QUERY
	router.NewRoute().PathPrefix("/api/{entity:[a-z]+}").MatcherFunc(dmatch).
		HandlerFunc(controllers.GetEntityByQuery).Methods("HEAD", "GET")

	//GET ENTITY
	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}",
		controllers.GetEntity).Methods("GET", "HEAD", "OPTIONS")

	router.HandleFunc("/api/{entity}/{name}",
		controllers.GetEntity).Methods("GET", "HEAD", "OPTIONS")

	//GET BY NAME OF PARENT
	router.NewRoute().PathPrefix("/api/{entity}/{tenant_name}").
		MatcherFunc(tmatch).HandlerFunc(controllers.GetEntitiesUsingNamesOfParents).Methods("GET", "HEAD", "OPTIONS")

	router.NewRoute().PathPrefix("/api/{entity}/{id:[a-zA-Z0-9]{24}}").
		MatcherFunc(pmatch).HandlerFunc(controllers.GetEntitiesUsingNamesOfParents).Methods("GET", "HEAD", "OPTIONS")

	// GET ALL ENTITY

	router.HandleFunc("/api/{entity}",
		controllers.GetAllEntities).Methods("HEAD", "GET")

	//GET ALL NONSTD
	router.HandleFunc("/api/tenants/{tenant_name}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}/all/nonstd",
		controllers.GetEntityHierarchyNonStd).Methods("GET")

	// CREATE ENTITY
	router.HandleFunc("/api/{entity}",
		controllers.CreateEntity).Methods("POST")

	//DELETE ENTITY
	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}",
		controllers.DeleteEntity).Methods("DELETE")

	router.HandleFunc("/api/{entity}/{name}",
		controllers.DeleteEntity).Methods("DELETE")

	// UPDATE ENTITY
	router.HandleFunc("/api/{entity}/{id:[a-zA-Z0-9]{24}}",
		controllers.UpdateEntity).Methods("PUT", "PATCH")

	router.HandleFunc("/api/{entity}/{name}",
		controllers.UpdateEntity).Methods("PUT", "PATCH")

	//OPTIONS BLOCK
	router.HandleFunc("/api/{entity}",
		controllers.BaseOption).Methods("OPTIONS")

	//VALIDATION
	router.HandleFunc("/api/validate/{entity}", controllers.ValidateEntity).Methods("POST", "OPTIONS")

	//Attach JWT auth middleware
	//router.Use(app.Log)
	router.Use(app.JwtAuthentication)

	//TODO:
	//Use the URL below to help make the router functions more
	//flexible and thus implement the http OPTIONS method
	//cleanly
	//https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81

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
