//	 Kube-admin:
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
	"flag"
	"kube-admin/services"
	"strconv"

	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}
}

func main() {
	port := flag.Int("port", 8081, "an int for the port used by the backend")
	isKube := flag.Bool("kube", false, "a boolean to use kubernetes mode instead of docker")
	flag.Parse()

	router := services.InitRouter(*isKube)
	router.Run(":" + strconv.Itoa(*port))
}
