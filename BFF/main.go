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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ogree-bff/models"
	"ogree-bff/services"
	"os"

	e "github.com/joho/godotenv"
)


func GetAPIInfo() ([]models.API) {
	jsonFile, err := os.Open("api.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var api []models.API
	json.Unmarshal(byteValue, &api)

	return api
}

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
	apiList := GetAPIInfo()
	fmt.Println(apiList)
	router := services.InitRouter(apiList,env)
	router.Run(":"+BFF_PORT)
}
