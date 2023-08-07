//	 Arrango API:
//	  version: 1.0.0
//	  title: Awsome API
//	 Schemes: http, https
//	 Host:
//	 BasePath: /api/v1
//		Consumes:
//		- application/json
//	 Produces:
//	 - application/json
//	 SecurityDefinitions:
//	  Bearer:
//	   type: apiKey
//	   name: Authorization
//	   in: header
//	 swagger:meta
package main

import (
	"arango-api/database"
	"arango-api/services"
	"fmt"
	"os"
	e "github.com/joho/godotenv"
)

func main() {

	env := os.Getenv("ENV")
	if env != "production" {
		fmt.Println("Loading environment variables from .env")
		err := e.Load()
		if err != nil {
			fmt.Println("Some error occured. Err: ", err)
			return
		}
	}

	addr := os.Getenv("ARANGO_URL")
	bdd := os.Getenv("ARANGO_DATABASE")
	user := os.Getenv("ARANGO_USER")
	password := os.Getenv("ARANGO_PASSWORD")

	
	db, err2 := database.ConnectToArango(addr,bdd, user, password)
	if err2 != nil {
		fmt.Println("Error connecting to database: ", err2.Message)
		return
	}

	database.CreateCollection(db, "devices")
	
	router := services.InitRouter(db,addr)
	router.Run(":8080")
}
