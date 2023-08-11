package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	BffPort          string `json:"bffPort"`
	HasWeb           bool   `json:"hasWeb"`
	HasDoc           bool   `json:"hasDoc"`
	HasBff           bool   `json:"hasBff"`
	AssetsDir        string `json:"assetsDir"`
	ImageTag         string `json:"imageTag"`
	BffApiListFile   string
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

type backup struct {
	DBPassword string `json:"password" binding:"required"`
	IsDownload bool   `json:"shouldDownload"`
}

func getTenants(c *gin.Context) {
	response := make(map[string][]tenant)
	response["tenants"] = getTenantsFromJSON()
	c.IndentedJSON(http.StatusOK, response)
}

func getTenantsFromJSON() []tenant {
	if _, err := os.Stat("tenants.json"); errors.Is(err, os.ErrNotExist) {
		// tenants.json does not exist, create it
		var file, e = os.Create("tenants.json")
		if e != nil {
			panic(e.Error())
		} else {
			file.WriteString("[]")
			file.Sync()
			defer file.Close()
			return []tenant{}
		}
	}
	data, e := ioutil.ReadFile("tenants.json")
	if e != nil {
		panic(e.Error())
	}
	var listTenants []tenant
	json.Unmarshal(data, &listTenants)
	fmt.Println(listTenants)
	return listTenants
}

func getTenantDockerInfo(c *gin.Context) {
	name := c.Param("name")
	println(name)
	if response, err := getDockerInfo(name); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
	} else {
		c.IndentedJSON(http.StatusOK, response)
	}
}

func getDockerInfo(name string) ([]container, error) {
	println(name)
	cmd := exec.Command("docker", "ps", "--all", "--format", "\"{{json .}}\"")
	cmd.Dir = DOCKER_DIR
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if output, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return nil, errors.New(stderr.String())
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
			if name == "netbox" {
				if strings.Contains(dc.Name, "netbox-1") {
					response = append(response, dc)
				}
			} else if match, _ := regexp.MatchString("^"+name+"_", dc.Name); match {
				response = append(response, dc)
			}
		}
		if s.Err() != nil {
			// handle scan error
			fmt.Println(s.Err().Error())
			return nil, s.Err()
		}

		return response, nil
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
		if err := dockerCreateTenant(newTenant); err != "" {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}

		// Add to local json and respond
		listTenants = append(listTenants, newTenant)
		data, _ := json.MarshalIndent(listTenants, "", "  ")
		_ = ioutil.WriteFile("tenants.json", data, 0755)
		c.IndentedJSON(http.StatusOK, "all good")
	}

}

func dockerCreateTenant(newTenant tenant) string {
	tenantLower := strings.ToLower(newTenant.Name)
	appDeployDir := DOCKER_DIR + "app-deploy/" + tenantLower + "/"
	err := os.MkdirAll(appDeployDir, 0755)
	if err != nil && !strings.Contains(err.Error(), "already") {
		println(err.Error())
	}

	// Image tagging
	if newTenant.ImageTag == "" {
		newTenant.ImageTag = "main"
	}

	// Docker compose prepare
	args := []string{"compose", "-p", tenantLower}
	if newTenant.HasWeb {
		args = append(args, "--profile")
		args = append(args, "web")
		// Create flutter assets folder
		newTenant.AssetsDir = appDeployDir + "flutter"
		addAppAssets(newTenant)
	} else {
		// docker does not accept it empty, even if it wont be created
		newTenant.AssetsDir = DOCKER_DIR
	}

	if newTenant.HasDoc {
		args = append(args, "--profile")
		args = append(args, "doc")
	}

	if newTenant.HasBff {
		args = append(args, "--profile")
		args = append(args, "arango")
		if newTenant.BffPort == "" {
			// Set BFF and API ports
			newTenant.BffPort = newTenant.ApiPort
			port, _ := strconv.Atoi(newTenant.ApiPort)
			newTenant.ApiPort = strconv.Itoa(port + 1)
		}
		file, _ := os.Create(appDeployDir + tenantLower + "-bff-api-list.json")
		err := bfftmplt.Execute(file, newTenant)
		if err != nil {
			fmt.Println("Error creating bff api list file: " + err.Error())
			newTenant.BffApiListFile = "./bff_api_list.json"
		} else {
			newTenant.BffApiListFile = "./app-deploy/" + tenantLower + "/" + tenantLower + "-bff-api-list.json"
		}
		file.Close()
	}
	args = append(args, "up")
	args = append(args, "--build")
	args = append(args, "-d")

	// Create .env file
	file, _ := os.Create(DOCKER_DIR + ".env")
	err = tmplt.Execute(file, newTenant)
	if err != nil {
		panic(err)
	}
	file.Close()
	// Create tenantName.env as a copy
	file, _ = os.Create(appDeployDir + tenantLower + ".env")
	err = tmplt.Execute(file, newTenant)
	if err != nil {
		fmt.Println("Error creating .env copy: " + err.Error())
	}
	file.Close()

	println("Run docker (may take a long time...)")

	// Run docker
	cmd := exec.Command("docker", args...)
	cmd.Dir = DOCKER_DIR
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return stderr.String()
	}
	println("Finished with docker")
	return ""
}

func addAppAssets(newTenant tenant) {
	// Create flutter assets folder with .env
	err := os.MkdirAll(newTenant.AssetsDir, 0755)
	if err != nil && !strings.Contains(err.Error(), "already") {
		println(err.Error())
	}
	file, err := os.Create(newTenant.AssetsDir + "/.env")
	if err != nil {
		println(err.Error())
	}
	err = apptmplt.Execute(file, newTenant)
	if err != nil {
		println(err.Error())
	}
	file.Close()

	// Add default logo if none already present
	userLogo := newTenant.AssetsDir + "/logo.png"
	defaultLogo := "flutter-assets/logo.png"
	if _, err := os.Stat(userLogo); err == nil {
		println("Logo already exists")
	} else {
		println("Setting logo by default")
		source, err := os.Open(defaultLogo)
		if err != nil {
			println("Error opening default logo")
		}
		defer source.Close()
		destination, err := os.Create(userLogo)
		if err != nil {
			println("Error creating tenant logo file")
		}
		defer destination.Close()
		_, err = io.Copy(destination, source)
		if err != nil {
			println("Error creating tenant logo")
		}
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
	assetsDir := DOCKER_DIR + "app-deploy/" + tenantName + "/flutter"
	err = os.MkdirAll(assetsDir, 0755)
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
	for _, str := range []string{"_webapp", "_api", "_db", "_doc", "_bff", "_arango_api", "_arango_db"} {
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
	os.RemoveAll(DOCKER_DIR + "app-deploy/" + tenantName + "/flutter")
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
	_ = ioutil.WriteFile("tenants.json", data, 0755)
	c.IndentedJSON(http.StatusOK, "all good")
}

func updateTenant(c *gin.Context) {
	tenantName := strings.ToLower(c.Param("name"))
	println(tenantName)

	// Call BindJSON to bind the received JSON
	var newTenant tenant
	if err := c.BindJSON(&newTenant); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	listTenants := getTenantsFromJSON()
	for i, tenant := range listTenants {
		if strings.ToLower(tenant.Name) == tenantName {
			// Docker compose stop
			println("Docker stop current tenant")
			args := []string{"compose", "-p", tenantName, "stop"}
			cmd := exec.Command("docker", args...)
			cmd.Dir = DOCKER_DIR
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			if _, err := cmd.Output(); err != nil {
				fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
				c.IndentedJSON(http.StatusInternalServerError, stderr.String())
				return
			}
			println("Finished with docker")

			// Docker compose up
			if err := dockerCreateTenant(newTenant); err != "" {
				c.IndentedJSON(http.StatusInternalServerError, err)
				// Try to recreate previous config
				err = dockerCreateTenant(tenant)
				if err != "" {
					println("Error recovering:" + err)
				}
				return
			}

			listTenants[i] = newTenant
			println(listTenants)
			data, _ := json.MarshalIndent(listTenants, "", "  ")
			_ = ioutil.WriteFile("tenants.json", data, 0755)
			break
		}
	}

	c.String(http.StatusOK, "")
}

func backupTenantDB(c *gin.Context) {
	tenantName := strings.ToLower(c.Param("name"))
	t := time.Now()

	// Call BindJSON to bind the received JSON
	var backupInfo backup
	if err := c.BindJSON(&backupInfo); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	println("Docker backup current tenant")
	args := []string{"exec", tenantName + "_db", "sh", "-c",
		"exec mongodump --username ogree" + tenantName + "Admin --password " + backupInfo.DBPassword + " -d ogree" + tenantName + " --archive"}
	cmd := exec.Command("docker", args...)
	cmd.Dir = DOCKER_DIR
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	outfile, err := os.Create(tenantName + "_db_" + t.Format("2006-01-02T150405") + ".archive")
	if err != nil {
		println(err.Error())
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	defer outfile.Close()
	cmd.Stdout = outfile
	if err := cmd.Run(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusInternalServerError, stderr.String())
		return
	}

	dir, _ := os.Getwd()
	println("Finished with docker")
	if backupInfo.IsDownload {
		c.File(outfile.Name())
	} else {
		c.String(http.StatusOK, "Backup file created as "+outfile.Name()+" at "+dir)
	}
}
