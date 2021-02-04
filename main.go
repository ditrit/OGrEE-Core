package main

//Test Case 1
//curl -X POST http://localhost:8000/api/user/login
//--data '{"email": "iamlegend@gmail.com", "password" : "secret"}'

//Test Case 2
//curl -H 'Accept: application/json'
//-H "Authorization: Bearer ${jj}" http://localhost:8000/api/me/contacts

//Test Case 3
//curl -X GET -H 'Content-Type: application/json' \
//-H 'Authorization: bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.
//eyJVc2VySWQiOjYyNTgwMjc5ODYzNjIwNDAzM30.
//FM-K77j5439O1irfJU_O8Rx7VlVkyGpuwmi87tWLTzU' \
//-i 'http://127.0.0.1:8000/api/me/contacts'

//Test Case 4 == FALSE
//curl -X POST http://localhost:8000/api/user/login
//--data '{"email": "realcheat@gmail.com", "password": "secret"}'

//Test Case 5
//curl -X POST -H 'Content-Type: application/json'
//-H 'Authorization: bearer TOKEN'
//--data '{"name": "SiteC", "category": "stupid",
//"domain": 630085240294703105, "description": "Cant evalutate anything",
//"color": "orange", "eorientation": "NW"}'
//-i 'http://localhost:8000/api/user/sites/new'

//Test Case 6
//curl -X POST -H 'Content-Type: application/json'
//-H 'Authorization: bearer TOKEN'
//--data '{"name": "CEA", "category": "research",
//"description": "Cant evalutate anything", "color": "red"}'
//-i 'http://localhost:8000/api/user/tenants/new'

//Test Case 7
//curl -X POST http://localhost:8000/api/user/new
//--data '{"email": "iamlegend@gmail.com", "password" : "secret"}'

//Test Case 9
//curl -X GET
//http://localhost:8000/api/user/sites -H 'Authorization: bearer TOKEN'

import (
	"fmt"
	"p3/app"
	"p3/controllers"

	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/api/user/new",
		controllers.CreateAccount).Methods("POST")

	router.HandleFunc("/api/user/login",
		controllers.Authenticate).Methods("POST")

	router.HandleFunc("/api/user/tenants/new",
		controllers.CreateTenant).Methods("POST")

	router.HandleFunc("/api/user/tenants",
		controllers.GetTenantFor).Methods("GET")

	router.HandleFunc("/api/user/tenants/all",
		controllers.GetAllTenants).Methods("GET")

	router.HandleFunc("/api/user/tenants/update",
		controllers.UpdateTenant).Methods("PUT")

	router.HandleFunc("/api/user/tenants/delete",
		controllers.DeleteTenant).Methods("DELETE")

	router.HandleFunc("/api/user/sites/new",
		controllers.CreateSite).Methods("POST")

	router.HandleFunc("/api/user/sites",
		controllers.GetSitesByUserID).Methods("GET")

	router.HandleFunc("/api/user/sites/all",
		controllers.GetSitesByParentID).Methods("GET")

	router.HandleFunc("/api/user/site/id",
		controllers.GetSite).Methods("GET")

	router.HandleFunc("/api/user/site/",
		controllers.DeleteSiteByID).Methods("DELETE")

	router.HandleFunc("/api/user/sites/",
		controllers.DeleteSites).Methods("DELETE")

	/*router.HandleFunc("/api/user/sites/single",
	controllers.GetSite).Methods("GET")*/

	/*router.HandleFunc("/api/me/contacts",
		controllers.GetContactsFor).Methods("GET")

	router.HandleFunc("/api/contacts/new",
		controllers.CreateContact).Methods("POST")*/

	//Attach JWT auth middleware
	router.Use(app.JwtAuthentication)

	//Get port from .env file, no port was specified
	//So this should return an empty string when
	//tested locally
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" //localhost
	}

	fmt.Println(port)

	//Start app, localhost:8000/api
	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		fmt.Print(err)
	}
}

//https://medium.com/@adigunhammedolalekan/build-and-deploy-a-secure-rest-api-with-go-postgresql-jwt-and-gorm-6fadf3da505b
