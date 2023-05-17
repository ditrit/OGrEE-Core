package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
)

type tenant struct {
	Name             string `json:"name" binding:"required"`
	CustomerPassword string `json:"customerPassword"`
	ApiUrl           string `json:"apiUrl"`
	WebUrl           string `json:"webUrl"`
	ApiPort          string `json:"apiPort"`
	WebPort          string `json:"webPort"`
	DocUrl           string `json:"docUrl"`
	DocPort          string `json:"docPort"`
	HasWeb           bool   `json:"hasWeb"`
	HasDoc           bool   `json:"hasDoc"`
	AssetsDir        string `json:"assetsDir"`
}

type container struct {
	Name       string `json:"Names"`
	RunningFor string `json:"RunningFor"`
	State      string `json:"State"`
	Image      string `json:"Image"`
	Size       string `json:"Size"`
	Ports      string `json:"Ports"`
}

type user struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Token    string `json:"token"`
}

func getTenants(c *gin.Context) {
	data, e := ioutil.ReadFile("tenants.json")
	if e != nil {
		if strings.Contains(e.Error(), "no such file") || strings.Contains(e.Error(), "cannot find") {
			var file, e = os.Create("tenants.json")
			if e != nil {
				panic(e.Error())
			} else {
				file.WriteString("[]")
				file.Sync()
				defer file.Close()
				response := make(map[string][]tenant)
				response["tenants"] = []tenant{}
				c.IndentedJSON(http.StatusOK, response)
				return
			}
		} else {
			panic(e.Error())
		}
	}
	var listTenants []tenant
	json.Unmarshal(data, &listTenants)
	fmt.Println(listTenants)
	response := make(map[string][]tenant)
	response["tenants"] = listTenants
	c.IndentedJSON(http.StatusOK, response)
}

func getTenantDockerInfo(c *gin.Context) {
	name := c.Param("name")
	println(name)
	cmd := exec.Command("docker", "ps", "--all", "--format", "\"{{json .}}\"")
	cmd.Dir = DOCKER_DIR
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if output, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusInternalServerError, stderr.String())
		return
	} else {
		response := []container{}
		s := bufio.NewScanner(bytes.NewReader(output))
		for s.Scan() {
			var dc container
			jsonOutput := s.Text()
			jsonOutput, _ = strings.CutPrefix(jsonOutput, "\"")
			jsonOutput, _ = strings.CutSuffix(jsonOutput, "\"")
			fmt.Println(jsonOutput)
			if err := json.Unmarshal([]byte(jsonOutput), &dc); err != nil {
				//handle error
				fmt.Println(err.Error())
			}
			fmt.Println(dc)
			if strings.Contains(dc.Name, name) {
				response = append(response, dc)
			}
		}
		if s.Err() != nil {
			// handle scan error
			fmt.Println(s.Err().Error())
		}

		c.IndentedJSON(http.StatusOK, response)
	}
}

func getContainerLogs(c *gin.Context) {
	name := c.Param("name")
	println(name)
	cmd := exec.Command("docker", "logs", name, "--tail", "1000")
	cmd.Dir = DOCKER_DIR
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if output, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusInternalServerError, stderr.String())
		return
	} else {
		response := map[string]string{}
		response["logs"] = string(output)

		c.IndentedJSON(http.StatusOK, response)
	}
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
		tenantLower := strings.ToLower(newTenant.Name)

		// Docker compose prepare
		args := []string{"-p", tenantLower}
		if newTenant.HasWeb {
			args = append(args, "--profile")
			args = append(args, "web")
			// Create flutter assets folder with .env and logo
			newTenant.AssetsDir = DOCKER_DIR + "app-deploy/" + tenantLower
			err = os.MkdirAll(newTenant.AssetsDir, 0644)
			if err != nil && !strings.Contains(err.Error(), "already") {
				panic(err)
			}
			file, err := os.Create(newTenant.AssetsDir + "/.env")
			if err != nil {
				panic(err)
			}
			err = apptmplt.Execute(file, newTenant)
			if err != nil {
				panic(err)
			}
			file.Close()
		}
		if newTenant.HasDoc {
			args = append(args, "--profile")
			args = append(args, "doc")
		}
		args = append(args, "up")
		args = append(args, "-d")

		// Create .env file
		file, _ := os.Create(DOCKER_DIR + ".env")
		err = tmplt.Execute(file, newTenant)
		if err != nil {
			panic(err)
		}
		file.Close()
		// Create .env copy
		file, _ = os.Create(DOCKER_DIR + tenantLower + ".env")
		err = tmplt.Execute(file, newTenant)
		if err != nil {
			fmt.Println("Error creating .env copy: " + err.Error())
		}
		file.Close()

		println("Run docker (may take a long time...)")

		// Run docker
		cmd := exec.Command("docker-compose", args...)
		cmd.Dir = DOCKER_DIR
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if _, err := cmd.Output(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			c.IndentedJSON(http.StatusInternalServerError, stderr.String())
			return
		}
		println("Finished with docker")

		// Add to local json and respond
		listTenants = append(listTenants, newTenant)
		data, _ := json.MarshalIndent(listTenants, "", "  ")
		_ = ioutil.WriteFile("tenants.json", data, 0644)
		c.IndentedJSON(http.StatusOK, "all good")
	}

}

func addTenantLogo(c *gin.Context) {
	tenantName := strings.ToLower(c.Param("name"))
	// Load image
	formFile, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	// Make sure destination dir is created
	assetsDir := DOCKER_DIR + "app-deploy/" + tenantName
	err = os.MkdirAll(assetsDir, 0644)
	if err != nil && !strings.Contains(err.Error(), "already") {
		c.String(http.StatusInternalServerError, err.Error())
	}
	// Save image
	err = c.SaveUploadedFile(formFile, assetsDir+"/logo.png")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	c.String(http.StatusOK, "")
}

func removeTenant(c *gin.Context) {
	tenantName := strings.ToLower(c.Param("name"))

	// Stop and remove containers
	for _, str := range []string{"_webapp", "_api", "_db", "_doc"} {
		cmd := exec.Command("docker", "rm", "--force", strings.ToLower(tenantName)+str)
		cmd.Dir = DOCKER_DIR
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if _, err := cmd.Output(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			c.IndentedJSON(http.StatusInternalServerError, stderr.String())
			return
		}
	}

	// Remove assets
	os.RemoveAll(DOCKER_DIR + "app-deploy/" + tenantName)
	os.Remove(DOCKER_DIR + tenantName + ".env")

	// Update local file
	data, e := ioutil.ReadFile("tenants.json")
	if e != nil {
		panic(e.Error())
	}
	var listTenants []tenant
	json.Unmarshal(data, &listTenants)
	for i, ten := range listTenants {
		if ten.Name == tenantName {
			listTenants = append(listTenants[:i], listTenants[i+1:]...)
		}
	}
	data, _ = json.MarshalIndent(listTenants, "", "  ")
	_ = ioutil.WriteFile("tenants.json", data, 0644)
	c.IndentedJSON(http.StatusOK, "all good")
}
