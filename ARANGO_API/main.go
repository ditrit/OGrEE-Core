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
	"fmt"
	"os"
	"go-api/services"
	"go-api/database"
)

func main() {
	addr := os.Getenv("ARRANGO_URL")
	bdd := os.Getenv("ARRANGO_DATABASE")
	user := os.Getenv("ARRANGO_USER")
	password := os.Getenv("ARRANGO_PASSWORD")

	
	db, err := database.ConnectToArrengo(addr,bdd, user, password)
	if err != nil {
		fmt.Println("Error connecting to database: ", err.Message)
		return
	}

	database.CreateCollection(db, "devices")
	
	router := services.InitRouter(db,addr)
	router.Run(":8080")
}
