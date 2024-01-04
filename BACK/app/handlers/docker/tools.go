package docker

import (
	"back-admin/models"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"github.com/gin-gonic/gin"
)

var netboxDir string = "netbox-docker"

func CreateNetbox(c *gin.Context) {
	var newNetbox models.Netbox
	if err := c.BindJSON(&newNetbox); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}
	// Default values
	if newNetbox.Port == "" {
		newNetbox.Port = "8000"
	}

	netboxDir := "netbox-docker"
	if _, err := os.Stat(netboxDir); os.IsNotExist(err) {
		// Clone github repo
		println("Cloning Netbox git repo...")
		args := []string{"clone", "-b", "release", "https://github.com/netbox-community/netbox-docker.git"}
		cmd := exec.Command("git", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if _, err := cmd.Output(); err != nil {
			fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
			c.IndentedJSON(http.StatusBadRequest, stderr.String())
			return
		}
	}

	// Create compose override file
	file, _ := os.Create(netboxDir + "/docker-compose.override.yml")
	err := netboxtmplt.Execute(file, newNetbox)
	if err != nil {
		panic(err)
	}
	file.Close()
	// Create compose override copy
	file, _ = os.Create(netboxDir + "/docker-compose.override.yml." + newNetbox.Username)
	err = netboxtmplt.Execute(file, newNetbox)
	if err != nil {
		fmt.Println("Error creating compose copy: " + err.Error())
	}
	file.Close()

	println("Run docker (may take a long time...)")
	// Run docker
	args := []string{"compose", "up", "-d"}
	cmd := exec.Command("docker", args...)
	cmd.Dir = netboxDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusBadRequest, stderr.String())
		return
	}
	println("Finished with docker")
	c.IndentedJSON(http.StatusOK, "all good")
}

func RemoveNetbox(c *gin.Context) {
	netboxDir := "netbox-docker"
	println("Run docker (may take a long time...)")
	// Run docker
	args := []string{"compose", "down", "-v"}
	cmd := exec.Command("docker", args...)
	cmd.Dir = netboxDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusBadRequest, stderr.String())
		return
	}
	println("Finished with docker")
	c.IndentedJSON(http.StatusOK, "all good")
}

func AddNetboxDump(c *gin.Context) {
	// Load file
	formFile, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	// Save file
	err = c.SaveUploadedFile(formFile, netboxDir+"/dump.sql")
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	c.String(http.StatusOK, "")
}

func ImportNetboxDump(c *gin.Context) {
	args := []string{"compose", "stop"}
	cmd := exec.Command("docker", args...)
	cmd.Dir = netboxDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	args = []string{"start", "netbox-docker-postgres-1"}
	cmd = exec.Command("docker", args...)
	cmd.Dir = netboxDir
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	args = []string{"cp", "dump.sql", "netbox-docker-postgres-1:/tmp"}
	cmd = exec.Command("docker", args...)
	cmd.Dir = netboxDir
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	args = []string{"exec", "netbox-docker-postgres-1", "bash", "-c", "su - postgres; psql -U netbox postgres -c 'drop database netbox;'"}
	cmd = exec.Command("docker", args...)
	cmd.Dir = netboxDir
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	args = []string{"exec", "netbox-docker-postgres-1", "bash", "-c", "su - postgres; psql -U netbox postgres -c 'create database netbox;'"}
	cmd = exec.Command("docker", args...)
	cmd.Dir = netboxDir
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	args = []string{"exec", "netbox-docker-postgres-1", "bash", "-c", "su - postgres; psql -U netbox postgres -c 'grant all privileges on database netbox to netbox;'"}
	cmd = exec.Command("docker", args...)
	cmd.Dir = netboxDir
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	args = []string{"exec", "netbox-docker-postgres-1", "bash", "-c", "su - postgres; psql -U netbox netbox < /tmp/dump.sql"}
	cmd = exec.Command("docker", args...)
	cmd.Dir = netboxDir
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}

	args = []string{"compose", "start"}
	cmd = exec.Command("docker", args...)
	cmd.Dir = netboxDir
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
	}
}

func CreateOpenDcim(c *gin.Context) {
	var newDcim models.OpenDCIM
	if err := c.BindJSON(&newDcim); err != nil {
		c.IndentedJSON(http.StatusBadRequest, err.Error())
		return
	}

	// Create .env file
	composeDir := "tools-assets"
	file, _ := os.Create(composeDir + "/.env")
	err := opendcimtmplt.Execute(file, newDcim)
	if err != nil {
		panic(err)
	}
	file.Close()

	println("Run docker (may take a long time...)")
	// Run docker
	args := []string{"compose", "-f", "docker-compose-opendcim.yml", "-p", "opendcim", "up", "-d"}
	cmd := exec.Command("docker", args...)
	cmd.Dir = composeDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusBadRequest, stderr.String())
		return
	}
	println("Finished with docker")
	c.IndentedJSON(http.StatusOK, "all good")
}

func RemoveOpenDcim(c *gin.Context) {
	composeDir := "tools-assets"
	println("Run docker (may take a long time...)")
	// Run docker
	args := []string{"compose", "-f", "docker-compose-opendcim.yml", "-p", "opendcim", "down", "-v"}
	cmd := exec.Command("docker", args...)
	cmd.Dir = composeDir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if _, err := cmd.Output(); err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		c.IndentedJSON(http.StatusBadRequest, stderr.String())
		return
	}
	println("Finished with docker")
	c.IndentedJSON(http.StatusOK, "all good")
}
