package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var tmplt *template.Template

func init() {
	tmplt = template.Must(template.ParseFiles("docker-env-template.txt"))
}

func main() {
	router := gin.Default()
	router.Use(cors.Default())

	router.GET("/api/tenants", getTenants)
	router.POST("/api/tenants", addTenant)
	router.POST("/api/login", login)

	router.Run(":8081")
}

// TENANTS
type tenant struct {
	Name             string `json:"name" binding:"required"`
	CustomerPassword string `json:"customerPassword"`
	ApiUrl           string `json:"apiUrl"`
	WebUrl           string `json:"webUrl"`
	ApiPort          string `json:"apiPort"`
	WebPort          string `json:"webPort"`
}

type user struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Token    string `json:"token"`
}

func getTenants(c *gin.Context) {
	data, e := ioutil.ReadFile("tenants.json")
	if e != nil {
		panic(e.Error())
	}
	var listTenants []tenant
	json.Unmarshal(data, &listTenants)
	fmt.Println(listTenants)
	response := make(map[string][]tenant)
	response["tenants"] = listTenants
	c.IndentedJSON(http.StatusOK, response)
}

func addTenant(c *gin.Context) {
	data, e := ioutil.ReadFile("tenants.json")
	if e != nil {
		panic(e.Error())
	}
	var listTenants []tenant
	json.Unmarshal(data, &listTenants)

	// Call BindJSON to bind the received JSON
	var newTenant tenant
	if err := c.BindJSON(&newTenant); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	} else {
		// Create .env file
		file, _ := os.Create("docker/.env")
		err = tmplt.Execute(file, newTenant)
		if err != nil {
			panic(err)
		}
		file.Close()

		// Docker compose up
		cmd := exec.Command("docker-compose", "-p", strings.ToLower(newTenant.Name), "up", "-d")
		cmd.Dir = "docker/"
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if _, err := cmd.Output(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			c.IndentedJSON(http.StatusInternalServerError, stderr.String())
			return
		}

		// Add to local json
		listTenants = append(listTenants, newTenant)
		data, _ := json.MarshalIndent(listTenants, "", "  ")
		_ = ioutil.WriteFile("tenants.json", data, 0644)
		c.IndentedJSON(http.StatusOK, "all good")
	}

}

func login(c *gin.Context) {
	var userIn user

	if err := c.BindJSON(&userIn); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
	} else {
		response := make(map[string]map[string]string)
		response["account"] = make(map[string]string)
		response["account"]["Email"] = userIn.Email
		response["account"]["token"] = userIn.Password
		c.IndentedJSON(http.StatusOK, response)
	}
}
