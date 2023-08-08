//	 Ogree BFF:
//	  version: 1.0.0
//	  title: Awsome API
//	 Schemes: http, https
//	 Host:
//	 BasePath: /api
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
	e "github.com/joho/godotenv"
	"ogree-bff/models"
	"ogree-bff/services"
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
	BFF_PORT := os.Getenv("BFF_PORT")
	arangoAPI := os.Getenv("ARANGO_API")
	mongoAPI := os.Getenv("MONGO_API")
	apiList := []models.API {
		{Name: "arango", URL: arangoAPI},
		{Name: "mongo", URL: mongoAPI},
	}
	fmt.Println(apiList)
	router := services.InitRouter(apiList,env)
	router.Run(":"+BFF_PORT)
}
