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
)

func main() {
	addr := os.Getenv("ARANGO_URL")
	bdd := os.Getenv("ARANGO_DATABASE")
	user := os.Getenv("ARANGO_USER")
	password := os.Getenv("ARANGO_PASSWORD")

	
	db, err := database.ConnectToArrengo(addr,bdd, user, password)
	if err != nil {
		fmt.Println("Error connecting to database: ", err.Message)
		return
	}

	database.CreateCollection(db, "devices")
	
	router := services.InitRouter(db,addr)
	router.Run(":8080")
}
