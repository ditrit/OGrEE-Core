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

// OPENDCIM

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
