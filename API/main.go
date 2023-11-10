package main

import (
	"fmt"
	"log"
	"p3/app"
	"p3/repository"
	"p3/router"

	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/joho/godotenv"
)

func connectToDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}

	err = repository.ConnectToDB(
		os.Getenv("db_host"),
		os.Getenv("db_port"),
		"ogree"+os.Getenv("db_user")+"Admin",
		os.Getenv("db_pass"),
		"ogree"+os.Getenv("db"),
		os.Getenv("db"),
	)
	if err != nil {
		log.Fatalln(err.Error())
	}

	log.Println("Successfully connected to DB")
}

func main() {
	connectToDB()
	//TODO:
	//Use the URL below to help make the router functions more
	//flexible and thus implement the http OPTIONS method
	//cleanly
	//https://medium.com/@matryer/writing-middleware-in-golang-and-how-go-makes-it-so-much-fun-4375c1246e81
	router := router.Router(app.JwtAuthentication)

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
